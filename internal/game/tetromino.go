package game

type Point struct{ X, Y int }

type vec struct {
	x, y int
}
type rotationStep struct {
	deltas [4]vec
}

type Tetromino struct {
	rotationPos   int
	rotationSteps []rotationStep
	Points        [4]Point
}

func NewTTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{3, 0},
			{4, 0},
			{5, 0},
			{4, 1},
		},
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{1, 1}, {0, 0}, {-1, -1}, {1, -1}}},
			{deltas: [4]vec{{1, -1}, {0, 0}, {-1, 1}, {-1, -1}}},
			{deltas: [4]vec{{-1, -1}, {0, 0}, {1, 1}, {-1, 1}}},
			{deltas: [4]vec{{-1, 1}, {0, 0}, {1, -1}, {1, 1}}},
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
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{2, -1}, {1, 0}, {0, 1}, {-1, 2}}},
			{deltas: [4]vec{{-2, 1}, {-1, 0}, {0, -1}, {1, -2}}},
		},
	}
}

func NewOTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{4, 0},
			{5, 0},
			{4, 1},
			{5, 1},
		},
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{0, 0}, {0, 0}, {0, 0}, {0, 0}}},
		},
	}
}

func NewSTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{4, 0},
			{5, 0},
			{3, 1},
			{4, 1},
		},
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{1, 0}, {0, 1}, {1, -2}, {0, -1}}},
			{deltas: [4]vec{{-1, 0}, {0, -1}, {-1, 2}, {0, 1}}},
		},
	}
}

func NewZTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{3, 0},
			{4, 0},
			{4, 1},
			{5, 1},
		},
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{2, -1}, {1, 0}, {0, -1}, {-1, 0}}},
			{deltas: [4]vec{{-2, 1}, {-1, 0}, {0, 1}, {1, 0}}},
		},
	}
}

func NewLTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{3, 0},
			{4, 0},
			{5, 0},
			{3, 1},
		},
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{1, 1}, {0, 0}, {-1, -1}, {2, 0}}},
			{deltas: [4]vec{{1, -1}, {0, 0}, {-1, 1}, {0, -2}}},
			{deltas: [4]vec{{-1, -1}, {0, 0}, {1, 1}, {-2, 0}}},
			{deltas: [4]vec{{-1, 1}, {0, 0}, {1, -1}, {0, 2}}},
		},
	}
}

func NewJTetro() *Tetromino {
	return &Tetromino{
		Points: [4]Point{
			{3, 0},
			{4, 0},
			{5, 0},
			{5, 1},
		},
		rotationSteps: []rotationStep{
			{deltas: [4]vec{{1, 1}, {0, 0}, {-1, -1}, {0, -2}}},
			{deltas: [4]vec{{1, -1}, {0, 0}, {-1, 1}, {-2, 0}}},
			{deltas: [4]vec{{-1, -1}, {0, 0}, {1, 1}, {0, 2}}},
			{deltas: [4]vec{{-1, 1}, {0, 0}, {1, -1}, {2, 0}}},
		},
	}
}

func (t *Tetromino) Rotate() {
	rule := t.rotationSteps[t.rotationPos]

	for i := 0; i < len(t.Points); i += 1 {
		t.Points[i].X += rule.deltas[i].x
		t.Points[i].Y += rule.deltas[i].y
	}

	t.rotationPos = (t.rotationPos + 1) % len(t.rotationSteps)
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
		rotationSteps: t.rotationSteps,
		rotationPos:   t.rotationPos,
	}
}
