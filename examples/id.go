/* goprep demo : id : changes nothing */

package main

import (
	. "goprep"
)

func main() {
	p := StdInit()
	//Lines(p)
	Pass(True)(p)
	Discard(p)
}
