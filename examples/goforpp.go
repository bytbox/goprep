// This command adds a 'gofor' keyword to the go syntax.
package main

import (
	. "github.com/bytbox/goprep/goprep"
	"go/token"
	"os"
)

// A stack of bools. Not thread-safe.
type BoolStack struct {
	Val  bool
	Base *BoolStack
}

func Push(s *BoolStack, i bool) *BoolStack {
	return &BoolStack{i, s}
}

func Peek(s *BoolStack) bool {
	return s.Val
}

func Pop(s *BoolStack) (*BoolStack, bool) {
	return s.Base, s.Val
}

func main() {
	p := PipeInit(os.Stdin)
	Lines(p)

	s, b := (*BoolStack)(nil), false
	Link(func(in chan Token,
		tOut chan Token,
		out chan string,
		sync chan interface{}) {
		for tok := range in {
			if tok.Token == token.LBRACE {
				// an lbrace not associated with gofor
				s = Push(s, false)
			}
			if tok.Token == token.RBRACE {
				s, b = Pop(s)
				if b {
					// end the gofor
					out <- "}}()"
					sync <- nil
					continue
				}
			}
			if tok.Str != "gofor" {
				// ignore
				out <- tok.Str
				sync <- nil
				continue
			}

			// this is a gofor. Print the first part and push onto
			// the stack
			out <- "go func() { for"
			sync <- nil
			tok = <-in
			for tok.Token != token.LBRACE {
				out <- tok.Str
				sync <- nil
				tok = <-in
			}
			out <- " {"
			s = Push(s, true)
			sync <- nil
		}
		close(tOut)
	})(p)

	PipeEnd(p, os.Stdout)
}
