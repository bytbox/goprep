package main

import (
	. "github.com/bytbox/goprep/goprep"
	"go/token"
	"os"
)

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

	nIn := make(chan TokenInfo)
	s := (*BoolStack)(nil)
	go func(in chan TokenInfo,
		tOut chan TokenInfo,
		out chan string,
		sync chan interface{}) {
		for tok := range in {
			if tok.Token == token.LBRACE {
				s = Push(s, false)
			}
			if tok.Token == token.RBRACE {
				var b bool
				s, b = Pop(s)
				if b {
					out <- "}}()"
					sync <- nil
					continue
				}
			}
			if tok.Str != "gofor" {
				out <- tok.Str
				sync <- nil
				continue
			}
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
	}(p.Input, nIn, p.Output, p.Sync)
	p.Input = nIn

	Discard(p)
	PipeEnd(p, os.Stdout)
}
