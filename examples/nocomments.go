/* goprep demo : nocomments : removes all comments */

package main

import (
	. "goprep"
	"go/token"
)

func main() {
	tokIn, tokOut, done := StdInit()
	tokIn = IgnoreType(tokIn, tokOut, token.COMMENT)
	tokIn = Pass(tokIn, tokOut, True)
	Discard(tokIn, tokOut)

	<-done
}
