package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type TokenType uint32

const (
	ErrorToken TokenType = iota
	// A KeywordToken looks like '*Part/Node/Xxx'.
	KeywordToken
	// A EndKeywordToken looks like '*End xxx'.
	EndKeywordToken
	// A CommentToken looks like '** xxx'.
	CommentToken

	// PartToken
	// NodeToken
	// ElementToken
	// NsetToken
	// ElsetToken
	// OutputToken
	// NodeOutputToken
)

// An Attribute is an attribute key-value triple.
type Attribute struct {
	Key, Val string
}

type Token struct {
	// Type is as TokenType defined
	Type       TokenType
	// Name is the keyword name when TokenType=KeywordToken
	Name       string
	// Data is the data lines followed when TokenType=KeywordToken
	Data       string
	// Attr is potential parameters
	Attr       []Attribute
}

// paramString returns a string representation of a Keyword Token's params
func (t Token) paramString() string {
	s := ""
	for _, p := range t.Attr {
		s += fmt.Sprintf(", %v", p.Key)
		if len(p.Val) > 0 {
			s += fmt.Sprintf("=%v", p.Val)
		}
	}
	return s
}

// String returns a string representation of the Token.
func (t Token) String() string {
	switch t.Type {
	case ErrorToken:
		return ""
	case CommentToken, EndKeywordToken:
		return t.Data
	case KeywordToken:
		// *Nset, nset=Set-2_Vz, generate
		return fmt.Sprintf("*%s%s\n%s", t.Name, t.paramString(), t.Data)
	}
	return "Invalid(" + strconv.Itoa(int(t.Type)) + ")"
}

// span is a range of bytes in a Tokenizer's buffer.
// [start, end)
type span struct {
	start, end int
}

type Tokenizer struct {
	// r is the source of the inp text.
	r io.Reader
	// tt is the TokenType of the current token.
	tt TokenType
	// err is the first error encountered during tokenization.
	err error
	// readErr is the error returned by the io.Reader r. It is separate from
	// err because it is valid for an io.Reader to return (n int, err1 error)
	// such that n > 0 && err1 != nil, and callers should always process the
	// n > 0 bytes before considering the error err1.
	readErr error
	// buf[raw.start:raw.end] holds the raw bytes of the current token.
	// buf[raw.end:] is buffered input that will yield future tokens.
	raw span
	buf []byte
	// buf[name.start:name.end] holds the raw bytes of current token's name:
	// if it's a KeywordToken
	name span
	// buf[data.start:data.end] holds the raw bytes of the current token's data:
	// if it's a KeywordToken and has data followed
	data span
	// buf[params.start:params.end] holds the raw bytes of params, not parsed yet.
	params span
}

// Err returns the error associated with the most recent ErrorToken token.
// This is typically io.EOF, meaning the end of tokenization.
func (z *Tokenizer) Err() error {
	if z.tt != ErrorToken {
		return nil
	}
	return z.err
}

func (z *Tokenizer) readByte() byte {
	if z.raw.end >= len(z.buf) {
		// Our buffer is exhausted and we have to read from z.r.
		// Check if the previous read resulted in an error.
		if z.readErr != nil {
			z.err = z.readErr
			return 0
		}
		// We copy z.buf[z.raw.start:z.raw.end] to the beginning of z.buf. If the length
		// z.raw.end - z.raw.start is more than half the capacity of z.buf, then we
		// allocate a new buffer before the copy.
		c := cap(z.buf)
		d := z.raw.end - z.raw.start
		var buf1 []byte
		if 2*d > c {
			buf1 = make([]byte, d, 2*c)
		} else {
			buf1 = z.buf[:d]
		}
		copy(buf1, z.buf[z.raw.start:z.raw.end])
		if x := z.raw.start; x != 0 {
			z.data.start -= x
			z.data.end -= x
			// TODO pendingAttr, attr
		}
		z.raw.start, z.raw.end, z.buf = 0, d, buf1[:d]
		// Now that we have copied the live bytes to the start of the buffer,
		// we read from z.r into the remainder.
		var n int
		n, z.readErr = readAtLeastOneByte(z.r, buf1[d:cap(buf1)])
		if n == 0 {
			z.err = z.readErr
			return 0
		}
		z.buf = buf1[:d+n]
	}

	x := z.buf[z.raw.end]
	z.raw.end++
	return x
}

// readAtLeastOneByte wraps an io.Reader so that reading cannot return (0, nil).
// It returns io.ErrNoProgress if the underlying r.Read method returns (0, nil)
// too many times in succession.
func readAtLeastOneByte(r io.Reader, b []byte) (int, error) {
	for i := 0; i < 100; i++ {
		if n, err := r.Read(b); n != 0 || err != nil {
			return n, err
		}
	}
	return 0, io.ErrNoProgress
}

// skipWhiteSpace skips past any white space.
func (z *Tokenizer) skipWhiteSpace() {
	if z.err != nil {
		return
	}
	for {
		c := z.readByte()
		if z.err != nil {
			return
		}
		switch c {
		case ' ':
			// No-op.
		default:
			z.raw.end--
			return
		}
	}
}

// readUntilNextOption reads until the next "*".
// The leading two bytes ("**" or "*E") has been consumed.
func (z *Tokenizer) readUntilLineBreak() {
	z.data.start = z.raw.end - len("**")

	if z.tt == EndKeywordToken {
		z.readKeywordName()
	}

	for {
		c := z.readByte()
		if z.err != nil {
			z.data.end = z.raw.end
			return
		}
		if c == '\n' {
			z.data.end = z.raw.end
			return
		}
	}
}

// readKeywordWithData reads the next keyword line and its following data (if any).
// The opening "*" has already been consumed.
func (z *Tokenizer) readKeywordWithData() TokenType {
	z.readKeywordName()
	if z.skipWhiteSpace(); z.err != nil {
		return ErrorToken
	}

	var lineNo = 0
	// go through keyword line, and stop at the line break
	for {
		c := z.readByte()
		if z.err != nil {
			return ErrorToken
		}

		if c == '\n' {
			lineNo++
			break
		}
	}

	// delimit params (might be none, will validate later)
	z.params.start = z.name.end + len(", ")
	z.params.end = z.raw.end - 1

	// delimit data
loop:
	for {
		c := z.readByte()
		if z.err != nil {
			return ErrorToken
		}

		switch c {
		case '\n':
			lineNo++
		case '*':
			z.raw.end--
			break loop
		}
	}
	if lineNo > 1 {
		z.data.start = z.params.end + 1
		z.data.end = z.raw.end
	}

	return KeywordToken
}

// readKeywordName sets z.data to, for example, the "Node" in "*Node", or the
// "Rate Dependent" in "*Rate Dependent". The leading "*" has already been consumed.
func (z *Tokenizer) readKeywordName() {
	z.name.start = z.raw.end - 1
	for {
		c := z.readByte()
		if z.err != nil {
			z.name.end = z.raw.end
			return
		}
		switch c {
		case ',':
			z.name.end = z.raw.end - 1
			return
		case '\n':
			z.raw.end--
			z.name.end = z.raw.end
			return
		}
	}
}

// Next scans the next token and returns its type.
func (z *Tokenizer) Next() TokenType {
	z.raw.start = z.raw.end
	z.data.start = z.raw.end
	z.data.end = z.raw.end
	if z.err != nil {
		z.tt = ErrorToken
		return z.tt
	}

loop:
	for {
		c := z.readByte()
		if z.err != nil {
			break loop
		}
		if c != '*' {
			continue loop
		}

		// check if the '*' we have just read is leading a keyword line
		// if not ...
		var tokenType = KeywordToken
		c = z.readByte()
		if z.err != nil {
			break loop
		}
		switch c {
		case '*':
			tokenType = CommentToken
		case 'E', 'e':
			c0 := z.readByte()
			if z.err != nil { break loop }
			c1 := z.readByte()
			if z.err != nil { break loop }
			if c0 == 'n' && c1 == 'd' {
				tokenType = EndKeywordToken
				z.raw.end -= len("nd")
			} else {
				z.raw.end -= len("End")
			}
		}

		switch tokenType {
		case KeywordToken:
			z.tt = z.readKeywordWithData()
			return z.tt
		case CommentToken, EndKeywordToken:
			z.readUntilLineBreak()
			z.tt = tokenType
			return z.tt
		}
	}
	// TODO
	// corner case
	z.tt = ErrorToken
	return z.tt
}

func (z *Tokenizer) Token() Token {
	t := Token{Type: z.tt}

	switch z.tt {
	case CommentToken:
		t.Data = string(z.buf[z.data.start:z.data.end])
	case EndKeywordToken:
		t.Data = string(z.buf[z.data.start:z.data.end])
		t.Name = string(z.buf[z.name.start:z.name.end])
	case KeywordToken:
		t.Data = string(z.buf[z.data.start:z.data.end])
		t.Name = string(z.buf[z.name.start:z.name.end])
		if z.params.start < z.params.end {
			sParams := string(z.buf[z.params.start:z.params.end])
			for _, s := range strings.Split(sParams, ",") {
				kv := strings.Split(s, "=")
				switch len(kv) {
				case 0:
					continue
				case 1:
					t.Attr = append(t.Attr, Attribute{Key: kv[0]})
				case 2:
					t.Attr = append(t.Attr, Attribute{Key: kv[0], Val: kv[1]})
				}
			}
		}
	}

	return t
}


func (z *Tokenizer) readKeyword() TokenType {

	// TODO recognize self-closing token like "*End Part"

	return KeywordToken
}

func NewTokenizer(r io.Reader) *Tokenizer {
	return &Tokenizer{
		r: r,
		buf: make([]byte, 0, 10240),
	}
}
