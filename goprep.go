// This package provides a framework for creating powerful lexical
// preprocessors for go code.
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
	Pos   token.Pos
	Token token.Token
	Str   string
}

// StdInit initializes appropriate processing channels for os.Stdin and
// os.Stdout. For the most part, this should be used instead of specific calls
// to Write and Read.
func StdInit() (<-chan TokenInfo, chan<- string, <-chan interface{}) {
	tokIn := Read(os.Stdin)
	tokOut, done := Write(os.Stdout)
	return tokIn, tokOut, done
}

// Write allows writing properly formatted go code to a given io.Writer via a
// series of token strings passed to the returned channel. The second returned
// channel will have a single nil value sent when writing is complete.
func Write(output io.Writer) (chan<- string, <-chan interface{}) {
	tokC := make(chan string)
	done := make(chan interface{})

	reader, writer := io.Pipe()

	// spit the tokens to the write end of the pipe
	go func(output io.WriteCloser, tokC <-chan string) {
		for tok := range tokC {
			fmt.Fprintf(output, " %s", tok)
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
		done <- nil
	}(reader, output, done)

	return tokC, done
}

// Read reads from the given io.Reader and writes a series of TokenInfo objects
// to the returned channel.
func Read(input io.Reader) <-chan TokenInfo {
	// start reading
	src, err := ioutil.ReadAll(input)
	if err != nil { panic(err) }

	fset := token.NewFileSet()
	file := fset.AddFile("<stdin>", fset.Base(), len(src))

	s := scanner.Scanner{}
	s.Init(file, src, nil, scanner.InsertSemis | scanner.ScanComments)

	tokC := make(chan TokenInfo)

	go func(s scanner.Scanner, tokC chan<- TokenInfo) {
		pos, tok, str := s.Scan()
		for tok != token.EOF {
			if tok == token.COMMENT {
				str = str + "\n"
			}
			tokC <- TokenInfo{pos, tok, str}
			pos, tok, str = s.Scan()
		}
		close(tokC)
	}(s, tokC)

	return tokC
}

// Ignore produces a modified input stream that does not include any tokens for
// which f evaluates to true.
func Ignore(tIn <-chan TokenInfo, out chan<- string, f func(TokenInfo) bool) <-chan TokenInfo {
	tOut := make(chan TokenInfo)
	go func() {
		for tok := range tIn {
			if !f(tok) {
				tOut <- tok
			}
		}
		close(tOut)
	}()
	return tOut
}
