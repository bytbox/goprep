/* goprep demo : nocomments : removes all comments */

package main

import (
	. "github.com/bytbox/goprep"
	"go/token"
)

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

func main() {
	tokIn, tokOut, done := StdInit()
	tokIn = Ignore(tokIn, tokOut, func (ti TokenInfo) bool {return ti.Token == token.COMMENT})
	for tok := range tokIn {
		str := tok.Str
		tokOut <- str
	}
	close(tokOut)
	<-done
}
