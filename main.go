package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const OffsetTop = 0
const OffsetLeft = 2

func main() {
	ctx := context.Background()

	app := NewApp(
		10,
		20,
		&Cursor{Stdout: os.Stdout},
		os.Stdout,
		os.Stdin,
		exec_,
		commandReader,
		NewAcceleratingTicker(500*time.Millisecond),
	)

	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

type App struct {
	width         int
	height        int
	field         [][]byte
	cursor        *Cursor
	stdout        io.Writer
	stdin         io.Reader
	exec          func(cmd string, args ...string) error
	commandReader func(ctx context.Context, stdin io.Reader) (<-chan Command, <-chan error)
	ticker        AppTicker
}

func NewApp(
	width, height int,
	cursor *Cursor,
	stdout io.Writer,
	stdin io.Reader,
	exec func(cmd string, args ...string) error,
	commandReader func(ctx context.Context, stdin io.Reader) (<-chan Command, <-chan error),
	ticker AppTicker,
) *App {
	return &App{
		width:         width,
		height:        height,
		field:         [][]byte{},
		cursor:        cursor,
		stdout:        stdout,
		stdin:         stdin,
		exec:          exec,
		commandReader: commandReader,
		ticker:        ticker,
	}
}

func (a *App) Start(ctx context.Context) error {
	log("starting..")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.configureTerminal(); err != nil {
		return fmt.Errorf("configure terminal: %w", err)
	}

	cmds, errc := a.commandReader(ctx, a.stdin)
	log("command reader kicked off")

	a.ticker.Start()
	defer a.ticker.Stop()

	clearScreen(a.stdout)
	a.DrawField()
	a.render()

	// todo: check if already lended to another tetro and do gameover
	tetro := a.NextTetro()

	tickCount := 0

	log("start loop")
	for {
		select {
		case cmd := <-cmds:
			log("cmd: %s", cmd)

			switch cmd {
			case Quit:
				a.cursor.SetPos(a.height+4, 0)
				fmt.Fprintln(a.stdout, "Bye")
				return nil
			case Rotate:
				tetro.Clear(a.cursor, a.stdout)
				tetro.Rotate()
				tetro.Draw(a.cursor, a.stdout)
			case Left:
				tetro.Clear(a.cursor, a.stdout)
				tetro.MoveHorizontaly(a.field, -1)
				tetro.Draw(a.cursor, a.stdout)
			case Right:
				tetro.Clear(a.cursor, a.stdout)
				tetro.MoveHorizontaly(a.field, 1)
				tetro.Draw(a.cursor, a.stdout)
			}

		case <-a.ticker.Channel():
			log("tick: %d", tickCount)
			tickCount++
			if tetro.IsLanded(a.field) {
				log("is landed")
				tetro.Cement(a.field)
				tetro = a.NextTetro()
			}
			tetro.Clear(a.cursor, a.stdout)
			tetro.MoveDown()
			tetro.Draw(a.cursor, a.stdout)

		case <-ctx.Done():
			a.cursor.SetPos(a.height+4, 0)
			fmt.Fprintln(a.stdout, "Bye")
			return nil
		case err := <-errc:
			return fmt.Errorf("read ui commands: %w", err)
		}
	}
}

// todo: impl
func (a *App) NextTetro() Tetro {
	return NewPinTetro()
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
	log("app: render")
	fmt.Fprintln(a.stdout, strings.Repeat(" ", a.width*2+4))
	for i := 1; i < len(a.field); i++ {
		fmt.Fprintf(a.stdout, "%s\n", a.field[i])
	}
}

func (a *App) configureTerminal() error {
	// disable input buffering
	err := a.exec("stty", "-f", "/dev/tty", "cbreak", "min", "1")
	if err != nil {
		return fmt.Errorf("disable input buffer: %w", err)
	}

	// do not display entered characters on the screen
	err = a.exec("stty", "-f", "/dev/tty", "-echo")
	if err != nil {
		return fmt.Errorf("disable entered characters on the screen: %w", err)
	}

	return nil
}

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

func commandReader(ctx context.Context, stdin io.Reader) (<-chan Command, <-chan error) {
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
			n, err := stdin.Read(buf)
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

type Tetro interface {
	MoveHorizontaly(field [][]byte, dir int)
	MoveDown()
	Rotate()

	// todo: Remove those and introduce Playfield
	Clear(*Cursor, io.Writer)
	Draw(*Cursor, io.Writer)
	IsLanded([][]byte) bool
	Cement([][]byte)
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
func (t *PinTetro) Clear(cursor *Cursor, stdout io.Writer) {
	for _, p := range t.Points {
		empty := ' '
		if p.bracket == ']' && p.y != OffsetTop {
			empty = '.'
		}
		cursor.SetPos(p.y+1, p.x+1)
		fmt.Fprintf(stdout, "%c", empty)
	}
}

func (t *PinTetro) Draw(cursor *Cursor, stdout io.Writer) {
	for _, p := range t.Points {
		cursor.SetPos(p.y+1, p.x+1)
		fmt.Fprintf(stdout, "%c", p.bracket)
	}
}

func (t *PinTetro) MoveDown() {
	for i := range 8 {
		t.Points[i].y += 1
	}
}

func (t *PinTetro) MoveHorizontaly(field [][]byte, dir int) {
	for _, p := range t.Points {
		nextPos := p.x + dir*2
		if nextPos < OffsetLeft || nextPos > OffsetLeft+20 {
			return
		}
		nextSymbol := field[p.y][nextPos]
		if nextSymbol == '[' || nextSymbol == ']' {
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

type Cursor struct { // todo: find better place
	Stdout io.Writer
}

// SetPos send escape sequence to the stdout.
// The line and column starts from 1 (not from 0).ikn
func (c *Cursor) SetPos(line, column int) {
	fmt.Fprintf(c.Stdout, "\033[%d;%dH", line, column)
}

func clearScreen(stdout io.Writer) {
	fmt.Fprint(stdout, "\033[H\033[2J")
}

func log(format string, a ...any) {
	if len(a) == 0 {
		fmt.Fprint(os.Stderr, format+"\n")
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
	}
}

type AppTicker interface {
	Channel() <-chan struct{}
	Start()
	Stop()
}

type AcceleratingTicker struct {
	C        chan struct{}
	stopped  chan struct{}
	duration time.Duration
}

func NewAcceleratingTicker(d time.Duration) *AcceleratingTicker {
	return &AcceleratingTicker{
		C:        make(chan struct{}),
		stopped:  make(chan struct{}),
		duration: d,
	}
}

func (t *AcceleratingTicker) Channel() <-chan struct{} {
	return t.C
}

func (t *AcceleratingTicker) Start() {
	go func() {
		timer := time.NewTimer(t.duration)
		defer timer.Stop()

		for {
			select {
			case <-t.stopped:
				log("ticker: got stopped")
				close(t.C)
				return
			case <-timer.C:
				t.C <- struct{}{}
				timer.Reset(t.duration) // todo: decrease
			}
		}
	}()
}

func (t *AcceleratingTicker) Stop() {
	log("ticker: call stop")
	close(t.stopped)
}

func exec_(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}
