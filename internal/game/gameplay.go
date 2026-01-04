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
	OnPreMoveTetromino  func(t *Tetromino)
	OnPostMoveTetromino func(t *Tetromino)
	OnLineChanged       func(i int)
	OnGameover          func()
	playfield           *Playfield
	currTetro           *Tetromino
}

func NewGameplay() *Gameplay {
	gp := &Gameplay{
		playfield: NewPlayfield(10, 20),
	}
	gp.currTetro = gp.nextTetro()
	return gp
}

func (g *Gameplay) Update() {
	if g.playfield.IsLanded(g.currTetro) {
		g.playfield.LockDown(g.currTetro)
		g.playfield.RemoveCompletedLines(g.OnLineChanged)
		g.currTetro = g.nextTetro()
		if !g.playfield.CanPlace(g.currTetro) {
			g.OnGameover()
		}
	}

	g.OnPreMoveTetromino(g.currTetro)
	g.currTetro.MoveVert(1)
	g.OnPostMoveTetromino(g.currTetro)
}

func (g *Gameplay) HandleCommand(cmd Command) {
	g.OnPreMoveTetromino(g.currTetro)

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

	g.OnPostMoveTetromino(g.currTetro)
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
