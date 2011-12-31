/* goprep demo : nocomments : removes all comments */

package main

import (
	. "github.com/bytbox/goprep"
	"go/token"
)

func main() {
	tokIn, tokOut, done := StdInit()
	tokIn = IgnoreType(tokIn, tokOut, token.COMMENT)
	tokIn = Pass(tokIn, tokOut, func (TokenInfo) bool { return true })
	close(tokOut)
	<-done
}
