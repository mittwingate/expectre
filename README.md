Expectre - Expect in Go
=======================

Expectre lets you automate interactive programs, by waiting for certain
patterns in their output and then providing canned input, just like the
`expect` Unix utility, but in Go.

A script like

```
$ cat who.sh
#!/bin/sh

read -p "Who are you? " input </dev/tty
echo "Hello, $input."
```

which reads from its tty (and not from stdin) can then be easily automated with
this Go code:

```
package main

import (
	"fmt"
	"github.com/mittwingate/expectre"
	"regexp"
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

	rx := regexp.MustCompile("(.)ello,")

	res, err := exp.ExpectRegexp(rx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Match was: %v\n", res)

	exp.ExpectEOF()
}
```

With debugging turned on, as seen above, you can see what's going on internally:

```
2021/10/03 14:01:52 master open: /dev/ttys017 3 <nil>
2021/10/03 14:01:52 slave starting with /dev/ttys017
2021/10/03 14:01:52 Starting [../scripts/who.sh]
2021/10/03 14:01:52 pid 53478 started
2021/10/03 14:01:52 Expecting Who are you? ...
2021/10/03 14:01:52 read returned 13 bytes
2021/10/03 14:01:52 Found match for: Who are you? ...
2021/10/03 14:01:52 Sending Fred
 ...
2021/10/03 14:01:52 Expecting (.)ello, ...
2021/10/03 14:01:52 read returned 6 bytes
2021/10/03 14:01:52 read returned 14 bytes
2021/10/03 14:01:52 Found match for: (.)ello, ...
Match was: [[Hello, H]]
2021/10/03 14:01:52 Expecting EOF ...
2021/10/03 14:01:52 read returned 0 bytes
2021/10/03 14:01:52 received EOF
2021/10/03 14:01:52 Shutting down 53478
2021/10/03 14:01:52 Shutdown of 53478 complete.
```

Configuration
=============

* expectre.Timeout - Time to wait for on an expected string before timing out

Communication Channels
=======================

* expectre.Stdin - Send messages to stdin of the process
* expectre.Stdout - Read process output on stdout
* expectre.Stderr - Read process output on stderr
* expectre.Released - Receive a message when the process has terminated

Attributes
==========

* expectre.Cancel() // Call this function to terminate the running process
* expectre.Ended // Flag to indicate if the process has ended
* expectre.ExitCode // Exit code of the previous spawned process after it ended

Changes
=======

* v0.0.3 2024-10-22 (ms) add ExitCode attribute
* v0.0.1 2022-03-05 fix self-termination, add docs
* v0.0.0 2021-04-11 first version

Todo
====

* interact

Cheers!

Mitt Wingate, 2020, <mittwingate@gmail.com>
