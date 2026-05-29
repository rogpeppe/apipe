// The apipe command pipes the contents of the current acme window
// through its argument shell command and updates them to the result
// by applying minimal changes.
//
// For example:
//
//	apipe gofmt
//
// will alter only the pieces of source code that
// have changed, leaving the rest untouched.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"unicode/utf8"

	"9fans.net/go/acme"

	"github.com/rogpeppe/apipe/internal/diff"
)

func main() {
	if err := Main(); err != nil {
		fmt.Fprintf(os.Stderr, "apipe: %v", err)
		os.Exit(1)
	}
}

func Main() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: apipe cmd [arg...]")
	}
	cmdArgs := os.Args[1:]
	win, err := acmeCurrentWin()
	if err != nil {
		return err
	}
	defer win.CloseFiles()
	var body bytes.Buffer
	if err := copyBody(&body, win); err != nil {
		return err
	}
	pcmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	pcmd.Stdin = bytes.NewReader(body.Bytes())
	pcmd.Stderr = os.Stderr
	newBody, err := pcmd.Output()
	if err != nil {
		return fmt.Errorf("cannot run %q: %v", cmdArgs[0], err)
	}
	return applyDiff(body.Bytes(), newBody, func(addr string, data []byte) error {
		if _, err := win.Write("addr", []byte(addr)); err != nil {
			return fmt.Errorf("cannot set address %q: %v", addr, err)
		}
		if err := writeData(win, data); err != nil {
			return fmt.Errorf("cannot write data: %v", err)
		}
		return nil
	})
}

func writeData(win *acme.Win, data []byte) error {
	if len(data) == 0 {
		_, err := win.Write("data", nil)
		return err
	}
	for len(data) > 0 {
		d := data
		if len(d) > 8000 {
			d = trimIncompleteRune(d[0:8000])
		}
		n, err := win.Write("data", d)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

// trimIncompleteRune returns b with any trailing
// incomplete rune sliced off.
func trimIncompleteRune(b []byte) []byte {
	i := len(b) - utf8.UTFMax
	if i < 0 {
		i = 0
	}
	lastStart := len(b)
	for ; i < len(b); i++ {
		if r, n := utf8.DecodeRune(b[i:]); r != utf8.RuneError || n > 1 {
			lastStart = len(b)
			continue
		}
		if utf8.RuneStart(b[i]) {
			lastStart = i
		}
	}
	return b[0:lastStart]
}

// applyDiff computes a byte-level diff between old and new, then calls
// edit for each change with an acme address and the replacement data.
func applyDiff(old, new []byte, edit func(addr string, data []byte) error) error {
	edits := diff.Bytes(old, new)

	runeAdj := 0
	bytePos := 0
	runePos := 0
	for _, e := range edits {
		runePos += utf8.RuneCount(old[bytePos:e.Start])
		oldRuneCount := utf8.RuneCount(old[e.Start:e.End])
		newRuneCount := utf8.RuneCountInString(e.New)
		bytePos = e.End

		adjStart := runePos + runeAdj
		adjEnd := adjStart + oldRuneCount
		addr := fmt.Sprintf("#%d,#%d", adjStart, adjEnd)
		if err := edit(addr, []byte(e.New)); err != nil {
			return err
		}

		runePos += oldRuneCount
		runeAdj += newRuneCount - oldRuneCount
	}
	return nil
}
