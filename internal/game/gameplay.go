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
	rand      func(n int) int
	playfield *Playfield
	currTetro *Tetromino
}

func NewGameplay(rand func(n int) int) *Gameplay {
	gp := &Gameplay{
		rand:      rand,
		playfield: NewPlayfield(10, 20),
	}
	gp.currTetro = gp.nextTetro()
	return gp
}

func (g *Gameplay) Update() []Event {
	events := []Event{}
	if g.playfield.IsLanded(g.currTetro) {
		g.playfield.LockDown(g.currTetro)
		events = append(events, TetroLockedEvent{})

		completed := g.playfield.RemoveCompletedLines()
		events = append(events, LinesUpdatedEvent{
			Cleared: map_(completed, func(l int) int { return l - 1 }),
		})

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
	switch g.rand(2) { // todo: make 7 when all tetromino implemented
	case 0:
		return NewTTetro()
	case 1:
		return NewITetro()
	}
	panic("should resolve tetromino")
}

type Event interface {
	IsEvent()
}

type TetroLockedEvent struct {
}

func (e TetroLockedEvent) IsEvent() {}

type LinesUpdatedEvent struct {
	Cleared []int
}

func (e LinesUpdatedEvent) IsEvent() {}

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
