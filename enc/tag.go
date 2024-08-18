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
}

// ParseTag parses a tag.
func ParseTag(tag string) (Tag, error) {
	eb := &oferrors.FirstBlock{}
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(tag))
	//	lexer.Whitespace = 0
	lexer.Mode = scanner.ScanChars | scanner.ScanFloats | scanner.ScanIdents | scanner.ScanInts | scanner.ScanRawStrings | scanner.ScanStrings
	lexer.Error = func(s *scanner.Scanner, msg string) {
		eb.AddError(fmt.Errorf("key scan error: %v", msg))
	}
	state := &tagParserState{eb: eb}
	state.push(&tagParserKeywordHandler{})
	args := tagParserArgs{state: state}
	for tok := lexer.Scan(); tok != scanner.EOF; tok = lexer.Scan() {
		// fmt.Println("TOK", tok, "name", scanner.TokenString(tok), "text", lexer.TokenText())
		args.token, args.text = tok, lexer.TokenText()
		state.handle(args)
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
		s.stack = append(s.stack, h)
	}
}

func (s *tagParserState) pop() {
	if len(s.stack) < 1 {
		s.eb.AddError(fmt.Errorf("Popping empty stack"))
	} else {
		s.stack = s.stack[0 : len(s.stack)-1]
	}
}

// tagParserKeywordHandler handles the top-level keywords.
type tagParserKeywordHandler struct {
	ctx tagParserHandler
}

func (h *tagParserKeywordHandler) Handle(args tagParserArgs) {
	switch args.text {
	case "-":
		// skip
		args.state.tag.Name = args.text
	case ",":
		h.ctx = nil
	case "(":
		args.state.push(h.ctx)
		h.ctx = nil
	case "name":
		h.ctx = &tagParserNameHandler{}
	case "key":
		args.state.tag.HasKey = true
		h.ctx = &tagParserKeyHandler{}
	case "format":
		h.ctx = &tagParserFormatHandler{}
	default:
		args.state.eb.AddError(fmt.Errorf("Unknown token \"%v\"", args.text))
	}
}

// tagParserNameHandler handles the name.
type tagParserNameHandler struct {
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
