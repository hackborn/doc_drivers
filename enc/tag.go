package enc

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"

	oferrors "github.com/hackborn/onefunc/errors"
)

type Tag struct {
	Name     string
	Format   string
	HasKey   bool
	KeyGroup string
	KeyIndex int
	Flags    Flags
	// True if the autoinc tag is set. Indicates this value
	// should be automatically set on item creation.
	// Deprecated, use flags
	AutoInc bool
}

func (t Tag) Validate() error {
	if t.AutoInc == true && t.HasKey == false {
		return fmt.Errorf("Tag autoinc can only be set on keys")
	}
	return nil
}

// ParseTag parses a tag expression.
func ParseTag(expr string) (Tag, error) {
	eb := &oferrors.FirstBlock{}
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(expr))
	//	lexer.Whitespace = 0
	lexer.Mode = scanner.ScanChars | scanner.ScanFloats | scanner.ScanIdents | scanner.ScanInts | scanner.ScanRawStrings | scanner.ScanStrings
	lexer.Error = func(s *scanner.Scanner, msg string) {
		eb.AddError(fmt.Errorf("key scan error: %v", msg))
	}
	state := &tagParserState{eb: eb}
	state.push(&tagParserKeywordHandler{})
	args := tagParserArgs{state: state}
	for tok := lexer.Scan(); tok != scanner.EOF; tok = lexer.Scan() {
		//		fmt.Println("TOK", tok, "name", scanner.TokenString(tok), "text", lexer.TokenText())
		args.token, args.text = tok, lexer.TokenText()
		state.handle(args)
	}
	for len(state.stack) > 0 {
		state.pop()
	}
	return state.tag, eb.Err
}

type tagParserArgs struct {
	state *tagParserState
	token rune
	text  string
}

// tagParserHandler defines a token handler.
type tagParserHandler interface {
	Start(*tagParserState)
	End(*tagParserState)
	Handle(tagParserArgs)
}

// tagParserState contains the parsing stack and other state.
type tagParserState struct {
	tag Tag
	eb  oferrors.Block

	stack []tagParserHandler
}

func (s *tagParserState) handle(args tagParserArgs) {
	if len(s.stack) < 1 {
		s.eb.AddError(fmt.Errorf("Empty stack on token \"%v\"", args.text))
	} else {
		s.stack[len(s.stack)-1].Handle(args)
	}
}

func (s *tagParserState) push(h tagParserHandler) {
	if h == nil {
		s.eb.AddError(fmt.Errorf("No handler for keyword"))
	} else {
		h.Start(s)
		s.stack = append(s.stack, h)
	}
}

func (s *tagParserState) pop() {
	if len(s.stack) < 1 {
		s.eb.AddError(fmt.Errorf("Popping empty stack"))
	} else {
		idx := len(s.stack) - 1
		s.stack[idx].End(s)
		s.stack = s.stack[0:idx]
	}
}

// tagParserKeywordHandler handles the top-level keywords.
type tagParserKeywordHandler struct {
}

func (h *tagParserKeywordHandler) Start(*tagParserState) {
}

func (h *tagParserKeywordHandler) End(*tagParserState) {
}

func (h *tagParserKeywordHandler) Handle(args tagParserArgs) {
	switch args.text {
	case "-":
		// skip
		args.state.tag.Name = args.text
	case ",":
		// This shouldn't be hit, it's always the wrapper
		args.state.pop()
	case "name":
		args.state.push(&tagParserHandlerWrapper{inner: &tagParserNameHandler{}})
	case "key":
		args.state.push(&tagParserHandlerWrapper{inner: &tagParserKeyHandler{}})
	case "format":
		args.state.push(&tagParserHandlerWrapper{inner: &tagParserFormatHandler{}})
	case "autoinc":
		args.state.push(&tagParserHandlerWrapper{inner: &tagParserAutoincHandler{}})
	default:
		args.state.eb.AddError(fmt.Errorf("Unknown token \"%v\"", args.text))
	}
}

// tagParserNameHandler handles the name.
type tagParserNameHandler struct {
}

func (h *tagParserNameHandler) Start(*tagParserState) {
}

func (h *tagParserNameHandler) End(*tagParserState) {
}

func (h *tagParserNameHandler) Handle(args tagParserArgs) {
	switch args.text {
	case ")":
		args.state.pop()
	default:
		if args.state.tag.Name == "" {
			args.state.tag.Name = args.text
		}
	}
}

// tagParserFormatHandler handles the format.
type tagParserFormatHandler struct {
}

func (h *tagParserFormatHandler) Start(*tagParserState) {
}

func (h *tagParserFormatHandler) End(*tagParserState) {
}

func (h *tagParserFormatHandler) Handle(args tagParserArgs) {
	switch args.text {
	case ")":
		args.state.pop()
	default:
		if args.state.tag.Format == "" {
			args.state.tag.Format = args.text
		}
	}
}

// tagParserKeyHandler handles the key.
type tagParserKeyHandler struct {
	idx int
}

func (h *tagParserKeyHandler) Start(s *tagParserState) {
	s.tag.HasKey = true
}

func (h *tagParserKeyHandler) End(*tagParserState) {
}

func (h *tagParserKeyHandler) Handle(args tagParserArgs) {
	switch args.text {
	case ")":
		args.state.pop()
	case ",":
		h.idx++
	default:
		switch h.idx {
		case 0:
			args.state.tag.KeyGroup = args.text
		case 1:
			i, err := strconv.Atoi(args.text)
			if err != nil {
				args.state.eb.AddError(fmt.Errorf("illegal key index \"%v\"", args.text))
			} else {
				args.state.tag.KeyIndex = i
			}
		default:
			args.state.eb.AddError(fmt.Errorf("Key index %v too high on token \"%v\"", h.idx, args.text))
		}
	}
}

// tagParserAutoincHandler handles the autoinc.
type tagParserAutoincHandler struct {
	flag Flags
}

func (h *tagParserAutoincHandler) Start(s *tagParserState) {
	h.flag = FlagAutoIncGlobal
	s.tag.AutoInc = true
}

func (h *tagParserAutoincHandler) End(s *tagParserState) {
	s.tag.Flags |= h.flag
}

func (h *tagParserAutoincHandler) Handle(args tagParserArgs) {
	t := strings.ToLower(args.text)
	switch t {
	case ")":
		args.state.pop()
	case "global":
		h.flag = FlagAutoIncGlobal
	case "local":
		h.flag = FlagAutoIncLocal
	}
}

// tagParserHandlerWrapper wraps a handler. This is installed
// from the main keyword, and the item I wrap will be installed from parens.
type tagParserHandlerWrapper struct {
	inner  tagParserHandler
	pushed bool
}

func (h *tagParserHandlerWrapper) Start(*tagParserState) {
}

func (h *tagParserHandlerWrapper) End(s *tagParserState) {
	if !h.pushed {
		h.inner.Start(s)
		h.inner.End(s)
	}
}

func (h *tagParserHandlerWrapper) Handle(args tagParserArgs) {
	switch args.text {
	case "(":
		h.pushed = true
		args.state.push(h.inner)
	case ",":
		args.state.pop()
	}
}
