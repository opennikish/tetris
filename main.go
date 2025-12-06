package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

const OffsetTop = 0 // todo: add empty line on top
const OffsetLeft = 2

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := NewApp(10, 20)
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

func (t *PinTetro) Draw(field [][]byte) {
	for _, p := range t.Points {
		field[p.y][p.x] = p.bracket
	}
}

func (t *PinTetro) MoveDown() {
	for i := range 8 {
		t.Points[i].y += 1
	}
}

type App struct {
	width  int
	height int
	dir    int
	field  [][]byte
}

func NewApp(width, height int) *App {
	return &App{
		width:  width,
		height: height,
		dir:    1,
		field:  [][]byte{},
	}
}

func (a *App) Start(ctx context.Context) {
	cmds, errc := a.readCommands(ctx)
	ticker := time.NewTicker(500 * time.Millisecond) // todo: custom dynamic ticker

	a.DrawField()
	tetro := NewPinTetro()
	tickCount := 0

	for {
		select {
		case cmd := <-cmds:
			log("cmd: %s", cmd)

			switch cmd {
			case Quit:
				clearScreen()
				fmt.Println("Bye")
				return
			case Left:
				// a.dir = 1
			case Rotate:
				// a.dir = -1
				tetro.Rotate()
			case Right:
				// a.dir = 1
			}

		case <-ticker.C:
			clearScreen()
			log("tick: %d", tickCount)
			tickCount++
			a.DrawField()
			tetro.Draw(a.field)
			a.render()
			tetro.MoveDown()

		case <-ctx.Done():
			clearScreen()
			fmt.Println("bye")
			return
		case err := <-errc:
			fmt.Fprintf(os.Stderr, "read ui commands: %v", err)
			return
		}
	}
}

func (a *App) DrawField() {
	// todo: Extract to struct and make convention to access `field [][]byte` for Draw method
	a.field = a.field[0:0]
	for i := 0; i < a.height; i++ {
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
	for i := 0; i < len(a.field); i++ {
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
