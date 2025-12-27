package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
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
		NewTerminalScreen(os.Stdout),
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
	screen        *TerminalScreen
	stdin         io.Reader
	exec          func(cmd string, args ...string) error
	commandReader func(ctx context.Context, stdin io.Reader) (<-chan Command, <-chan error)
	ticker        AppTicker
	playfield     *Playfield
	currTetro     *Tetromino
	tickCount     int
	ctxCancel     context.CancelFunc
}

func NewApp(
	width, height int,
	screen *TerminalScreen,
	stdin io.Reader,
	exec func(cmd string, args ...string) error,
	commandReader func(ctx context.Context, stdin io.Reader) (<-chan Command, <-chan error),
	ticker AppTicker,
) *App {
	return &App{
		width:         width,
		height:        height,
		screen:        screen,
		stdin:         stdin,
		exec:          exec,
		commandReader: commandReader,
		ticker:        ticker,
	}
}

func (a *App) Start(ctx context.Context) error {
	log("starting..")
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	a.ctxCancel = stop
	defer stop()

	if err := a.configureTerminal(); err != nil {
		return fmt.Errorf("configure terminal: %w", err)
	}

	cmds, errc := a.commandReader(ctx, a.stdin)
	log("command reader kicked off")

	a.ticker.Start()
	defer a.ticker.Stop()

	a.screen.Clearscreen()

	a.playfield = NewPlayfield(a.width, a.height)
	a.playfield.Init()
	a.playfield.Render(a.screen)

	a.currTetro = a.spawnNext()

	log("start loop")
	for {
		select {
		case cmd := <-cmds:
			a.onInput(cmd)
		case <-a.ticker.Channel():
			a.onTick()
		case <-ctx.Done():
			a.screen.SetCursor(a.height+4, 0)
			a.screen.Print("Bye")
			return nil
		case err := <-errc:
			return fmt.Errorf("read ui commands: %w", err)
		}
	}
}

func (a *App) onTick() {
	log("tick: %d", a.tickCount)
	a.tickCount++

	if a.playfield.IsLanded(a.currTetro) {
		log("tetro landed")
		a.playfield.LockDown(a.currTetro)
		a.currTetro = a.spawnNext()
		if !a.playfield.CanPlace(a.currTetro) {
			a.quit()
		}
	}

	a.playfield.ClearTetro(a.screen, a.currTetro)
	a.currTetro.MoveVert(1)
	a.playfield.DrawTetro(a.screen, a.currTetro)
}

func (a *App) onInput(cmd Command) {
	log("cmd: %s", cmd)

	switch cmd {
	case Quit:
		a.quit()
	case Rotate:
		a.playfield.ClearTetro(a.screen, a.currTetro)
		a.currTetro.Rotate()
		if !a.playfield.CanPlace(a.currTetro) {
			for range 3 {
				a.currTetro.Rotate()
			}
		}
		a.playfield.DrawTetro(a.screen, a.currTetro)
	case Left:
		a.playfield.ClearTetro(a.screen, a.currTetro)
		a.currTetro.MoveHorizontaly(-1)
		if !a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveHorizontaly(1)
		}
		a.playfield.DrawTetro(a.screen, a.currTetro)
	case Right:
		a.playfield.ClearTetro(a.screen, a.currTetro)
		a.currTetro.MoveHorizontaly(1)
		if !a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveHorizontaly(-1)
		}
		a.playfield.DrawTetro(a.screen, a.currTetro)
	case HardDrop:
		a.playfield.ClearTetro(a.screen, a.currTetro)
		for a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveVert(1)
		}
		a.currTetro.MoveVert(-1)
		a.playfield.DrawTetro(a.screen, a.currTetro)
	}
}

func (a *App) quit() {
	a.ctxCancel()
}

// todo: impl
func (a *App) spawnNext() *Tetromino {
	return NewPinTetro()
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

type Playfield struct {
	width  int
	height int
	field  [][]byte
}

func NewPlayfield(width, height int) *Playfield {
	return &Playfield{
		width:  width,
		height: height,
	}
}

func (pf *Playfield) CanPlace(tetro *Tetromino) bool {
	for _, p := range tetro.Points {
		if p.x < OffsetLeft || p.x > OffsetLeft+20 {
			return false
		}
		if p.y >= pf.height+1 {
			return false
		}
		symbol := pf.field[p.y][p.x]
		if symbol == '[' || symbol == ']' {
			return false
		}
	}
	return true
}

func (pf *Playfield) Init() {
	pf.field = append(pf.field, bytes.Repeat([]byte{' '}, pf.width*2+4)) // invisible row

	for i := 0; i < pf.height; i++ {
		row := []byte{'<', '!'}
		row = append(row, bytes.Repeat([]byte{' ', '.'}, pf.width)...)
		row = append(row, '!', '>')
		pf.field = append(pf.field, row)
	}

	row := []byte{'<', '!'}
	row = append(row, bytes.Repeat([]byte{'=', '='}, pf.width)...)
	row = append(row, '!', '>')
	pf.field = append(pf.field, row)

	row = []byte{'<', '!'}
	row = append(row, bytes.Repeat([]byte{'\\', '/'}, pf.width)...)
	row = append(row, '!', '>')
	pf.field = append(pf.field, row)
}

func (pf *Playfield) Render(printer *TerminalScreen) {
	log("playfield: render")
	for i := 0; i < len(pf.field); i++ {
		printer.Printf("%s\n", pf.field[i])
	}
}

func (pf *Playfield) IsLanded(tetro *Tetromino) bool {
	for _, p := range tetro.Points {
		nextPoint := pf.field[p.y+1][p.x]
		if p.y == 20 || nextPoint == '[' || nextPoint == ']' {
			return true
		}
	}
	return false
}

func (pf *Playfield) LockDown(tetro *Tetromino) {
	for _, p := range tetro.Points {
		pf.field[p.y][p.x] = p.symbol
	}
}

func (pf *Playfield) ClearTetro(screen *TerminalScreen, tetro *Tetromino) {
	for _, p := range tetro.Points {
		screen.SetCursor(p.y+1, p.x+1)
		screen.Printf("%c", pf.field[p.y][p.x])
	}
}

func (pf *Playfield) DrawTetro(screen *TerminalScreen, tetro *Tetromino) {
	for _, p := range tetro.Points {
		screen.SetCursor(p.y+1, p.x+1)
		screen.Printf("%c", p.symbol)
	}
}

type Command int

const (
	Left Command = iota
	Right
	Rotate
	HardDrop
	Quit
)

var cmdNames = map[Command]string{
	Left:     "left",
	Right:    "right",
	Rotate:   "rotate",
	HardDrop: "hard-drop",
	Quit:     "quit",
}

func (c Command) String() string {
	return cmdNames[c]
}

func commandReader(ctx context.Context, stdin io.Reader) (<-chan Command, <-chan error) {
	cmds, errc := make(chan Command), make(chan error, 1)
	buf := make([]byte, 3)

	cmdMap := map[string]Command{
		"\033[A": Rotate,
		"\033[C": Right,
		"\033[D": Left,
		" ":      HardDrop,
		"q":      Quit,
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

type Point struct {
	x, y   int
	symbol byte
}

type Dir struct {
	x, y int
}
type Rule struct {
	Dirs [4]Dir
}

type Tetromino struct {
	rotationPos   int
	rotationRules [4]Rule
	Points        [8]Point
}

func NewPinTetro() *Tetromino {
	return &Tetromino{
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
		rotationRules: [4]Rule{
			{Dirs: [4]Dir{{1, 1}, {1, -1}, {0, 0}, {-1, 1}}},
			{Dirs: [4]Dir{{-1, 1}, {1, 1}, {0, 0}, {-1, -1}}},
			{Dirs: [4]Dir{{-1, -1}, {-1, 1}, {0, 0}, {1, -1}}},
			{Dirs: [4]Dir{{1, -1}, {-1, -1}, {0, 0}, {1, 1}}},
		},
	}
}

func (t *Tetromino) Rotate() {
	rule := t.rotationRules[t.rotationPos]

	for i := 0; i < len(t.Points); i += 2 {
		t.Points[i].x += rule.Dirs[i/2].x * 2
		t.Points[i+1].x += rule.Dirs[i/2].x * 2
		t.Points[i].y += rule.Dirs[i/2].y
		t.Points[i+1].y += rule.Dirs[i/2].y
	}

	t.rotationPos = (t.rotationPos + 1) % 4
}

func (t *Tetromino) MoveVert(dir int) {
	for i := range 8 {
		t.Points[i].y += dir
	}
}

func (t *Tetromino) MoveHorizontaly(dir int) {
	for i := range 8 {
		t.Points[i].x += dir * 2
	}
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

type TerminalScreen struct {
	stdout io.Writer
}

func NewTerminalScreen(stdout io.Writer) *TerminalScreen {
	return &TerminalScreen{stdout: stdout}
}

func (t *TerminalScreen) Print(s string) {
	fmt.Fprint(t.stdout, s)
}

func (t *TerminalScreen) Clearscreen() {
	fmt.Fprint(t.stdout, "\033[H\033[2J")
}

// SetPos send escape sequence to the stdout.
// The line and column starts from 1 (not from 0).ikn
func (t *TerminalScreen) SetCursor(line, column int) {
	fmt.Fprintf(t.stdout, "\033[%d;%dH", line, column)
}

func (t *TerminalScreen) Printf(format string, a ...any) {
	fmt.Fprintf(t.stdout, format, a...)
}
