/* goprep demo : id : changes nothing */

package main

import (
	. "goprep"
	"os"
)

func main() {
	p := PipeInit(os.Stdin)
	Lines(p)
	Pass(True)(p)
	PipeEnd(p, os.Stdout)
}
