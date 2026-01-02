package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/opennikish/tetris/internal/game"
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
	playfield     *game.Playfield
	currTetro     *game.Tetromino
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

	a.playfield = game.NewPlayfield(a.width, a.height)

	a.render(a.playfield)

	a.currTetro = a.nextTetro()

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

func (a *App) render(playfield *game.Playfield) {
	a.screen.Clearscreen()
	a.screen.SetCursor(1, 1)

	a.screen.Print(strings.Repeat(" ", a.width*2+4))
	a.screen.Print("\n")

	leftBorder := "<!"
	rightBorder := "!>"

	pfLine := make([]game.CellKind, a.width)

	for i := range a.height {
		a.screen.Print(leftBorder)

		playfield.CopyLine(i+1, pfLine)
		a.renderPlayfieldLine(pfLine)

		a.screen.Print(rightBorder)
		a.screen.Print("\n")
	}

	a.screen.Print(leftBorder)
	a.screen.Print(strings.Repeat("==", a.width))
	a.screen.Print(rightBorder)
	a.screen.Print("\n")
	a.screen.Print(leftBorder)
	a.screen.Print(strings.Repeat("\\/", a.width))
	a.screen.Print(rightBorder)
	a.screen.Print("\n")
}

func (a *App) renderPlayfieldLine(line []game.CellKind) {
	log("line: %v", line)
	for _, ck := range line {
		a.renderCell(ck)
	}
}

func (a *App) renderCell(ck game.CellKind) {
	switch ck {
	case game.CellBlock:
		a.screen.Printf("%c%c", '[', ']')
	case game.CellEmpty:
		a.screen.Printf("%c%c", ' ', '.')
	case game.CellHidden:
		a.screen.Printf("%c%c", ' ', ' ')
	default:
		a.screen.Printf("%c%c", '?', '?')
	}
}

func (a *App) onTick() {
	log("tick: %d", a.tickCount)
	a.tickCount++

	if a.playfield.IsLanded(a.currTetro) {
		log("tetro landed")
		a.playfield.LockDown(a.currTetro)
		pfLine := make([]game.CellKind, a.width)
		a.playfield.RemoveCompletedLines(func(i int) {
			log("redraw line: %d", i)
			a.screen.SetCursor(OffsetTop+i+1, OffsetLeft+1)
			a.playfield.CopyLine(i, pfLine)
			a.renderPlayfieldLine(pfLine)
		})
		a.currTetro = a.nextTetro()
		if !a.playfield.CanPlace(a.currTetro) {
			log("gameover")
			a.quit() // todo: gameover
		}
	}

	a.clearTetro(a.currTetro, a.playfield)
	a.currTetro.MoveVert(1)
	a.drawTetro(a.currTetro)
}

func (a *App) onInput(cmd Command) {
	log("cmd: %s", cmd)

	switch cmd {
	case Quit:
		a.quit()
	case Rotate:
		a.clearTetro(a.currTetro, a.playfield)
		a.currTetro.Rotate()
		if !a.playfield.CanPlace(a.currTetro) {
			for range 3 {
				a.currTetro.Rotate()
			}
		}
		a.drawTetro(a.currTetro)
	case Left:
		a.clearTetro(a.currTetro, a.playfield)
		a.currTetro.MoveHoriz(-1)
		if !a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveHoriz(1)
		}
		a.drawTetro(a.currTetro)
	case Right:
		a.clearTetro(a.currTetro, a.playfield)
		a.currTetro.MoveHoriz(1)
		if !a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveHoriz(-1)
		}
		a.drawTetro(a.currTetro)
	case HardDrop:
		a.clearTetro(a.currTetro, a.playfield)
		for a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveVert(1)
		}
		a.currTetro.MoveVert(-1)
		a.drawTetro(a.currTetro)
	}
}

func (a *App) quit() {
	a.ctxCancel()
}

// todo: impl
func (a *App) nextTetro() *game.Tetromino {
	return game.NewPinTetro()
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

func (a *App) clearTetro(tetro *game.Tetromino, playfield *game.Playfield) {
	for _, p := range tetro.Points {
		a.screen.SetCursor(OffsetTop+p.Y+1, OffsetLeft+p.X*2+1)
		a.renderCell(playfield.Cell(p.Y, p.X))
	}
}

func (a *App) drawTetro(tetro *game.Tetromino) {
	for _, p := range tetro.Points {
		a.screen.SetCursor(OffsetTop+p.Y+1, OffsetLeft+p.X*2+1)
		a.renderCell(game.CellBlock)
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
