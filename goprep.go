// This package provides a framework for creating powerful lexical
// preprocessors for go code.
//
// TODO add 'line' directives
package goprep

import (
	"fmt"
	"io"
	"io/ioutil"
	"go/printer"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
)

// Represents a token as returned by scanner.Scanner.Scan(), with the position,
// token type, and string representation.
type TokenInfo struct {
	Pos   token.Position
	Token token.Token
	Str   string
}

// The pipeline from TokenInfo inputs to string outputs.
//
// We're using channels to serve as an abstracted for-loop, not because we want
// parallelism. All processing is synchronized with Sync.
type Pipeline struct {
	Input  <-chan TokenInfo
	Output chan<- string
	Sync   chan interface{}
}

// StdInit initializes appropriate processing channels for os.Stdin and
// os.Stdout. For the most part, this should be used instead of specific calls
// to Write and Read.
func StdInit() *Pipeline {
	tokOut, sync := Write(os.Stdout)
	tokIn := Read(os.Stdin, sync)
	return &Pipeline{tokIn, tokOut, sync}
}

// Write allows writing properly formatted go code to a given io.Writer via a
// series of token strings passed to the returned channel. The second returned
// channel will have a single nil value sent when writing is complete.
func Write(output io.Writer) (chan<- string, chan interface{}) {
	tokC := make(chan string)
	done := make(chan interface{})

	reader, writer := io.Pipe()

	// spit the tokens to the write end of the pipe
	go func(output io.WriteCloser, tokC <-chan string) {
		for tok := range tokC {
			fmt.Fprintf(output, " %s", tok)
			done <- nil
		}
		output.Close()
	}(writer, tokC)

	// parse the tokens into an AST and write to output
	go func(reader io.ReadCloser, output io.Writer, done chan interface{}) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(
			fset, "<stdin>", reader, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		printer.Fprint(output, fset, file)
		close(done)
	}(reader, output, done)

	return tokC, done
}

// Read reads from the given io.Reader and writes a series of TokenInfo objects
// to the returned channel. Read will synchronize with the given channel sync
// if it is not nil.
func Read(input io.Reader, sync <-chan interface{}) <-chan TokenInfo {
	// start reading
	src, err := ioutil.ReadAll(input)
	if err != nil { panic(err) }

	fset := token.NewFileSet()
	file := fset.AddFile("<stdin>", fset.Base(), len(src))

	s := scanner.Scanner{}
	s.Init(file, src, nil, scanner.InsertSemis | scanner.ScanComments)

	tokC := make(chan TokenInfo)

	go func(fset *token.FileSet, s scanner.Scanner, tokC chan<- TokenInfo) {
		pos, tok, str := s.Scan()
		for tok != token.EOF {
			if tok == token.COMMENT {
				str = str + "\n"
			}
			tokC <- TokenInfo{fset.Position(pos), tok, str}
			<-sync
			pos, tok, str = s.Scan()
		}
		close(tokC)
	}(fset, s, tokC)

	return tokC
}

// True takes a TokenInfo and always returns true.
func True(TokenInfo) bool { return true }

// False takes a TokenInfo and always returns false.
func False(TokenInfo) bool { return false}

// Lines passes line pragma information as needed to the output channel, thus
// ensuring that line numbers in the output match line numbers in the input.
func Lines(pipeline *Pipeline) {
	tIn := pipeline.Input
	tOut := make(chan TokenInfo)
	fname := ""
	lno := 0
	go func() {
		for tok := range tIn {
			fname = tok.Pos.Filename
			lno = tok.Pos.Line
			tOut <- tok
		}
		close(tOut)
	}()
	pipeline.Input = tOut
}

// Ignore produces a modified input stream that does not include any tokens for
// which f evaluates to true, thus discarding a certain class of tokens.
func Ignore(f func(TokenInfo) bool) func(*Pipeline) {
	return func(p *Pipeline) {
		tOut := make(chan TokenInfo)
		tIn := p.Input
		sync := p.Sync
		go func() {
			for tok := range tIn {
				if !f(tok) {
					tOut <- tok
				} else {
					sync <- nil
				}
			}
			close(tOut)
		}()
		p.Input = tOut
	}
}

// IgnoreToken is like Ignore, discarding all tokens whose string content is
// equal to the given string.
func IgnoreToken(str string) func(*Pipeline) {
	return Ignore(func(ti TokenInfo) bool {
		return ti.Str == str
	})
}

// IgnoreType is like Ignore, discarding all tokens of a certain type.
func IgnoreType(tok token.Token) func(*Pipeline) {
	return Ignore(func(ti TokenInfo) bool {
		return ti.Token == tok
	})
}

// Pass redirects all tokens for which f evaluates to true to the output
// channel, returning the altered input channel.
func Pass(f func(TokenInfo) bool) func(*Pipeline) {
	return func(p *Pipeline) {
		tOut := make(chan TokenInfo)
		tIn := p.Input
		out := p.Output
		go func() {
			for tok := range tIn {
				if f(tok) {
					out <- tok.Str
				} else {
					tOut <- tok
				}
			}
			close(tOut)
		}()
		p.Input = tOut
	}
}

// PassToken is like Pass, passing all tokens whose string content is equal to
// the given string.
func PassToken(str string) func(*Pipeline) {
	return Pass(func(ti TokenInfo) bool {
		return ti.Str == str
	})
}

// PassType is like Pass, passing all tokens of a certain type.
func PassType(tok token.Token) func(*Pipeline) {
	return Pass(func(ti TokenInfo) bool {
		return ti.Token == tok
	})
}

// Discard discards whatever tokens remain and waits for the channel to close,
// then closing the output channel as well.  It is an appropriate way to end a
// long list of processing declarations.
func Discard(p *Pipeline) {
	for _ = range p.Input {}
	close(p.Output)
	for _ = range p.Sync {}
}

