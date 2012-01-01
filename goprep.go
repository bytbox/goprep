// This package provides a framework for creating powerful lexical
// preprocessors for go code.
//
// TODO add 'line' directives
package goprep

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"go/printer"
	"go/parser"
	"go/scanner"
	"go/token"
)

// Represents a token as returned by scanner.Scanner.Scan(), with the position,
// token type, and string representation.
type TokenInfo struct {
	Pos   token.Position
	Token token.Token
	Str   string
}

// The pipeline from TokenInfo inputs to string outputs. This would probably be
// a monad in haskell.
//
// We're using channels to serve as an abstracted for-loop, not because we want
// parallelism. All processing is synchronized with Sync. A message to Sync
// signals that the token has been 'used', and the sender is ready for the
// next. Since adding a read to sync (which is 'always' being read from by the
// frontend goroutine) would result in undeterministic behaviour, routines
// originating artificial tokens must create a new synchronization channel.
type Pipeline struct {
	Input  chan TokenInfo
	Output chan string
	Sync   chan interface{}
}

// PipeInit initializes a pipeline; input will be read from iReader.
func PipeInit(iReader io.Reader) *Pipeline {
	input := make(chan TokenInfo)
	output := make(chan string)
	sync := make(chan interface{})
	p := &Pipeline{ input, output, sync }

	src, err := ioutil.ReadAll(iReader)
	if err != nil { panic(err) }

	fset := token.NewFileSet()
	file := fset.AddFile("<stdin>", fset.Base(), len(src))

	s := scanner.Scanner{}
	s.Init(file, src, nil, scanner.InsertSemis | scanner.ScanComments)

	go func() {
		pos, tok, str := s.Scan()
		for tok != token.EOF {
			if tok == token.COMMENT {
				str = str + "\n"
			}
			input <- TokenInfo{fset.Position(pos), tok, str}
			<-sync // wait for sent token to land
			pos, tok, str = s.Scan()
		}
		close(input)
	}()


	return p
}

// PipeEnd implements the closing (writing) portion of a pipeline.
func PipeEnd(p *Pipeline, oWriter io.Writer) {
	outbuf := new(bytes.Buffer)
	_, output, sync := p.Input, p.Output, p.Sync

	// spit the tokens to the write end of the pipe
	for tok := range output {
		fmt.Fprintf(outbuf, " %s", tok)
		sync <- nil
	}

	// parse the tokens into an AST and write to output
	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset, "<stdin>", outbuf, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	printer.Fprint(oWriter, fset, file)
	close(sync)
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
				<-osync
			}

			isync <- nil
		}
		close(tOut)
	}()
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
	input, output, sync := p.Input, p.Output, p.Sync
	go func() {
		for _ = range input {
			sync <- nil
		}
		close(output)
	}()
}
