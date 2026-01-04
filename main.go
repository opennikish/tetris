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
		game.NewGameplay(),
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
	gameplay  *game.Gameplay
	term      *terminal.Terminal
	renderer  *ui.PlayfieldRender
	ticker    Ticker
	tickCount int
	ctxCancel context.CancelFunc
}

func NewApp(
	gameplay *game.Gameplay,
	term *terminal.Terminal,
	renderer *ui.PlayfieldRender,
	ticker Ticker,
) *App {
	return &App{
		gameplay: gameplay,
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

	setupGameplayHooks(a)

	a.renderer.Draw(a.gameplay.Field())

	log("start loop")
	for {
		select {
		case k := <-keys:
			a.onInput(k)
		case <-a.ticker.Channel():
			a.onTick()
		case <-ctx.Done():
			a.term.SetCursor(a.gameplay.Field().Height()+4, 0)
			a.term.Print("Bye")
			log("stop loop")
			return nil
		case err := <-errc:
			return fmt.Errorf("read ui commands: %w", err)
		}
	}
}

func setupGameplayHooks(a *App) {
	a.gameplay.OnPreMoveTetromino = func(t *game.Tetromino) {
		a.renderer.ClearTetro(t, a.gameplay.Field())
	}

	a.gameplay.OnPostMoveTetromino = func(t *game.Tetromino) {
		a.renderer.DrawTetro(t)
	}

	pfLine := make([]game.CellKind, a.gameplay.Field().Width())
	a.gameplay.OnLineChanged = func(i int) {
		log("redraw line: %d", i)
		a.gameplay.Field().CopyLine(i, pfLine)
		a.renderer.RedrawPlayfieldLine(i, pfLine)
	}

	a.gameplay.OnGameover = func() {
		a.quit()
	}
}

func (a *App) onTick() {
	log("tick: %d", a.tickCount)
	a.tickCount++
	a.gameplay.Update()
}

func (a *App) onInput(k terminal.Key) {
	if k.Char == 'q' {
		a.quit()
	}
	if cmd, ok := a.cmdByKey(k); ok {
		a.gameplay.HandleCommand(cmd)
	}
}

func (a *App) cmdByKey(key terminal.Key) (game.Command, bool) {
	switch key.Kind {
	case terminal.Right:
		return game.MoveRight, true
	case terminal.Left:
		return game.MoveLeft, true
	case terminal.Up:
		return game.Rotate, true
	case terminal.Letter:
		switch key.Char {
		case ' ':
			return game.HardDrop, true
		}
	}

	log("unsupported key: %+v", key)
	return 0, false
}

func (a *App) quit() {
	a.ctxCancel()
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
