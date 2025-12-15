package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Command int

const (
	Left Command = iota
	Right
	Rotate
	Quit
)

var cmdNames = map[Command]string{
	Left:   "left",
	Right:  "right",
	Rotate: "rotate",
	Quit:   "quit",
}

func (c Command) String() string {
	return cmdNames[c]
}

const OffsetTop = 1
const OffsetLeft = 2

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := NewApp(10, 20, &Cursor{})
	app.Start(ctx)
}

type Point struct {
	x, y    int
	bracket byte
}

type PinTetro struct {
	rotationPos int
	Points      [8]Point
}

func NewPinTetro() *PinTetro {
	return &PinTetro{
		Points: [8]Point{
			{OffsetLeft + 4*2, OffsetTop, '['},
			{OffsetLeft + 4*2 + 1, OffsetTop, ']'},
			{OffsetLeft + 3*2, OffsetTop + 1, '['},
			{OffsetLeft + 3*2 + 1, OffsetTop + 1, ']'},
			{OffsetLeft + 4*2, OffsetTop + 1, '['},
			{OffsetLeft + 4*2 + 1, OffsetTop + 1, ']'},
			{OffsetLeft + 5*2, OffsetTop + 1, '['},
			{OffsetLeft + 5*2 + 1, OffsetTop + 1, ']'},
		},
	}
}

func (t *PinTetro) Rotate() {
	type Dir struct {
		x, y int
	}
	type Rule struct {
		Dirs [4]Dir
	}

	rulesByState := [4]Rule{
		{Dirs: [4]Dir{{1, 1}, {1, -1}, {0, 0}, {-1, 1}}},
		{Dirs: [4]Dir{{-1, 1}, {1, 1}, {0, 0}, {-1, -1}}},
		{Dirs: [4]Dir{{-1, -1}, {-1, 1}, {0, 0}, {1, -1}}},
		{Dirs: [4]Dir{{1, -1}, {-1, -1}, {0, 0}, {1, 1}}},
	}

	rule := rulesByState[t.rotationPos]

	for i := 0; i < len(t.Points); i += 2 {
		t.Points[i].x += rule.Dirs[i/2].x * 2
		t.Points[i+1].x += rule.Dirs[i/2].x * 2
		t.Points[i].y += rule.Dirs[i/2].y
		t.Points[i+1].y += rule.Dirs[i/2].y
	}

	t.rotationPos = (t.rotationPos + 1) % 4
}

// todo: Extract Clear and Draw somewhere, it's too much responsibility
func (t *PinTetro) Clear(cursor *Cursor) {
	for _, p := range t.Points {
		empty := ' '
		if p.bracket == ']' {
			empty = '.'
		}
		cursor.SetPos(p.y+1, p.x+1)
		fmt.Printf("%c", empty)
	}
}

func (t *PinTetro) Draw(cursor *Cursor) {
	for _, p := range t.Points {
		cursor.SetPos(p.y+1, p.x+1)
		fmt.Printf("%c", p.bracket)
	}
}

func (t *PinTetro) MoveDown() {
	for i := range 8 {
		t.Points[i].y += 1
	}
}

func (t *PinTetro) MoveHorizontaly(dir int) {
	for _, p := range t.Points {
		next := p.x + dir*2
		if next < OffsetLeft || next > OffsetLeft+20 {
			return
		}
	}

	for i := range 8 {
		t.Points[i].x += dir * 2
	}
}

func (t *PinTetro) Cement(field [][]byte) {
	for _, p := range t.Points {
		field[p.y][p.x] = p.bracket
	}
}

func (t *PinTetro) IsLanded(field [][]byte) bool {
	for _, p := range t.Points {
		nextPoint := field[p.y+1][p.x]
		if p.y == 20 || nextPoint == '[' || nextPoint == ']' {
			return true
		}
	}
	return false
}

type App struct {
	width  int
	height int
	field  [][]byte
	cursor *Cursor
}

func NewApp(width, height int, cursor *Cursor) *App {
	return &App{
		width:  width,
		height: height,
		field:  [][]byte{},
		cursor: cursor,
	}
}

type Tetro interface {
	MoveHorizontaly(dir int)
	MoveDown()
	Rotate()
}

type Drawer interface { // todo: Remove?
	Draw(field [][]byte)
}

type Cursor struct {
}

// SetPos send escape sequence to the stdout.
// The line and column starts from 1 (not from 0).
func (c *Cursor) SetPos(line, column int) {
	fmt.Printf("\033[%d;%dH", line, column)
}

func (a *App) Start(ctx context.Context) {
	cmds, errc := a.readCommands(ctx)
	ticker := time.NewTicker(500 * time.Millisecond) // todo: custom accelerating ticker

	clearScreen()
	a.DrawField()
	a.render()

	tetro := a.NextTetro()
	tetro.Draw(a.cursor)

	tickCount := 0

	for {
		select {
		case cmd := <-cmds:
			log("cmd: %s", cmd)

			switch cmd {
			case Quit:
				a.cursor.SetPos(a.height+4, 0)
				fmt.Println("Bye")
				return
			case Rotate:
				tetro.Clear(a.cursor)
				tetro.Rotate()
				tetro.Draw(a.cursor)
			case Left:
				tetro.Clear(a.cursor)
				tetro.MoveHorizontaly(-1)
				tetro.Draw(a.cursor)
			case Right:
				tetro.Clear(a.cursor)
				tetro.MoveHorizontaly(1)
				tetro.Draw(a.cursor)
			}

		case <-ticker.C:
			log("tick: %d", tickCount)
			tickCount++
			if tetro.IsLanded(a.field) {
				log("is landed")
				tetro.Cement(a.field)
				tetro = a.NextTetro()
			}
			tetro.Clear(a.cursor)
			tetro.MoveDown()
			tetro.Draw(a.cursor)

		case <-ctx.Done():
			a.cursor.SetPos(a.height+4, 0)
			fmt.Println("bye")
			return
		case err := <-errc:
			fmt.Fprintf(os.Stderr, "read ui commands: %v", err)
			return
		}
	}
}

// todo: impl
func (a *App) NextTetro() *PinTetro {
	return NewPinTetro()
}

func (a *App) Rerender(drawer Drawer) {
	a.DrawField()
	drawer.Draw(a.field)
	clearScreen()
	a.render()
}

func (a *App) DrawField() {
	// todo: Extract to struct and make convention to access `field [][]byte` for Draw method
	for i := 0; i <= a.height; i++ {
		row := []byte{'<', '!'}
		row = append(row, bytes.Repeat([]byte{' ', '.'}, a.width)...)
		row = append(row, '!', '>')
		a.field = append(a.field, row)
	}

	row := []byte{'<', '!'}
	row = append(row, bytes.Repeat([]byte{'=', '='}, a.width)...)
	row = append(row, '!', '>')
	a.field = append(a.field, row)

	row = []byte{'<', '!'}
	row = append(row, bytes.Repeat([]byte{'\\', '/'}, a.width)...)
	row = append(row, '!', '>')
	a.field = append(a.field, row)
}

func (a *App) render() {
	fmt.Printf("%s\n", strings.Repeat("  ", a.width+4))
	for i := 1; i < len(a.field); i++ {
		fmt.Printf("%s\n", a.field[i])
	}
}

func (a *App) readCommands(ctx context.Context) (<-chan Command, <-chan error) {
	// disable input buffering
	err := exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run()
	if err != nil {
		panic(err)
	}

	// do not display entered characters on the screen
	err = exec.Command("stty", "-f", "/dev/tty", "-echo").Run()
	if err != nil {
		panic(err)
	}

	cmds, errc := make(chan Command), make(chan error, 1)
	buf := make([]byte, 3)

	cmdMap := map[string]Command{
		string([]byte{27, 91, 67}): Right, // escape sequence for right arrow
		string([]byte{27, 91, 68}): Left,  // escape sequence for left arrow
		" ":                        Rotate,
		"q":                        Quit,
	}

	go func() {
		defer close(errc)
		defer close(cmds)

		for ctx.Err() == nil {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				errc <- fmt.Errorf("read stdin: %w", err)
				break
			}

			if cmd, ok := cmdMap[string(buf[:n])]; ok {
				cmds <- cmd
			}
		}
	}()

	return cmds, errc
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func log(format string, a ...any) {
	if len(a) == 0 {
		fmt.Fprint(os.Stderr, format+"\n")
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", a)
	}
}
