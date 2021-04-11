package expectre

import (
	"log"
	"strings"
	"testing"
)

func TestTTyinScript(t *testing.T) {
	exp := New()
	err := exp.Spawn("scripts/ttyin.sh")
	if err != nil {
		panic(err)
	}
	exp.Debug = false

	if exp.Debug {
		log.Printf("Testing ttyin")
	}

	// startup message
	_ = <-exp.Stdout

	text := <-exp.Stdout
	if text != "Input: " {
		t.Fatalf("Unexpected text: '%s'\n", text)
	}

	exp.Stdin <- "blah blah\n"

	// throw away echo
	_ = <-exp.Stdout

	text = strings.TrimRight(<-exp.Stdout, " \n\r")
	if text != "Input was: blah blah" {
		t.Fatalf("Unexpected text: '%s'\n", text)
	}

	exp.Cancel()
	<-exp.Released
}

func TestWhoScript(t *testing.T) {
	exp := New()
	err := exp.Spawn("scripts/who.sh")
	if err != nil {
		panic(err)
	}
	exp.Debug = false

	if exp.Debug {
		log.Printf("Testing who")
	}

	exp.ExpectString("Who are you?")
	exp.Send("Fred\n")
	match, err := exp.ExpectString("Hello,")
	if err != nil {
		panic(err)
	}

	out := strings.TrimRight(match, " \n\r")
	if out != "Hello, Fred." {
		t.Fatalf("Unexpected text: '%s'\n", match)
	}

	exp.ExpectEOF()
	exp.Cancel()
}
