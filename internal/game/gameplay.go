package game

type Command int

const (
	MoveLeft Command = iota
	MoveRight
	Rotate
	HardDrop
)

var cmdNames = map[Command]string{
	MoveLeft:  "move-left",
	MoveRight: "move-right",
	Rotate:    "rotate",
	HardDrop:  "hard-drop",
}

func (c Command) String() string {
	return cmdNames[c]
}

type Gameplay struct {
	playfield       *Playfield
	playfieldBefore *Playfield
	currTetro       *Tetromino
}

func NewGameplay() *Gameplay {
	gp := &Gameplay{
		playfield:       NewPlayfield(10, 20),
		playfieldBefore: NewPlayfield(10, 20),
	}
	gp.currTetro = gp.nextTetro()
	return gp
}

func (g *Gameplay) Update() []Event {
	events := []Event{}
	if g.playfield.IsLanded(g.currTetro) {
		g.playfield.LockDown(g.currTetro)

		completed := g.playfield.CompletedLines()
		if len(completed) > 0 {
			g.playfield.CopyTo(g.playfieldBefore)
			g.playfield.RemoveCompletedLines(completed)

			events = append(events, LinesClearedEvent{
				Cleared: map_(completed, func(l int) int { return l - 1 }),
				Before:  g.playfieldBefore,
				After:   g.playfield,
			})
		}

		g.currTetro = g.nextTetro()

		if !g.playfield.CanPlace(g.currTetro) {
			events = append(events, GameOverEvent{})
		}
	}

	g.currTetro.MoveVert(1)

	return events
}

func (g *Gameplay) HandleCommand(cmd Command) {
	switch cmd {
	case Rotate:
		cand := g.currTetro.Clone()
		cand.Rotate()
		if g.playfield.CanPlace(cand) {
			g.currTetro = cand
		}
	case MoveLeft:
		g.currTetro.MoveHoriz(-1)
		if !g.playfield.CanPlace(g.currTetro) {
			g.currTetro.MoveHoriz(1)
		}
	case MoveRight:
		g.currTetro.MoveHoriz(1)
		if !g.playfield.CanPlace(g.currTetro) {
			g.currTetro.MoveHoriz(-1)
		}
	case HardDrop:
		for g.playfield.CanPlace(g.currTetro) {
			g.currTetro.MoveVert(1)
		}
		g.currTetro.MoveVert(-1)
	}
}

func (g *Gameplay) CurrentTetromino() *Tetromino {
	return g.currTetro
}

func (g *Gameplay) Field() *Playfield {
	return g.playfield
}

func (g *Gameplay) nextTetro() *Tetromino {
	return NewPinTetro()
}

type Event interface {
	IsEvent()
}

type LinesClearedEvent struct {
	Cleared []int
	Before  *Playfield
	After   *Playfield
}

func (e LinesClearedEvent) IsEvent() {}

type GameOverEvent struct {
}

func (e GameOverEvent) IsEvent() {}

func map_[T any, R any](in []T, fn func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}
