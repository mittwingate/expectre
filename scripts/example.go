package main

import (
	"fmt"
	"github.com/mittwingate/expectre"
)

func main() {
	exp := expectre.New()
	exp.Debug = true

	err := exp.Spawn("./who.sh")
	if err != nil {
		panic(err)
	}

	exp.ExpectString("Who are you?")
	exp.Send("Fred\n")

	match, err := exp.ExpectString("Hello,")
	fmt.Printf("Output was: %s\n", match)

	exp.ExpectEOF()
}
