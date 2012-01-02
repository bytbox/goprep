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

// Link
func Link(f func(chan Token, chan Token, chan string, chan interface{})) func(*Pipe) {
	return func(p *Pipe) {
		nIn := make(chan Token)
		go f(p.Input, nIn, p.Output, p.Sync)
		p.Input = nIn
	}
}

// Trigger
func Trigger(p func(Token) bool,
	f func(chan Token, chan Token, chan string, chan interface{})) func(*Pipe) {
	return func(p *Pipe) {

	}
}

func main() {
	p := PipeInit(os.Stdin)
	Lines(p)

	s := (*BoolStack)(nil)
	Link(func(in chan Token,
		tOut chan Token,
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
	})(p)

	PipeEnd(p, os.Stdout)
}
