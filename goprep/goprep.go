// This package provides a framework for creating powerful lexical
// preprocessors for go code.
package goprep

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
)

// Represents a token as returned by scanner.Scanner.Scan(), with the position,
// token type, and string representation.
type Token struct {
	Pos   token.Position
	Token token.Token
	Str   string
}

// Pipe represents a unit of the pipeline from Token inputs to string
// outputs. This would probably be a monad in haskell.
//
// We're using channels to serve as an abstracted for-loop, not because we want
// parallelism. All processing is synchronized with Sync. A message to Sync
// signals that the token has been 'used' (sent to Output or discarded), and
// the sender is ready for the next. Since adding a read to sync (which is
// 'always' being read from by the frontend goroutine) would result in
// undeterministic behaviour, routines originating artificial tokens must
// create a new synchronization channel.
type Pipe struct {
	Input  chan Token
	Output chan string
	Sync   chan interface{}
}

// PipeInit initializes a pipeline; input will be read from iReader.
func PipeInit(iReader io.Reader) *Pipe {
	input := make(chan Token)
	output := make(chan string)
	sync := make(chan interface{})
	p := &Pipe{input, output, sync}

	src, err := ioutil.ReadAll(iReader)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	file := fset.AddFile("<stdin>", fset.Base(), len(src))

	s := scanner.Scanner{}
	s.Init(file, src, nil, scanner.InsertSemis|scanner.ScanComments)

	go func() {
		pos, tok, str := s.Scan()
		for tok != token.EOF {
			if tok == token.COMMENT {
				str = str + "\n"
			}
			input <- Token{fset.Position(pos), tok, str}
			<-sync // wait for sent token to land
			pos, tok, str = s.Scan()
		}
		close(input)
	}()

	return p
}

// PipeEnd implements the closing (writing) portion of a pipeline.
func PipeEnd(p *Pipe, oWriter io.Writer) {
	go func() {
		for _ = range p.Input {
			panic("Leftovers")
		}
		close(p.Output)
		close(p.Sync)
	}()

	outbuf := new(bytes.Buffer)
	output := p.Output

	// spit the tokens to the write end of the pipe
	for tok := range output {
		fmt.Fprintf(outbuf, " %s", tok)
	}

	// parse the tokens into an AST and write to output
	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset, "<stdin>", outbuf, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	printer.Fprint(oWriter, fset, file)
}

// True takes a Token and always returns true.
func True(Token) bool { return true }

// False takes a Token and always returns false.
func False(Token) bool { return false }

// Lines passes line pragma information as needed to the output channel, thus
// ensuring that line numbers in the output match line numbers in the input.
func Lines(pipeline *Pipe) {
	tIn := pipeline.Input
	tOut := make(chan Token)
	out := pipeline.Output
	isync := pipeline.Sync
	osync := make(chan interface{})
	fname := ""
	lno := 0
	pipeline.Input = tOut
	pipeline.Sync = osync
	go func() {
		for tok := range tIn {
			fname = tok.Pos.Filename
			lno = tok.Pos.Line
			tOut <- tok
			<-osync // wait for the bounce

			// now send the 'line' directive if necessary
			if tok.Str == "\n" {
				ld := fmt.Sprintf("//line %s:%d\n", fname, lno)
				out <- ld
			}

			isync <- nil
		}
		close(tOut)
	}()
}

// Ignore produces a modified input stream that does not include any tokens for
// which f evaluates to true, thus discarding a certain class of tokens.
func Ignore(f func(Token) bool) func(*Pipe) {
	return func(p *Pipe) {
		tOut := make(chan Token)
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
func IgnoreToken(str string) func(*Pipe) {
	return Ignore(func(ti Token) bool {
		return ti.Str == str
	})
}

// IgnoreType is like Ignore, discarding all tokens of a certain type.
func IgnoreType(tok token.Token) func(*Pipe) {
	return Ignore(func(ti Token) bool {
		return ti.Token == tok
	})
}

// Pass redirects all tokens for which f evaluates to true to the output
// channel, returning the altered input channel.
func Pass(f func(Token) bool) func(*Pipe) {
	return func(p *Pipe) {
		tOut := make(chan Token)
		tIn := p.Input
		out := p.Output
		sync := p.Sync
		go func() {
			for tok := range tIn {
				if f(tok) {
					out <- tok.Str
					sync <- nil
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
func PassToken(str string) func(*Pipe) {
	return Pass(func(ti Token) bool {
		return ti.Str == str
	})
}

// PassType is like Pass, passing all tokens of a certain type.
func PassType(tok token.Token) func(*Pipe) {
	return Pass(func(ti Token) bool {
		return ti.Token == tok
	})
}
