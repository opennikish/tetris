package game

import (
	"slices"
)

type CellKind uint8

const (
	CellHidden CellKind = iota
	CellEmpty
	CellBlock
)

type Playfield struct {
	width     int
	field     [][]CellKind
	emptyLine []CellKind
}

func NewPlayfield(width, height int) *Playfield {
	field := make([][]CellKind, height+1)

	hidden := make([]CellKind, width)
	fill(hidden, CellHidden)
	field[0] = hidden

	for i := 1; i < height+1; i++ {
		empty := make([]CellKind, width)
		fill(empty, CellEmpty)
		field[i] = empty
	}

	empty := make([]CellKind, width)
	fill(empty, CellEmpty)

	return &Playfield{
		width:     width,
		field:     field,
		emptyLine: empty,
	}
}

func (pf *Playfield) CopyLine(i int, dst []CellKind) {
	copy(dst, pf.field[i+1]) // 1-based b/c hidden line
}

func (pf *Playfield) Cell(i, j int) CellKind {
	return pf.field[i+1][j]
}

func (pf *Playfield) Height() int {
	return len(pf.field) - 1
}

func (pf *Playfield) Width() int {
	return pf.width
}

func (pf *Playfield) CanPlace(tetro *Tetromino) bool {
	for _, p := range tetro.Points {
		if p.X < 0 || p.X >= pf.width {
			return false
		}

		if p.Y >= len(pf.field) {
			return false
		}

		cell := pf.field[p.Y][p.X]
		if cell == CellBlock {
			return false
		}
	}

	return true
}

func (pf *Playfield) RemoveCompletedLines() []int {
	completed := pf.completedLines()
	for _, k := range completed {
		copy(pf.field[k], pf.emptyLine)
	}

	step := 0
	for i := len(pf.field) - 1; i >= 1; i -= 1 { // ignore hidden line
		if slices.Contains(completed, i) {
			step++
			continue
		}

		if step == 0 {
			continue
		}

		pf.field[i], pf.field[i+step] = pf.field[i+step], pf.field[i]
	}

	return completed
}

func (pf *Playfield) completedLines() []int {
	completed := make([]int, 0, 4)

	for i := 1; i < len(pf.field); i++ { // ignore hidden line
		if !slices.Contains(pf.field[i], CellEmpty) {
			completed = append(completed, i)
		}
	}

	return completed
}

func (pf *Playfield) IsLanded(tetro *Tetromino) bool {
	for _, p := range tetro.Points {
		if p.Y == len(pf.field)-1 || pf.field[p.Y+1][p.X] == CellBlock {
			return true
		}
	}
	return false
}

func (pf *Playfield) LockDown(tetro *Tetromino) {
	for _, p := range tetro.Points {
		pf.field[p.Y][p.X] = CellBlock
	}
}

func (pf *Playfield) IsHidden(tetro *Tetromino) bool {
	for _, p := range tetro.Points {
		cell := pf.field[p.Y][p.X]
		if cell == CellHidden {
			return true
		}
	}

	return false
}

func fill[T any](xs []T, x T) {
	for i := range len(xs) {
		xs[i] = x
	}
}
