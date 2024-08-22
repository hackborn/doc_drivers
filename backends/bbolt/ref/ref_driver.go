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
				if node.value == nil {
					return fmt.Errorf("No value for %v", node.domainName)
				}
				b, lastErr = b.CreateBucketIfNotExists(node.value)
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
		err = b.Put(key, data.value)
		return err
		//		return b.Put(key, data.value)
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
		//		it, err := newGetIterator(get.meta, tx, get.meta.rootBucket, get.keyValuesS, a)
		it, err := newGetIterator(get.meta, tx, get.p, a)
		if err != nil {
			return err
		}
		item := it.Next()
		for item != nil {
			item = it.Next()
		}
		return it.Err()
	})
	return nil, err
}

type getData struct {
	meta *_refMetadata
	p    *path

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
	if !ok {
		return get, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	get.p = newPath(meta.rootBucket, meta.buckets)
	err := extractExpr(req.Condition, get.p)

	get.domainNames = meta.DomainKeys()
	get.keyValuesA = make([]any, len(get.domainNames))
	get.keyValuesS = make([]string, len(get.domainNames))
	if !ok {
		return get, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	err = extractExpr(req.Condition, &get)
	return get, err
}

func (d *_refDriver) Delete(req doc.DeleteRequestAny, a doc.Allocator) (*doc.Optional, error) {
	del, err := d.prepareDelete(req, a)
	if err != nil {
		return nil, err
	}
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(del.meta.rootBucket))
		if b == nil {
			return fmt.Errorf("missing root bucket %v", del.meta.rootBucket)
		}
		for i, node := range del.p.nodes {
			if node.leaf {
				return b.Delete([]byte(del.key))
			}
			b = b.Bucket(node.value)
			if b == nil {
				return fmt.Errorf("missing bucket")
			}

			if i >= len(del.p.nodes)-1 {
				return b.Delete([]byte(del.key))
			}
		}
		return fmt.Errorf("Fell out of loop")
	})
	return nil, err
}

type deleteData struct {
	meta *_refMetadata
	p    *path
	key  boltKey
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
	del.p = newPath(meta.rootBucket, meta.buckets)
	values.Get(item, del.p)
	k, err := del.p.makeKey()
	if err != nil {
		return del, err
	}
	del.key = k

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

type getIterator interface {
	Next() any
	Err() error
}

func newGetIterator(meta *_refMetadata,
	tx *bolt.Tx,
	p *path,
	a doc.Allocator) (getIterator, error) {
	// Get root bucket
	b := tx.Bucket([]byte(meta.rootBucket))
	if b == nil {
		return nil, fmt.Errorf("No root bucket %v", meta.rootBucket)
	}

	//	fmt.Println("GET PATH", *p)
	steps := make([]wildcardIteratorStep, 0, len(p.nodes))
	steps = append(steps, wildcardIteratorStep{})
	req := values.SetRequest{FieldNames: meta.DomainKeys(),
		NewValues: make([]any, len(meta.DomainKeys()))}
	return &wildcardIterator{meta: meta,
		a:     a,
		tx:    tx,
		b:     b,
		p:     p,
		steps: steps,
		req:   req}, nil
}

// wildcardIterator allows missing key values, in which
// case buckets are iterated.
type wildcardIterator struct {
	meta *_refMetadata
	a    doc.Allocator
	tx   *bolt.Tx
	// root bucket
	b     *bolt.Bucket
	p     *path
	steps []wildcardIteratorStep
	req   values.SetRequest
	err   error
}

func (w *wildcardIterator) Next() any {
	k, v, err := w.step()
	for (len(k) < 1 || len(v) < 1) && err == nil {
		k, v, err = w.step()
	}
	if err != nil {
		if err != errFinished {
			w.err = cmp.Or(w.err, err)
		}
		return nil
	}
	return w.domainItem(k, v)
}

// domainItem converst the key and value into a domain item.
func (w *wildcardIterator) domainItem(k, v []byte) any {
	item, err := w.meta.fromDb(w.a.New(), v)
	//	fmt.Println("key", string(k), "value", string(v))
	w.err = cmp.Or(w.err, err)
	// Set the keys. The values should be in the steps.
	for i, node := range w.p.nodes {
		if i < len(w.req.NewValues) {
			var value any
			if i < len(w.steps) {
				if node.ft == stringType {
					value = string(w.steps[i].key)
				}
			}
			w.req.NewValues[i] = value
		}
	}
	values.Set(w.req, item)
	return item
}

func (g *wildcardIterator) Err() error {
	return g.err
}

func (g *wildcardIterator) step() ([]byte, []byte, error) {
	/*
		tabs := ""
		for range g.steps {
			tabs += "\t"
		}
		fmt.Println(tabs, "step a")
	*/
	if len(g.steps) < 1 {
		return nil, nil, errFinished
	}

	idx := len(g.steps) - 1
	step := &g.steps[idx]
	//	fmt.Println(tabs, "step b type", step.stepType, "idx", idx)
	switch step.stepType {
	case initStep:
		// always need a current bucket to operate on
		currentBucket := g.bucket()
		//		fmt.Println(tabs, "step c bucket", currentBucket)
		if currentBucket == nil {
			return nil, nil, fmt.Errorf("init step missing bucket")
		}

		// If we're a) past the path or b) the path is only 1 level
		// deep then this has to be a single composite key into the current bucket.
		if len(g.steps) > len(g.p.nodes) || len(g.p.nodes) == 1 {
			//			fmt.Println(tabs, "getting an item and popping")
			if idx < len(g.p.nodes) {
				step.key = g.p.nodes[idx].value
			}
			k := g.key()
			g.popStep()
			return g.getStep(currentBucket, k)
		}
		if step.finished {
			g.popStep()
			return nil, nil, nil
		}
		// Handle keys (leaf)
		node := g.p.nodes[idx]
		if node.leaf {
			g.popStep()
			return g.getStep(currentBucket, g.key())
		}
		// Handle buckets - direct
		//		fmt.Println(tabs, "step d key", node.value)
		if node.value != nil {
			//			fmt.Println(tabs, "step e direct key", node.value)
			step.key = node.value
			step.finished = true
			step.b = currentBucket.Bucket(step.key)
			if step.b == nil {
				//				fmt.Println(tabs, "get bucket for key", string(step.key))
				return nil, nil, fmt.Errorf("direct bucket missing bucket")
			}
			g.steps = append(g.steps, wildcardIteratorStep{})
			return nil, nil, nil
		}
		// Handle buckets - wildcard
		step.stepType = cursorStep
		step.c = currentBucket.Cursor()
		k, v := step.c.First()
		//		fmt.Println(tabs, "cursor first", string(k))
		return g.cursorStep(k, v, step)
	case cursorStep:
		k, v := step.c.Next()
		//		fmt.Println(tabs, "cursor step", string(k))
		return g.cursorStep(k, v, step)
	}
	// Shouldn't reach here
	g.err = cmp.Or(g.err, fmt.Errorf("fell through step loop"))
	return nil, nil, nil
}

func (g *wildcardIterator) getStep(b *bolt.Bucket, k boltKey) ([]byte, []byte, error) {
	if b == nil {
		return nil, nil, fmt.Errorf("getStep missing bucket")
	}
	if k == nil {
		return nil, nil, fmt.Errorf("getStep missing key")
	}
	return k, b.Get(k), nil
}

func (g *wildcardIterator) cursorStep(k, v []byte, step *wildcardIteratorStep) ([]byte, []byte, error) {
	if k == nil {
		g.popStep()
	} else if v == nil {
		step.key = k
		step.b = nil
		b := g.bucket()
		if b == nil {
			return nil, nil, fmt.Errorf("cursor step missing parent bucket")
		}
		step.b = b.Bucket(k)
		if step.b == nil {
			return nil, nil, fmt.Errorf("cursor step can't find bucket")
		}
		g.steps = append(g.steps, wildcardIteratorStep{})
	} else {
		// I think I ignore these, I'm in a bucket and need to
		// go a step deeper to find values.
	}
	return nil, nil, nil
}

func (g *wildcardIterator) popStep() {
	if len(g.steps) > 0 {
		g.steps = g.steps[:len(g.steps)-1]
	}
}

// bucket answers the currently active bucket. This will
// either be the bucket at the end of the steps or the
// root bucket. nil for an error.
func (g *wildcardIterator) bucket() *bolt.Bucket {
	for i := len(g.steps); i > 0; i-- {
		step := &g.steps[i-1]
		if step.b != nil {
			return step.b
		}
	}
	return g.b
}

// key answers the current key based on the path rules.
// nil for error.
func (g *wildcardIterator) key() boltKey {
	var k boltKey
	for i, step := range g.steps {
		if i >= len(g.p.nodes) {
			return k
		}
		node := g.p.nodes[i]
		if step.key == nil {
			g.err = cmp.Or(g.err, fmt.Errorf("missing key for %v", node.domainName))
			return nil
		}
		if node.leaf {
			return step.key
		}
		if k != nil {
			k = append(k, _refKeySep...)
		}
		k = append(k, step.key...)
	}
	return k
}

type wildcardStepType int

const (
	initStep wildcardStepType = iota
	cursorStep
)

type wildcardIteratorStep struct {
	stepType wildcardStepType
	//	nodeIdx  int
	key boltKey
	//	name     string
	b *bolt.Bucket
	c *bolt.Cursor
	//	rooted bool
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
		pn := pathNode{pt: bucketType,
			domainName: b.domainName,
			boltName:   b.boltName,
			ft:         b.ft,
			leaf:       b.leaf,
			autoinc:    b.autoInc,
		}
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
	boltName   string
	// value -- name of the bucket or key
	value boltKey
	//	value   string
	ft      fieldType
	leaf    bool
	autoinc bool
}

type path struct {
	rootBucket string
	nodes      []pathNode
}

// Handle is used by the reflection system to extact my
// node values from a domain object.
func (p *path) Handle(name string, value any) (string, any) {
	for i, node := range p.nodes {
		if node.domainName == name {
			if key, ok := _refToBoltKey(value, node.ft); ok {
				node.value = key
				p.nodes[i] = node
			}
			break
		}
	}
	return name, value
}

// BinaryConjunction is used by the expression parsing to
// extract my node values from an expression.
func (p *path) BinaryConjunction(keyword string) error {
	if keyword != doc.AndKeyword {
		return fmt.Errorf("Unsupported binary: %v", keyword)
	}
	return nil
}

// BinaryAssignment is used by the expression parsing to
// extract my node values from an expression.
func (p *path) BinaryAssignment(lhs string, rhs any) error {
	for i, node := range p.nodes {
		if node.boltName == lhs {
			if value, ok := _refToBoltKey(rhs, node.ft); ok {
				node.value = value
				p.nodes[i] = node
			}
			//			fmt.Println("Extract", lhs, rhs, "value", node.value)
		}
	}
	return nil
}

// makeKey returns a value to be used as the key in the database.
func (p *path) makeKey() (boltKey, error) {
	// Validate
	if len(p.nodes) < 1 {
		return nil, fmt.Errorf("No path nodes")
	}

	// The key is a composite of all my buckets.
	var key boltKey
	for _, n := range p.nodes {
		if n.value == nil {
			return nil, fmt.Errorf("Missing value for %v", n.domainName)
		}
		// If this is a leaf then it's the only key we use
		if n.leaf {
			return n.value, nil
		}

		if key != nil {
			key = append(key, _refKeySep...)
		}
		key = append(key, n.value...)
	}
	return key, nil
}

var errFinished = fmt.Errorf("finished")
