package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// todo: Find better way awaiting for rendered state instead of time.Sleep(). Consider extend app with render hooks or maybe add hook around ScreenBuffer.

func TestHorizontalMoveDoesNotCrossWalls(t *testing.T) {
	stdout := NewScreenBuffer(25)
	stdin, stdinWriter := io.Pipe()
	defer stdinWriter.Close()

	ticker := NewTestTicker()

	app := NewApp(
		10,
		20,
		&Cursor{Stdout: stdout},
		stdout,
		stdin,
		func(cmd string, args ...string) error { return nil },
		commandReader,
		ticker,
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	go func() {
		err := app.Start(ctx)
		if err != nil {
			log("app.Start() returned err: %s", err)
		}
	}()

	cmdController := NewCommandController(stdinWriter)

	ticker.Tick(2)
	time.Sleep(1 * time.Millisecond)
	expected := `                        
<! . . . . . . . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual := stdout.String()
	eq(t, expected, actual)

	for range 10 {
		cmdController.PressLeft()
	}
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . . . . . . . .!>
<! .[] . . . . . . . .!>
<![][][] . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	for range 10 {
		cmdController.PressRight()
	}
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . . . . . . . .!>
<! . . . . . . . .[] .!>
<! . . . . . . .[][][]!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)
}

func TestFallDownWithRotationPinTetro(t *testing.T) {
	stdout := NewScreenBuffer(25)
	stdin, stdinWriter := io.Pipe()
	defer stdinWriter.Close()

	ticker := NewTestTicker()

	app := NewApp(
		10,
		20,
		&Cursor{Stdout: stdout},
		stdout,
		stdin,
		func(cmd string, args ...string) error { return nil },
		commandReader,
		ticker,
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	go func() {
		err := app.Start(ctx)
		if err != nil {
			log("app.Start() returned err: %s", err)
		}
	}()

	cmdController := NewCommandController(stdinWriter)

	ticker.Tick(1)
	time.Sleep(1 * time.Millisecond)
	expected := `                        
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual := stdout.String()
	eq(t, expected, actual)

	ticker.Tick(1)
	cmdController.PressRotate()
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . . . . . . . .!>
<! . . . .[] . . . . .!>
<! . . . .[][] . . . .!>
<! . . . .[] . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	ticker.Tick(1)
	cmdController.PressRotate()
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . .[][][] . . . .!>
<! . . . .[] . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	ticker.Tick(1)
	cmdController.PressRotate()
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][] . . . . .!>
<! . . . .[] . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	cmdController.PressRotate()
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)
}

func TestCementAfterFallDownToTheGroundOrAnotherTetromino(t *testing.T) {
	stdout := NewScreenBuffer(25)
	stdin, stdinWriter := io.Pipe()
	defer stdinWriter.Close()

	ticker := NewTestTicker()

	app := NewApp(
		10,
		20,
		&Cursor{Stdout: stdout},
		stdout,
		stdin,
		func(cmd string, args ...string) error { return nil },
		commandReader,
		ticker,
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	go func() {
		err := app.Start(ctx)
		if err != nil {
			log("app.Start() returned err: %s", err)
		}
	}()

	ticker.Tick(1)
	time.Sleep(1 * time.Millisecond)
	expected := `                        
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual := stdout.String()
	eq(t, expected, actual)

	ticker.Tick(19)
	time.Sleep(1 * time.Millisecond)
	expected = `                        
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	ticker.Tick(17)
	time.Sleep(2 * time.Millisecond)
	expected = `                        
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . . . . . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)
}

func eq[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Fatalf(fmt.Sprintf("expected: %v got: %v", expected, actual))
	}
}

type TestTicker struct {
	C chan struct{}
}

func NewTestTicker() *TestTicker {
	return &TestTicker{
		C: make(chan struct{}),
	}
}

func (t *TestTicker) Channel() <-chan struct{} {
	return t.C
}

func (t *TestTicker) Start() {}

func (t *TestTicker) Stop() {}

func (t *TestTicker) Tick(n int) {
	for range n {
		t.C <- struct{}{}
	}
}

type ScreenBuffer struct {
	bytes     []byte
	pos       int
	maxColLen int
	mu        sync.Mutex
}

func NewScreenBuffer(maxColLen int) *ScreenBuffer {
	return &ScreenBuffer{maxColLen: maxColLen}
}

func (b *ScreenBuffer) Write(source []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if bytes.Equal(source, []byte("\033[H\033[2J")) {
		log("test: clearscreen")
		return len(source), nil
	}

	if bytes.Contains(source, []byte("\033[")) {
		row, col := b.extractPos(string(source))

		// row and col in escape sequence starts from 1, not from zero
		row -= 1
		col -= 1

		b.pos = row*b.maxColLen + col

		return len(source), nil
	}

	if len(source)+b.pos > len(b.bytes) {
		b.bytes = append(b.bytes, make([]byte, len(source)+b.pos-len(b.bytes))...)
	}

	copy(b.bytes[b.pos:b.pos+len(source)], source)
	b.pos += len(source)

	return len(source), nil
}

func (b *ScreenBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.bytes)
}

// extractPos extracts line and column from the escape sequence: "\033[{line};{col}H"
func (b *ScreenBuffer) extractPos(s string) (int, int) {
	sep := strings.Index(s, ";")
	rawRow := s[2:sep]
	rawCol := s[sep+1 : len(s)-1]

	row, err1 := strconv.Atoi(rawRow)
	col, err2 := strconv.Atoi(rawCol)
	if err1 != nil || err2 != nil {
		panic(fmt.Sprintf("extractPos, given: %q" + s))
	}

	return row, col
}

type CommandController struct {
	stdinWriter io.Writer
}

func NewCommandController(stdinWriter io.Writer) *CommandController {
	return &CommandController{stdinWriter: stdinWriter}
}

func (c *CommandController) PressLeft() {
	c.stdinWriter.Write([]byte{27, 91, 68})
}

func (c *CommandController) PressRight() {
	c.stdinWriter.Write([]byte{27, 91, 67})
}

func (c *CommandController) PressRotate() {
	c.stdinWriter.Write([]byte(" "))
}

func (c *CommandController) PressQuite() {
	c.stdinWriter.Write([]byte("q"))
}
