package game

type Point struct{ X, Y int }

type dir struct {
	x, y int
}
type rotationRule struct {
	dirs [4]dir
}

type Tetromino struct {
	rotationPos   int
	rotationRules []rotationRule
	Points        [4]Point
}

func NewPinTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{4, 0},
			{3, 1},
			{4, 1},
			{5, 1},
		},
		rotationRules: []rotationRule{
			{dirs: [4]dir{{1, 1}, {1, -1}, {0, 0}, {-1, 1}}},
			{dirs: [4]dir{{-1, 1}, {1, 1}, {0, 0}, {-1, -1}}},
			{dirs: [4]dir{{-1, -1}, {-1, 1}, {0, 0}, {1, -1}}},
			{dirs: [4]dir{{1, -1}, {-1, -1}, {0, 0}, {1, 1}}},
		},
	}
}

func NewITetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{3, 0},
			{4, 0},
			{5, 0},
			{6, 0},
		},
		rotationRules: []rotationRule{
			{dirs: [4]dir{{2, -1}, {1, 0}, {0, 1}, {-1, 2}}},
			{dirs: [4]dir{{-2, 1}, {-1, 0}, {0, -1}, {1, -2}}},
			// {dirs: [4]dir{{2, -1}, {1, 0}, {0, 1}, {-1, 2}}},
			// {dirs: [4]dir{{-2, 1}, {-1, 0}, {0, -1}, {1, -2}}},
		},
	}
}

func (t *Tetromino) Rotate() {
	rule := t.rotationRules[t.rotationPos]

	for i := 0; i < len(t.Points); i += 1 {
		t.Points[i].X += rule.dirs[i].x
		t.Points[i].Y += rule.dirs[i].y
	}

	t.rotationPos = (t.rotationPos + 1) % len(t.rotationRules)
}

func (t *Tetromino) MoveVert(dir int) {
	for i := range len(t.Points) {
		t.Points[i].Y += dir
	}
}

func (t *Tetromino) MoveHoriz(dir int) {
	for i := range len(t.Points) {
		t.Points[i].X += dir
	}
}

func (t *Tetromino) Clone() *Tetromino {
	return &Tetromino{
		Points:        t.Points, // arrays are values
		rotationRules: t.rotationRules,
		rotationPos:   t.rotationPos,
	}
}
