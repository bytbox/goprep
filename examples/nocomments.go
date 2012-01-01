/* goprep demo : nocomments : removes all comments */

package main

import (
	. "goprep"
	"go/token"
	"os"
)

func main() {
	p := PipeInit(os.Stdin)
	Lines(p)
	IgnoreType(token.COMMENT)(p)
	Pass(True)(p)
	Discard(p)
	PipeEnd(p, os.Stdout)
}
