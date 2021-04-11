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
```

With debugging turned on, as seen above, you can see what's going on internally:

```
2021/04/11 11:22:13 master open: /dev/ttys008 3 <nil>
2021/04/11 11:22:13 slave starting with /dev/ttys008
2021/04/11 11:22:13 Starting [./who.sh]
2021/04/11 11:22:13 pid 80199 started
2021/04/11 11:22:13 Expecting Who are you? ...
2021/04/11 11:22:13 read returned 13 bytes
2021/04/11 11:22:13 Found match for: Who are you? ...
2021/04/11 11:22:13 Sending Fred
 ...
2021/04/11 11:22:13 Expecting Hello, ...
2021/04/11 11:22:13 read returned 6 bytes
2021/04/11 11:22:13 read returned 14 bytes
2021/04/11 11:22:13 Found match for: Hello, ...
Output was: Hello, Fred.

2021/04/11 11:22:13 Expecting EOF ...
2021/04/11 11:22:13 read returned 0 bytes
2021/04/11 11:22:13 received EOF
2021/04/11 11:22:13 Shutting down 80199
2021/04/11 11:22:13 Shutdown of 80199 complete.
```

TODO
====

* expectRegexp
* interact
* document set timeout
* abort sends/reads after receiving EOF

Cheers!

Mitt Wingate, 2020, <mittwingate@gmail.com>
