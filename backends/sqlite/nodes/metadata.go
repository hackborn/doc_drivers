package nodes

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"text/scanner"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofslices "github.com/hackborn/onefunc/slices"
)

// metadata is parallel to pipeline.StructData, except with parsed tags.
type metadata struct {
	// The table name
	Name   string
	Fields []structField
	Keys   map[string][]structKey
}

func (d metadata) TagNames() []string {
	return ofslices.ArrayFrom(d.Fields, func(f structField) string {
		return f.Tag
	})
}

func (d metadata) FieldNames() []string {
	return ofslices.ArrayFrom(d.Fields, func(f structField) string {
		return f.Field
	})
}

func (d metadata) KeyTagNames(key string) []string {
	return ofslices.ArrayFrom(d.Keys[key], func(key structKey) string {
		return key.Tag
	})
}

func (d metadata) KeyFieldNames(key string) []string {
	return ofslices.ArrayFrom(d.Keys[key], func(key structKey) string {
		return key.Field
	})
}

// PrimaryKey answers the "primary" key in the key map. This is
// just defined as the first key in alphabetic order.
func (d metadata) PrimaryKey() (string, bool) {
	if len(d.Keys) < 1 {
		return "", false
	}
	first := true
	key := ""
	for k, _ := range d.Keys {
		if first {
			key = k
			first = false
		} else {
			if strings.Compare(k, key) < 0 {
				key = k
			}
		}
	}
	return key, true
}

type structField struct {
	Tag   string
	Field string
	Type  string
}

type structKey struct {
	Tag   string
	Field string
}

type parsedKey struct {
	name     string
	position int

	tagName   string
	fieldName string
}

// makeMetadata answers the results of parsing the struct
// data, including the tags, into a parallel structure.
func makeMetadata(pin *pipeline.StructData, tablePrefix string) (metadata, error) {
	eb := oferrors.FirstBlock{}
	md := metadata{Name: pin.Name}
	md.Keys = make(map[string][]structKey)
	keys := make(map[string][]*parsedKey)
	for _, f := range pin.Fields {
		sf, pk := makeStructField(f, &eb)
		// Skip indicator
		if sf.Tag == "-" {
			continue
		}
		// Default field name indicator.
		if sf.Tag == "" {
			sf.Tag = sf.Field
		}
		md.Fields = append(md.Fields, sf)
		if pk != nil {
			pk.tagName = sf.Tag
			pk.fieldName = sf.Field
			if found, ok := keys[pk.name]; ok {
				found = append(found, pk)
				keys[pk.name] = found
			} else {
				keys[pk.name] = []*parsedKey{pk}
			}
		}
	}
	// Compile the keys
	for k, v := range keys {
		slices.SortFunc(v, func(a, b *parsedKey) int {
			if a.position < b.position {
				return -1
			} else if a.position > b.position {
				return 1
			} else {
				return 0
			}
		})
		value := make([]structKey, 0, len(v))
		for _, vv := range v {
			value = append(value, structKey{Tag: vv.tagName, Field: vv.fieldName})
		}
		md.Keys[k] = value
	}
	makeTableMetadata(pin, &md, &eb)
	md.Name = tablePrefix + md.Name
	return md, eb.Err
}

func makeTableMetadata(pin *pipeline.StructData, md *metadata, eb oferrors.Block) {
	for _, f := range pin.UnexportedFields {
		// The tag was filtered for my "doc" keyword, so any non-empty
		// tag will be a table specification
		if f.Tag == "" {
			continue
		}
		sf, _ := makeStructField(f, eb)
		if sf.Tag != "" {
			md.Name = sf.Tag
		}
	}
}

// makeStructField parses a single field in the struct, including
// metadata.
func makeStructField(f pipeline.StructField, eb oferrors.Block) (structField, *parsedKey) {
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(f.Tag))
	lexer.Whitespace = 0
	lexer.Mode = scanner.ScanChars | scanner.ScanFloats | scanner.ScanIdents | scanner.ScanInts | scanner.ScanRawStrings | scanner.ScanStrings
	lexer.Error = func(s *scanner.Scanner, msg string) {
		eb.AddError(fmt.Errorf("key scan error: %v", msg))
	}
	h := startKeyScanHandler()
	kt := keyToken{}
	for tok := lexer.Scan(); tok != scanner.EOF; tok = lexer.Scan() {
		// fmt.Println("TOK", tok, "name", scanner.TokenString(tok), "text", lexer.TokenText())
		kt.text = lexer.TokenText()
		switch tok {
		case scanner.Float:
			kt.tokenType = floatToken
		case scanner.Int:
			kt.tokenType = intToken
		case scanner.String:
			kt.tokenType = stringToken
		case scanner.Ident:
			kt.tokenType = stringToken
		case ' ', '\r', '\t', '\n':
			kt.tokenType = whitespaceToken
			kt.text = ""
		default:
			kt.tokenType = stringToken
		}
		h.HandleToken(kt, eb)
	}
	sd := structField{Tag: h.nameHandler.name, Field: f.Name, Type: f.Type}
	var key *parsedKey = nil
	if h.keyHandler.exists {
		key = &parsedKey{name: h.keyHandler.keyName, position: h.keyHandler.keyPosition}
	} else if h.identHandler.name == "key" {
		// This happens when we have a dangling key definition, i.e. nothing
		// after the "key" keyword. For example, an unnamed key.
		key = &parsedKey{}
	}
	return sd, key
}

type keyTokenType int

const (
	stringToken keyTokenType = iota
	floatToken
	intToken
	whitespaceToken
)

type keyToken struct {
	tokenType keyTokenType
	text      string
}

type fieldScanHandler interface {
	HandleToken(t keyToken, eb oferrors.Block)
}

type lifecycleHandler interface {
	onPushed()
}

type baseKeyScanHandler struct {
	stack        []fieldScanHandler
	nameHandler  nameKeyScanHandler
	identHandler identKeyScanHandler
	keyHandler   keyFieldScanHandler
}

func startKeyScanHandler() *baseKeyScanHandler {
	base := &baseKeyScanHandler{}
	base.nameHandler.base = base
	base.identHandler.base = base
	base.keyHandler.base = base
	base.stack = append(base.stack, &base.nameHandler)
	return base
}

func (h *baseKeyScanHandler) HandleToken(t keyToken, eb oferrors.Block) {
	if len(h.stack) < 1 {
		eb.AddError(fmt.Errorf("key scan missing handler"))
	} else {
		h.stack[len(h.stack)-1].HandleToken(t, eb)
	}
}

func (h *baseKeyScanHandler) push(next fieldScanHandler) {
	h.stack = append(h.stack, next)
	if lh, ok := next.(lifecycleHandler); ok {
		lh.onPushed()
	}
}

func (h *baseKeyScanHandler) pop() {
	if len(h.stack) < 1 {
		// error
	} else {
		h.stack = h.stack[0:(len(h.stack) - 1)]
	}
}

type nameKeyScanHandler struct {
	base *baseKeyScanHandler
	name string
}

func (h *nameKeyScanHandler) HandleToken(t keyToken, eb oferrors.Block) {
	switch t.tokenType {
	case stringToken:
		h.HandleStringToken(t, eb)
	case intToken, floatToken:
		eb.AddError(fmt.Errorf("key scan name received illegal token %v", t.text))
	}
}

func (h *nameKeyScanHandler) HandleStringToken(t keyToken, eb oferrors.Block) {
	if t.text == "," {
		h.base.pop()
		h.base.push(&h.base.identHandler)
	} else if h.name == "" {
		h.name = t.text
	} else {
		eb.AddError(fmt.Errorf("key scan has name %v but wants %v", h.name, t.text))
	}
}

type identKeyScanHandler struct {
	base *baseKeyScanHandler
	name string
}

func (h *identKeyScanHandler) HandleToken(t keyToken, eb oferrors.Block) {
	switch t.tokenType {
	case stringToken:
		h.HandleStringToken(t, eb)
	case intToken, floatToken:
		eb.AddError(fmt.Errorf("key scan name received illegal token %v", t.text))
	}
}

func (h *identKeyScanHandler) HandleStringToken(t keyToken, eb oferrors.Block) {
	if t.text == "(" {
		if h.name == "key" {
			h.name = ""
			h.base.push(&h.base.keyHandler)
		}
	} else if h.name == "" {
		h.name = t.text
	} else {
		eb.AddError(fmt.Errorf("key scan has name %v but wants %v", h.name, t.text))
	}
}

type keyFieldScanHandler struct {
	base        *baseKeyScanHandler
	exists      bool
	keyName     string
	keyPosition int
	index       int
}

func (h *keyFieldScanHandler) HandleToken(t keyToken, eb oferrors.Block) {
	switch t.tokenType {
	case stringToken:
		h.HandleStringToken(t, eb)
	case intToken:
		h.HandleIntToken(t, eb)
	case whitespaceToken:
	default:
		eb.AddError(fmt.Errorf("keyFieldScanHandler received illegal token %v", t.text))
	}
}

func (h *keyFieldScanHandler) HandleStringToken(t keyToken, eb oferrors.Block) {
	if t.text == "," {
		h.index++
	} else if t.text == ")" {
		h.base.pop()
	} else if h.index != 0 {
		eb.AddError(fmt.Errorf("keyFieldScanHandler received name \"%v\" at wrong index %v", t.text, h.index))
	} else if h.keyName != "" {
		eb.AddError(fmt.Errorf("keyFieldScanHandler received name %v but has name %v", t.text, h.keyName))
	} else {
		h.keyName = t.text
	}
}

func (h *keyFieldScanHandler) HandleIntToken(t keyToken, eb oferrors.Block) {
	if h.index != 1 {
		eb.AddError(fmt.Errorf("keyFieldScanHandler received int at wrong index %v", t.text))
	} else {
		i, err := strconv.Atoi(t.text)
		if err != nil {
			eb.AddError(fmt.Errorf("keyFieldScanHandler int converstion error %w", err))
		} else {
			h.keyPosition = i
		}
	}
}

func (h *keyFieldScanHandler) onPushed() {
	h.exists = true
	h.keyName = ""
	h.keyPosition = 0
	h.index = 0
}
