package bboltrefdriver

import (
	"cmp"
	"encoding/json"
	"fmt"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/values"
	bolt "go.etcd.io/bbolt"
)

type _refDriver struct {
	db     *bolt.DB
	format doc.Format
}

func (d *_refDriver) Private(a any) error {
	switch t := a.(type) {
	case string:
		switch t {
		case "print":
			return d.print()
		}
	}
	return nil
}

func (d *_refDriver) Open(dataSourceName string) (doc.Driver, error) {
	eb := &errors.FirstBlock{}
	db, err := bolt.Open(dataSourceName, 0600, nil)
	eb.AddError(err)
	if eb.Err != nil {
		return nil, eb.Err
	}

	f := doc.FormatWithDefaults(_refNewFormat())
	return &_refDriver{db: db, format: f}, nil
}

func (d *_refDriver) Close() error {
	db := d.db
	d.db = nil
	if db != nil {
		db.Close()
	}
	return nil
}

func (d *_refDriver) Format() doc.Format {
	return d.format
}

func (d *_refDriver) Set(req doc.SetRequestAny, a doc.Allocator) (*doc.Optional, error) {
	data, err := d.prepareSet(req, a)
	if err != nil {
		return nil, err
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		b, lastErr := tx.CreateBucketIfNotExists([]byte(data.p.rootBucket))
		for _, node := range data.p.nodes {
			if lastErr != nil {
				return lastErr
			}
			if b == nil {
				return fmt.Errorf("Missing bucket")
			}
			if node.pt == bucketType {
				if node.value == "" {
					return fmt.Errorf("No value for %v", node.domainName)
				}
				b, lastErr = b.CreateBucketIfNotExists([]byte(node.value))
			} else if node.pt == keyType && node.autoinc {
				id, err := b.NextSequence()
				if err != nil {
					return err
				}
				return b.Put(_refItob(id), data.value)
			}
		}
		if lastErr != nil {
			return lastErr
		}
		if b == nil {
			return fmt.Errorf("Missing bucket")
		}
		key, err := data.p.makeKey()
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), data.value)
		return err
		//		return b.Put([]byte(key), data.value)
	})

	return nil, err
}

type setData struct {
	meta  *_refMetadata
	p     *path
	value []byte
}

func (d *_refDriver) prepareSet(req doc.SetRequestAny, a doc.Allocator) (setData, error) {
	tn := a.TypeName()
	meta, ok := _refMetadatas[tn]
	ps := setData{meta: meta}
	if !ok {
		return ps, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	ps.p = newPath(meta.rootBucket, meta.buckets)
	values.Get(req.ItemAny(), ps.p)

	// Marshal the data.
	dbitem, err := meta.toDb(req.ItemAny())
	if err != nil {
		return ps, err
	}
	dat, err := json.Marshal(dbitem)
	if err != nil {
		return ps, err
	}

	ps.value = dat
	return ps, nil
}

func (d *_refDriver) Get(req doc.GetRequest, a doc.Allocator) (*doc.Optional, error) {
	get, err := d.prepareGet(req, a)
	if err != nil {
		return nil, err
	}
	err = d.db.View(func(tx *bolt.Tx) error {
		it, err := newGetIterator(get.meta, tx, get.meta.rootBucket, get.keyValuesS, a)
		if err != nil {
			return err
		}
		item := it.next()
		for item != nil {
			item = it.next()
		}
		return it.err
	})
	return nil, err
}

type getData struct {
	meta *_refMetadata

	// These slices are all the same size, which will be
	// the metadata keys length. Store the various pieces of info needed.
	// keyValuesA and S are the same data, just composed for where they're
	// being used.
	domainNames []string
	keyValuesA  []any
	keyValuesS  []string
}

func (g *getData) BinaryConjunction(keyword string) error {
	if keyword != doc.AndKeyword {
		return fmt.Errorf("Unsupported binary: %v", keyword)
	}
	return nil
}

func (g *getData) BinaryAssignment(lhs, rhs string) error {
	for i, bucket := range g.meta.buckets {
		if bucket.boltName == lhs {
			g.keyValuesA[i] = rhs
			g.keyValuesS[i] = rhs
		}
	}
	return nil
}

func (d *_refDriver) prepareGet(req doc.GetRequest, a doc.Allocator) (getData, error) {
	tn := a.TypeName()
	meta, ok := _refMetadatas[tn]
	get := getData{meta: meta}
	get.domainNames = meta.DomainKeys()
	get.keyValuesA = make([]any, len(get.domainNames))
	get.keyValuesS = make([]string, len(get.domainNames))
	if !ok {
		return get, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	err := extractExpr(req.Condition, &get)
	return get, err
}

func (d *_refDriver) Delete(req doc.DeleteRequestAny, a doc.Allocator) (*doc.Optional, error) {
	del, err := d.prepareDelete(req, a)
	if err != nil {
		return nil, err
	}
	finalKeyIdx := len(del.keyValues) - 1
	err = d.db.Update(func(tx *bolt.Tx) error {
		compositeKey := ""
		b := tx.Bucket([]byte(del.meta.rootBucket))
		if b == nil {
			return fmt.Errorf("no root %v", del.meta.rootBucket)
		}
		for i, _key := range del.keyValues {
			key, err := keyBytes(_key)
			if err != nil {
				return err
			}
			if i == finalKeyIdx {
				// TODO: Some keys will be a composite and some will use the leaf
				// value, figure out the rules and abstract.
				if compositeKey != "" {
					compositeKey += "/"
				}
				compositeKey += fmt.Sprintf("%v", _key)
				return b.Delete([]byte(compositeKey))
			} else {
				b = b.Bucket(key)
				if b == nil {
					return fmt.Errorf("no bucket for key", _key)
				}
				// TODO: Some keys will be a composite and some will use the leaf
				// value, figure out the rules and abstract.
				if compositeKey != "" {
					compositeKey += "/"
				}
				compositeKey += fmt.Sprintf("%v", _key)
			}
		}
		return nil
	})
	return nil, err
}

type deleteData struct {
	meta *_refMetadata
	// The associated values for each of the keys.
	keyValues []any
}

func (d *deleteData) Handle(name string, value any) (string, any) {
	for i, bucket := range d.meta.buckets {
		if bucket.domainName == name {
			d.keyValues[i] = value
		}
	}
	return name, value
}

func (d *_refDriver) prepareDelete(req doc.DeleteRequestAny, a doc.Allocator) (deleteData, error) {
	item := req.ItemAny()
	if item == nil {
		return deleteData{}, fmt.Errorf("missing item")
	}
	tn := a.TypeName()
	meta, ok := _refMetadatas[tn]
	del := deleteData{meta: meta}
	if !ok {
		return del, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	if len(meta.buckets) < 1 {
		return del, fmt.Errorf("no keys for %v", tn)
	}
	del.keyValues = make([]any, len(meta.buckets))
	values.Get(item, &del)
	for i, b := range meta.buckets {
		if del.keyValues[i] == nil {
			return del, fmt.Errorf("missing value for field %v", b.domainName)
		}
	}
	return del, nil
}

func (d *_refDriver) print() error {
	if d.db == nil {
		return fmt.Errorf("No database")
	}
	err := d.db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			fmt.Printf("key=%s\n", k)
			d.printBucket(1, tx, tx.Bucket(k))
		}
		return nil
	})
	return err
}

func (d *_refDriver) printBucket(tabs int, tx *bolt.Tx, b *bolt.Bucket) {
	if b == nil {
		return
	}
	c := b.Cursor()
	tabStr := ""
	for i := 0; i < tabs; i++ {
		tabStr += "\t"
	}
	for k, v := c.First(); k != nil; k, v = c.Next() {
		fmt.Printf(tabStr)
		if v == nil {
			fmt.Printf("key=%s\n", k)
			d.printBucket(tabs+1, tx, b.Bucket(k))
		} else {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
	}
}

// ---------------------------------------------------------
// KEYS

func keyBytes(a any) ([]byte, error) {
	switch t := a.(type) {
	case string:
		return []byte(t), nil
	}
	return nil, fmt.Errorf("no key conversion for %v %T", a, a)
}

func extractExpr(expr doc.Expr, dst any) error {
	if expr != nil {
		expr, err := expr.Compile()
		if err != nil {
			return err
		}
		err = expr.Extract(dst)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------
// ITERATOR

func newGetIterator(meta *_refMetadata,
	tx *bolt.Tx,
	rootBucket string,
	bucketNames []string,
	a doc.Allocator) (*getIterator, error) {

	// Get root bucket
	b := tx.Bucket([]byte(rootBucket))
	if b == nil {
		return nil, fmt.Errorf("No root bucket %v", rootBucket)
	}

	if len(bucketNames) == 1 {
		return newPrefilledIterator(meta, tx, rootBucket, b, bucketNames, a)
	}

	// Follow the existing buckets as far as we can before we have to iterate.
	steps := []getIterBucket{}
	stepIdx := -1
	for i, name := range bucketNames {
		if name == "" {
			// wildcard, so break. iteration must be done below this.
			break
		}
		b = b.Bucket([]byte(name))
		if b == nil {
			return nil, fmt.Errorf("No bucket %v", name)
		}
		stepIdx = i
		gib := getIterBucket{name: name, b: b, rooted: true}
		steps = append(steps, gib)
	}

	// Initialize the cursor on the last known bucket
	if stepIdx < 0 {
		return nil, fmt.Errorf("No buckets")
	}
	gib := steps[stepIdx]
	steps[stepIdx] = gib

	req := values.SetRequest{FieldNames: meta.DomainKeys(),
		NewValues: make([]any, len(meta.DomainKeys()))}
	return &getIterator{meta: meta,
		a:           a,
		tx:          tx,
		bucketNames: bucketNames,
		steps:       steps,
		stepIdx:     stepIdx,
		badIdx:      stepIdx - 1,
		req:         req}, nil
}

// special case when the item is a key in the root bucket.
func newPrefilledIterator(meta *_refMetadata,
	tx *bolt.Tx,
	rootBucket string,
	b *bolt.Bucket,
	bucketNames []string,
	a doc.Allocator) (*getIterator, error) {
	fmt.Println("NEW PREFILLED")
	gi := &getIterator{meta: meta,
		a:           a,
		tx:          tx,
		bucketNames: bucketNames,
		stepIdx:     0,
		badIdx:      -1,
		oneShot:     true,
		print:       false,
	}
	gib := getIterBucket{name: rootBucket, b: b}
	gi.steps = append(gi.steps, gib)
	return gi, nil
}

type getIterator struct {
	meta        *_refMetadata
	a           doc.Allocator
	tx          *bolt.Tx
	bucketNames []string
	steps       []getIterBucket
	stepIdx     int
	badIdx      int
	req         values.SetRequest
	err         error
	oneShot     bool
	print       bool
}

func (g *getIterator) next() any {
	k, v, err := g.boltNext()
	for k == nil && err == nil {
		k, v, err = g.boltNext()
	}
	if err != nil {
		return nil
	}
	item, err := g.meta.fromDb(g.a.New(), v)
	//	fmt.Println("key", string(k), "value", string(v))
	g.err = cmp.Or(g.err, err)
	// How to do keys
	if len(g.meta.buckets) != len(g.steps) {
		g.err = cmp.Or(g.err, fmt.Errorf("Can't read keys for %v: step len %v should be %v", string(k), len(g.steps), len(g.meta.buckets)))
		return nil
	}
	if !g.oneShot {
		for i := 0; i < len(g.meta.buckets); i++ {
			g.req.NewValues[i] = g.steps[i].name
		}
	}
	values.Set(g.req, item)
	return item
}

func (g *getIterator) boltNext() ([]byte, []byte, error) {
	if g.print {
		fmt.Println("boltNext step", g.stepIdx, "bad", g.badIdx)
	}
	if g.stepIdx <= g.badIdx {
		return nil, nil, errFinished
	}

	if g.print {
		fmt.Println("boltNext idx", g.stepIdx)
	}
	gbi := g.steps[g.stepIdx]

	// Handle first get
	if gbi.c == nil {
		if g.print {
			fmt.Println("handle first")
		}
		gbi.c = gbi.b.Cursor()
		k, v := gbi.c.First()
		if g.print {
			fmt.Println("handle first", k, v)
		}
		if g.oneShot {
			g.stepIdx = g.badIdx
			return k, v, nil
		}
		if k == nil {
			g.stepIdx--
			return nil, nil, nil
		}
		g.steps[g.stepIdx] = gbi
		// Handle buckets
		if v == nil {
			g.handleBucket(k, gbi.b)
			return nil, nil, nil
		}
		return k, v, nil
	}

	// Handle next get
	k, v := gbi.c.Next()
	if k == nil {
		g.stepIdx--
		return nil, nil, nil
	}
	// Handle buckets
	if v == nil {
		g.handleBucket(k, gbi.b)
		return nil, nil, nil
	}
	return k, v, nil
}

func (g *getIterator) handleBucket(k []byte, parent *bolt.Bucket) {
	// TODO: This will iterate all bucket keys even if I have a specific
	// one to select.
	//	fmt.Println("bucketName at", g.stepIdx+1, "current key", string(k))
	sk := string(k)
	match := g.bucketNameAt(g.stepIdx + 1)
	//	fmt.Println("bucketName match", string(match))
	if match == "" || match == sk {
		b := parent.Bucket(k)
		if b != nil {
			next := getIterBucket{name: sk, b: b}
			g.steps = append(g.steps, next)
			g.stepIdx++
		}
	}
}

func (g *getIterator) bucketNameAt(i int) string {
	if i >= 0 && i < len(g.bucketNames) {
		return g.bucketNames[i]
	}
	return ""
}

// getIterBucket is a single step in the key/bucket iteration.
// It contains everything about the step: The bucket name, actual
// bucket, iteration cursor, etc.
type getIterBucket struct {
	name   string
	b      *bolt.Bucket
	c      *bolt.Cursor
	rooted bool
	// finished is set to true once this bucket is done iterating
	// it's cursor (and also set to true for buckets that don't
	// need iteration, such as those found in the initial path traversal.
	finished bool
}

// ---------------------------------------------------------
// PATH

type pathType uint8

const (
	bucketType pathType = iota
	keyType
)

func newPath(rootBucket string, buckets []_refKeyMetadata) *path {
	nodes := make([]pathNode, 0, len(buckets))
	for _, b := range buckets {
		pn := pathNode{pt: bucketType, domainName: b.domainName, ft: b.ft, leaf: b.leaf, autoinc: b.autoInc}
		if b.leaf {
			pn.pt = keyType
		}
		nodes = append(nodes, pn)
	}
	return &path{rootBucket: rootBucket, nodes: nodes}
}

type pathNode struct {
	pt pathType
	// domainName for the field
	domainName string
	// value -- name of the bucket or key
	value   string
	ft      fieldType
	leaf    bool
	autoinc bool
}

type path struct {
	rootBucket string
	nodes      []pathNode
}

func (p *path) Handle(name string, value any) (string, any) {
	for i, b := range p.nodes {
		if b.domainName == name {
			if s, ok := value.(string); ok {
				b.value = s
				p.nodes[i] = b
			}
			break
		}
	}
	return name, value
}

// makeKey returns a value to be used as the key in the database.
func (p *path) makeKey() (string, error) {
	// Validate
	if len(p.nodes) < 1 {
		return "", fmt.Errorf("No path nodes")
	}

	// If we end in a leaf, use that single value as the key.
	tail := p.nodes[len(p.nodes)-1]
	if tail.leaf {
		if tail.value == "" {
			return "", fmt.Errorf("No value for leaf key")
		}
		return tail.value, nil
	}

	// If we don't have a leaf then the key is a composite of
	// all my buckets.
	key := p.rootBucket
	for _, n := range p.nodes {
		if n.value == "" {
			return "", fmt.Errorf("Missing value for %v", n.domainName)
		}
		key += "/"
		key += n.value
	}
	return key, nil
}

var errFinished = fmt.Errorf("finished")
