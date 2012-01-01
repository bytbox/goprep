/* goprep demo : nocomments : removes all comments */

package main

import (
	. "goprep"
	"go/token"
)

func main() {
	p := StdInit()
	IgnoreType(token.COMMENT)(p)
	Pass(True)(p)
	Discard(p)
}
