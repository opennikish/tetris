package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/opennikish/tetris/internal/game"
	"github.com/opennikish/tetris/internal/terminal"
	"github.com/opennikish/tetris/internal/tui"
)

func main() {
	ctx := context.Background()

	term := terminal.NewTerminal(os.Stdin, os.Stdout, exec_)
	app := NewApp(
		game.NewGameplay(func(n int) int { return rand.IntN(n) }),
		term,
		tui.NewPlayfieldRenderer(term, 0, 0),
		NewRealTicker(500*time.Millisecond),		
	)

	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

type App struct {
	gameplay   *game.Gameplay
	term       *terminal.Terminal
	renderer   *tui.PlayfieldRenderer
	ticker     Ticker
	tickCount  int
	ctxCancel  context.CancelFunc
	fieldCache [][]game.CellKind
}

func NewApp(
	gameplay *game.Gameplay,
	term *terminal.Terminal,
	renderer *tui.PlayfieldRenderer,
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

	if err := a.term.UseRawModeNoEcho(); err != nil {
		return fmt.Errorf("configure terminal: %w", err)
	}

	keys, errc := a.term.WatchKeystrokes(ctx)
	log("keystroke reader kicked off")

	a.ticker.Start()
	defer a.ticker.Stop()

	a.renderer.Draw(a.gameplay.Field())

	a.fieldCache = a.createFieldCache(a.gameplay.Field().Height(), a.gameplay.Field().Width())

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

func (a *App) createFieldCache(h, w int) [][]game.CellKind {
	cache := make([][]game.CellKind, h)
	for i := range h {
		cache[i] = make([]game.CellKind, w)
		for j := range w {
			cache[i][j] = a.gameplay.Field().Cell(i, j)
		}
	}

	return cache
}

func (a *App) onTick() {
	log("tick: %d", a.tickCount)
	a.tickCount++

	if !a.gameplay.Field().IsHidden(a.gameplay.CurrentTetromino()) {
		a.renderer.DrawTetro(a.gameplay.CurrentTetromino(), game.CellEmpty)
	}

	events := a.gameplay.Update()

	for _, e := range events {
		switch evt := e.(type) {
		case game.LinesUpdatedEvent:
			log("line updated event")

			a.clearLines(evt.Cleared)
			log("lines cleared: %v", evt.Cleared)

			a.redrawLines()
			log("lines redrawed")
		case game.GameOverEvent:
			a.quit()
		}
	}

	a.renderer.DrawTetro(a.gameplay.CurrentTetromino(), game.CellBlock)
}

func (a *App) clearLines(lines []int) {
	w := a.gameplay.Field().Width()
	for _, i := range slices.Backward(lines) {
		empty := make([]game.CellKind, w)
		fill(empty, game.CellEmpty)
		a.renderer.RedrawPlayfieldLine(i, empty)
		a.fieldCache[i] = empty
	}
}

func (a *App) redrawLines() {
	for i := range a.gameplay.Field().Height() {
		for j := range a.gameplay.Field().Width() {
			actual := a.gameplay.Field().Cell(i, j)
			if a.fieldCache[i][j] != actual {
				a.renderer.RedrawCell(i, j, actual)
				a.fieldCache[i][j] = actual
			}
		}
	}
}

func (a *App) onInput(k terminal.Key) {
	if k.Char == 'q' {
		a.quit()
		return
	}
	if cmd, ok := a.cmdByKey(k); ok {
		a.renderer.DrawTetro(a.gameplay.CurrentTetromino(), game.CellEmpty)
		log("cmd: %s", cmd)
		a.gameplay.HandleCommand(cmd)
		a.renderer.DrawTetro(a.gameplay.CurrentTetromino(), game.CellBlock)
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

func fill[T any](xs []T, x T) {
	for i := range len(xs) {
		xs[i] = x
	}
}
