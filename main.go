package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/opennikish/tetris/internal/game"
	"github.com/opennikish/tetris/internal/terminal"
	"github.com/opennikish/tetris/internal/ui"
)

func main() {
	ctx := context.Background()

	term := terminal.NewTerminal(os.Stdin, os.Stdout, exec_)
	app := NewApp(
		10,
		20,
		term,
		ui.NewPlayfieldRenderer(term, 0, 0),
		NewRealTicker(500*time.Millisecond),
	)

	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

type App struct {
	width     int
	height    int
	term      *terminal.Terminal
	renderer  *ui.PlayfieldRender
	ticker    Ticker
	playfield *game.Playfield
	currTetro *game.Tetromino
	tickCount int
	ctxCancel context.CancelFunc
}

func NewApp(
	width, height int,
	term *terminal.Terminal,
	renderer *ui.PlayfieldRender,
	ticker Ticker,
) *App {
	return &App{
		width:    width,
		height:   height,
		renderer: renderer,
		term:     term,
		ticker:   ticker,
	}
}

func (a *App) Start(ctx context.Context) error {
	log("starting..")
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	a.ctxCancel = stop
	defer stop()

	if err := a.term.ConfigureTerminal(); err != nil {
		return fmt.Errorf("configure terminal: %w", err)
	}

	keys, errc := a.term.WatchKeystrokes(ctx)
	log("keystroke reader kicked off")

	a.ticker.Start()
	defer a.ticker.Stop()

	a.playfield = game.NewPlayfield(a.width, a.height)

	a.renderer.Draw(a.playfield)

	a.currTetro = a.nextTetro()

	log("start loop")
	for {
		select {
		case k := <-keys:
			a.onInput(a.cmdByKey(k))
		case <-a.ticker.Channel():
			a.onTick()
		case <-ctx.Done():
			a.term.SetCursor(a.height+4, 0)
			a.term.Print("Bye")
			log("stop loop")
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
		pfLine := make([]game.CellKind, a.width)
		a.playfield.RemoveCompletedLines(func(i int) {
			log("redraw line: %d", i)
			a.playfield.CopyLine(i, pfLine)
			a.renderer.RerenderPlayfieldLine(i, pfLine)
		})
		a.currTetro = a.nextTetro()
		if !a.playfield.CanPlace(a.currTetro) {
			log("gameover")
			a.quit() // todo: gameover
		}
	}

	a.renderer.ClearTetro(a.currTetro, a.playfield)
	a.currTetro.MoveVert(1)
	log("tetro points: %+v", a.currTetro.Points)
	a.renderer.DrawTetro(a.currTetro)
}

func (a *App) onInput(cmd Command) {
	log("cmd: %s", cmd)

	switch cmd {
	case Quit:
		a.quit()
	case Rotate:
		a.renderer.ClearTetro(a.currTetro, a.playfield)
		a.currTetro.Rotate()
		if !a.playfield.CanPlace(a.currTetro) {
			for range 3 {
				a.currTetro.Rotate()
			}
		}
		a.renderer.DrawTetro(a.currTetro)
	case Left:
		a.renderer.ClearTetro(a.currTetro, a.playfield)
		a.currTetro.MoveHoriz(-1)
		if !a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveHoriz(1)
		}
		a.renderer.DrawTetro(a.currTetro)
	case Right:
		a.renderer.ClearTetro(a.currTetro, a.playfield)
		a.currTetro.MoveHoriz(1)
		if !a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveHoriz(-1)
		}
		a.renderer.DrawTetro(a.currTetro)
	case HardDrop:
		a.renderer.ClearTetro(a.currTetro, a.playfield)
		for a.playfield.CanPlace(a.currTetro) {
			a.currTetro.MoveVert(1)
		}
		a.currTetro.MoveVert(-1)
		a.renderer.DrawTetro(a.currTetro)
	}
}

func (a *App) cmdByKey(key terminal.Key) Command {
	switch key.Kind {
	case terminal.Right:
		return Right
	case terminal.Left:
		return Left
	case terminal.Up:
		return Rotate
	case terminal.Letter:
		switch key.Char {
		case ' ':
			return HardDrop
		case 'q':
			return Quit
		}
	}

	log("unsupported key: %+v", key)
	return NoOp
}

func (a *App) quit() {
	a.ctxCancel()
}

// todo: impl
func (a *App) nextTetro() *game.Tetromino {
	return game.NewPinTetro()
}

type Command int

const (
	Left Command = iota
	Right
	Rotate
	HardDrop
	Quit
	NoOp
)

var cmdNames = map[Command]string{
	Left:     "left",
	Right:    "right",
	Rotate:   "rotate",
	HardDrop: "hard-drop",
	Quit:     "quit",
	NoOp:     "noop",
}

func (c Command) String() string {
	return cmdNames[c]
}

func log(format string, a ...any) {
	if len(a) == 0 {
		fmt.Fprint(os.Stderr, format+"\n")
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
	}
}

type Ticker interface {
	Channel() <-chan time.Time
	Start()
	Stop()
	Reset(d time.Duration)
}

type RealTicker struct {
	ticker *time.Ticker
	d      time.Duration
}

func NewRealTicker(d time.Duration) *RealTicker {
	return &RealTicker{d: d}
}

func (t *RealTicker) Channel() <-chan time.Time {
	return t.ticker.C
}

func (t *RealTicker) Start() {
	t.ticker = time.NewTicker(t.d)
}

func (t *RealTicker) Stop() {
	t.ticker.Stop()
}

func (t *RealTicker) Reset(d time.Duration) {
	t.ticker.Reset(d)
}

func exec_(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}
