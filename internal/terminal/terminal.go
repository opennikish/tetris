package terminal

import (
	"context"
	"fmt"
	"io"
)

type KeyKind uint8

const (
	Left KeyKind = iota
	Right
	Up
	Letter
)

var kkNames = map[KeyKind]string{
	Left:   "left",
	Right:  "right",
	Up:     "up",
	Letter: "letter",
}

func (kk KeyKind) String() string {
	return kkNames[kk]
}

type Key struct {
	Kind KeyKind
	Char byte
}

type Terminal struct {
	stdin  io.Reader
	stdout io.Writer
	exec   func(cmd string, args ...string) error
}

func NewTerminal(
	stdin io.Reader,
	stdout io.Writer,
	exec func(cmd string, args ...string) error,
) *Terminal {
	return &Terminal{
		stdin:  stdin,
		stdout: stdout,
		exec:   exec,
	}
}

func (t *Terminal) Print(s string) {
	fmt.Fprint(t.stdout, s)
}

func (t *Terminal) Println(s string) {
	fmt.Fprintf(t.stdout, "%s\n", s)
}

func (t *Terminal) Clearscreen() {
	fmt.Fprint(t.stdout, "\033[H\033[2J")
}

func (t *Terminal) WatchKeystrokes(ctx context.Context) (<-chan Key, <-chan error) {
	keys, errc := make(chan Key), make(chan error, 1)
	buf := make([]byte, 3)

	keyMap := map[string]KeyKind{
		"\033[A": Up,
		"\033[C": Right,
		"\033[D": Left,
	}

	go func() {
		defer close(errc)
		defer close(keys)

		for ctx.Err() == nil {
			n, err := t.stdin.Read(buf)
			if err != nil {
				errc <- fmt.Errorf("read stdin: %w", err)
				break
			}

			if n == 1 {
				keys <- Key{
					Kind: Letter,
					Char: buf[0],
				}
				continue
			}

			if kk, ok := keyMap[string(buf[:n])]; ok {
				keys <- Key{Kind: kk}
			}
		}
	}()

	return keys, errc
}

// SetCursor send escape sequence to the stdout.
// The line and column starts from 1 (not from 0).ikn
func (t *Terminal) SetCursor(line, column int) {
	fmt.Fprintf(t.stdout, "\033[%d;%dH", line, column)
}

func (t *Terminal) MoveCursorRight(n int) {
	if n > 0 {
		fmt.Fprintf(t.stdout, "\033[%dC", n)
	}
}

func (t *Terminal) Printf(format string, a ...any) {
	fmt.Fprintf(t.stdout, format, a...)
}

func (t *Terminal) ConfigureTerminal() error {
	// disable input buffering
	err := t.exec("stty", "-f", "/dev/tty", "cbreak", "min", "1")
	if err != nil {
		return fmt.Errorf("disable input buffer: %w", err)
	}

	// do not display entered characters on the screen
	err = t.exec("stty", "-f", "/dev/tty", "-echo")
	if err != nil {
		return fmt.Errorf("disable entered characters on the screen: %w", err)
	}

	return nil
}
