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

	"github.com/opennikish/tetris/internal/game"
	"github.com/opennikish/tetris/internal/terminal"
	"github.com/opennikish/tetris/internal/ui"
)

// todo: Find better way awaiting for rendered state instead of time.Sleep(). Consider extend app with render hooks or maybe add hook around ScreenBuffer.

func TestHorizontalMoveDoesNotCrossWalls(t *testing.T) {
	stdout := NewScreenBuffer(25)
	stdin, stdinWriter := io.Pipe()
	defer stdinWriter.Close()

	ticker := NewTestTicker()
	term := terminal.NewTerminal(stdin, stdout, func(cmd string, args ...string) error { return nil })
	app := NewApp(
		game.NewGameplay(),
		term,
		ui.NewPlayfieldRenderer(term, 0, 0),
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

	cmdController.PressLeft(10)

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

	cmdController.PressRight(10)

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
	term := terminal.NewTerminal(stdin, stdout, func(cmd string, args ...string) error { return nil })
	app := NewApp(
		game.NewGameplay(),
		term,
		ui.NewPlayfieldRenderer(term, 0, 0),
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
	cmdController.PressRotate(1)
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
	cmdController.PressRotate(1)
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
	cmdController.PressRotate(1)
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

	cmdController.PressRotate(1)
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
	term := terminal.NewTerminal(stdin, stdout, func(cmd string, args ...string) error { return nil })
	app := NewApp(
		game.NewGameplay(),
		term,
		ui.NewPlayfieldRenderer(term, 0, 0),
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

func TestActiveTetrominoDoesNotCrossCementedTetrominos(t *testing.T) {
	stdout := NewScreenBuffer(25)
	stdin, stdinWriter := io.Pipe()
	defer stdinWriter.Close()

	ticker := NewTestTicker()
	term := terminal.NewTerminal(stdin, stdout, func(cmd string, args ...string) error { return nil })
	app := NewApp(
		game.NewGameplay(),
		term,
		ui.NewPlayfieldRenderer(term, 0, 0),
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

	ticker.Tick(19)
	time.Sleep(2 * time.Millisecond)
	expected := `                        
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
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual := stdout.String()
	eq(t, expected, actual)

	ticker.Tick(15)
	cmdController.PressRotate(2)
	ticker.Tick(1)
	time.Sleep(2 * time.Millisecond)
	expected = `                        
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
<! . . .[][][] . . . .!>
<! . . . .[] . . . . .!>
<! . . . .[] . . . . .!>
<! . . .[][][] . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	ticker.Tick(1)
	cmdController.PressRight(3)
	time.Sleep(2 * time.Millisecond)
	expected = `                        
<! . . . . . . .[] . .!>
<! . . . . . .[][][] .!>
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
<! . . .[][][] . . . .!>
<! . . . .[] . . . . .!>
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
<! . . .[][][] . . . .!>
<! . . . .[] . .[] . .!>
<! . . . .[] .[][][] .!>
<! . . .[][][] . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)

	cmdController.PressLeft(10)
	time.Sleep(2 * time.Millisecond)
	expected = `                        
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
<! . . .[][][] . . . .!>
<! . . . .[] .[] . . .!>
<! . . . .[][][][] . .!>
<! . . .[][][] . . . .!>
<!====================!>
<!\/\/\/\/\/\/\/\/\/\/!>
`
	actual = stdout.String()
	eq(t, expected, actual)
}

func eq[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Fatalf("expected: %v got: %v", expected, actual)
	}
}

type TestTicker struct {
	C chan time.Time
}

func NewTestTicker() *TestTicker {
	return &TestTicker{
		C: make(chan time.Time),
	}
}

func (t *TestTicker) Channel() <-chan time.Time {
	return t.C
}

func (t *TestTicker) Start() {}

func (t *TestTicker) Stop() {}

func (t *TestTicker) Tick(n int) {
	for range n {
		t.C <- time.Time{}
	}
}

func (t *TestTicker) Reset(d time.Duration) {}

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

func (c *CommandController) PressLeft(n int) {
	for range n {
		c.stdinWriter.Write([]byte{27, 91, 68})
	}
}

func (c *CommandController) PressRight(n int) {
	for range n {
		c.stdinWriter.Write([]byte{27, 91, 67})
	}
}

func (c *CommandController) PressRotate(n int) {
	for range n {
		c.stdinWriter.Write([]byte("\033[A"))
	}
}

func (c *CommandController) PressQuite(n int) {
	for range n {
		c.stdinWriter.Write([]byte("q"))
	}
}
