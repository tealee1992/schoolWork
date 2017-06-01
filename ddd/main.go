// ddd project main.go
package main

import (
	"fmt"
)

type as struct {
	ss int
}

var s = as{
	ss: 1,
}

func main() {
	var f as
	f = s

	fmt.Println(s.ss)
	f.ss = 2
	fmt.Println(f.ss)
	fmt.Println(s.ss)
}
