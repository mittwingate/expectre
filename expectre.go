//go:build (aix || darwin || dragonfly || freebsd || (linux && !android) || netbsd || openbsd) && cgo
// +build aix darwin dragonfly freebsd linux,!android netbsd openbsd
// +build cgo

// This package is a simple (but incomplete) implementation of the unix utility
// Expect. It spawns new processes and lets the caller send messages to their
// stdin, and watch for matching output on their stdout/stderr.
package expectre

/*
#define _XOPEN_SOURCE 600
#include <fcntl.h>
#include <stdlib.h>
#include <unistd.h>
*/
import "C"

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const TimeoutDefault time.Duration = 60
const Version = "0.02"

type Expectre struct {
	Ctx      context.Context
	Cancel   context.CancelFunc // call Cancel() to terminate running process
	Stdin    chan string // Send messages to Stdin
	Stdout   chan string // Read messages from Stdout
	Stderr   chan string // Read messages from Stderr
	Released chan bool // Receive a message when the process has ended
	Timeout  time.Duration // Time to wait for expected patterns before giving up
	Debug    bool // Print additional messages on process status
	Ended    bool // Flag to indicate if process has ended
}

func New() *Expectre {
	e := Expectre{
		Released: make(chan bool),
		Timeout:  TimeoutDefault,
	}
	return &e
}

// Spawn new process, watch its stdin/out/err
func (e *Expectre) Spawn(args ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	e.Ctx = ctx
	e.Cancel = cancel

	slaveName, masterFd, err := ptyOpen()
	if e.Debug {
		log.Printf("master open: %s %d %v\n", slaveName, masterFd, err)
	}

	fmaster := os.NewFile(uintptr(masterFd), "master")

	if e.Debug {
		log.Printf("slave starting with %s\n", slaveName)
	}
	file, err := os.OpenFile(slaveName, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	e.Stdin = make(chan string)
	e.Stdout = make(chan string)

	// Read from master, and send arriving data chunks up the stdout channel
	go func() {
		data := make([]byte, 1024)
		for {
			n, err := fmaster.Read(data)
			if e.Debug {
				log.Printf("read returned %d bytes\n", n)
			}
			if err == io.EOF {
				if e.Debug {
					log.Printf("received EOF\n")
				}
				e.Cancel()
				break
			}
			if err != nil {
				panic(err)
			}
			e.Stdout <- string(data)[0:n]
		}
	}()

	// Relay messages arriving on the stdin channel to the master pty
	go func() {
		w := bufio.NewWriter(fmaster)
		for {
			select {
			case msg := <-e.Stdin:
				w.WriteString(msg)
				w.Flush()
			}
		}
	}()

	if e.Debug {
		log.Printf("Starting %v\n", args)
	}

	// Spawn process, attach stdin/out/err to pty slave
	procAttr := os.ProcAttr{
		Files: []*os.File{file, file, file},
		Sys: &syscall.SysProcAttr{
			Setsid:     true,
			Foreground: false,
			Setctty:    false,
		},
	}

	p, err := os.StartProcess(args[0], args, &procAttr)
	if err != nil {
		panic(err)
	}
	if e.Debug {
		log.Printf("pid %d started ...\n", p.Pid)
	}

	go func() {
		if e.Debug {
			log.Printf("Setting up Wait().\n")
		}
		_, err = p.Wait()
		if err != nil {
			log.Printf("Wait %d returned %v\n", p.Pid, err)
		}
		if e.Debug {
			log.Printf("Shutdown of %d complete.\n", p.Pid)
		}
		e.Ended = true
		e.Released <- true
	}()

	go func() {
		for {
			select {
			// shut down process
			case <-e.Ctx.Done():
				if e.Ended {
					return
				}
				if e.Debug {
					log.Printf("Shutting down %d\n", p.Pid)
				}
				err := p.Kill()
				if err != nil {
					log.Printf("Kill %d returned %v\n", p.Pid, err)
				}
				return
			}
		}
	}()

	return nil
}

func ptyOpen() (string, int, error) {
	m, err := C.posix_openpt(C.O_RDWR | C.O_NOCTTY)
	if err != nil {
		panic(err)
	}
	// defer C.close(m) // don't do it, master needs to prevail

	if _, err := C.grantpt(m); err != nil {
		panic(err)
	}

	if _, err := C.unlockpt(m); err != nil {
		panic(err)
	}

	slaveName, err := C.ptsname(m)

	if err != nil {
		panic(err)
	}

	return C.GoString(slaveName), int(m), nil
}

// Expect a string in stdout of the running process
func (e *Expectre) ExpectString(waitFor string) (string, error) {
	if e.Debug {
		log.Printf("Expecting %s ...", waitFor)
	}
	for {
		select {
		case line := <-e.Stdout:
			if strings.Contains(line, waitFor) {
				if e.Debug {
					log.Printf("Found match for: %s ...", waitFor)
				}
				return line, nil
			}
			continue
		case <-time.After(e.Timeout * time.Second):
			return "", errors.New(fmt.Sprintf(
				"Timeout (%ds) exceeded waiting for %s", int(e.Timeout), waitFor))
		}
	}
}

// Expect string matching regex in stdout of the running process
func (e *Expectre) ExpectRegexp(waitFor *regexp.Regexp) ([][]string, error) {
	if e.Debug {
		log.Printf("Expecting %v ...", waitFor)
	}
	for {
		select {
		case line := <-e.Stdout:
			res := waitFor.FindAllStringSubmatch(line, -1)
			if len(res) == 0 {
				continue
			}
			if e.Debug {
				log.Printf("Found match for: %v ...", waitFor)
			}
			return res, nil
			continue
		case <-time.After(e.Timeout * time.Second):
			return [][]string{}, errors.New(fmt.Sprintf(
				"Timeout (%ds) exceeded waiting for %s", int(e.Timeout), waitFor))
		}
	}
}

// Wait for the spawned process to terminate
func (e *Expectre) ExpectEOF() error {
	if e.Debug {
		log.Printf("Expecting EOF ...")
	}
	select {
	case <-e.Released:
		return nil
	case <-time.After(e.Timeout * time.Second):
		return errors.New(fmt.Sprintf(
			"Timeout (%ds) exceeded waiting for EOF", e.Timeout))
	}
}

// Send a message to stdin of the running process
func (e *Expectre) Send(msg string) error {
	if e.Debug {
		log.Printf("Sending %s ...", msg)
	}
	select {
	case e.Stdin <- msg:
		return nil
	case <-time.After(e.Timeout * time.Second):
		return errors.New(fmt.Sprintf(
			"Timeout (%ds) exceeded waiting for send", e.Timeout))
	}
}
