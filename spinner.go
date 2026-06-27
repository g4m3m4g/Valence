package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// spinner displays an animated progress indicator on stderr while work runs
// in the background. When stderr is not a TTY it degrades to a plain one-line
// message printed once (no animation, no control codes).
type spinner struct {
	msg    string
	mu     sync.Mutex
	done   chan struct{}
	tty    bool
}

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// startSpinner prints msg and begins animating. Call Stop() when the work is done.
func startSpinner(msg string) *spinner {
	s := &spinner{
		msg:  msg,
		done: make(chan struct{}),
		tty:  isTTY(os.Stderr),
	}
	if s.tty {
		go s.animate()
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
	return s
}

// SetMsg updates the message shown next to the spinner without stopping it.
func (s *spinner) SetMsg(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
	if !s.tty {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
}

// Stop halts the animation and clears the spinner line so the caller can
// print the final status cleanly on the same line position.
func (s *spinner) Stop() {
	if s.tty {
		close(s.done)
		// Give the goroutine time to clear the line before we print over it.
		time.Sleep(20 * time.Millisecond)
	}
}

func (s *spinner) animate() {
	i := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			fmt.Fprintf(os.Stderr, "\r\033[K") // clear the spinner line
			return
		case <-ticker.C:
			s.mu.Lock()
			msg := s.msg
			s.mu.Unlock()
			fmt.Fprintf(os.Stderr, "\r%s  %s", spinFrames[i%len(spinFrames)], msg)
			i++
		}
	}
}

// isTTY reports whether f is connected to an interactive terminal.
// Uses only the standard library — no external dependencies.
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
