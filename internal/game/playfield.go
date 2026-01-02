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
	width int
	field [][]CellKind
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

	return &Playfield{
		width: width,
		field: field,
	}
}

func (pf *Playfield) CopyLine(i int, dst []CellKind) {
	copy(dst, pf.field[i])
}

func (pf *Playfield) Cell(i, j int) CellKind {
	return pf.field[i][j]
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

func (pf *Playfield) RemoveCompletedLines(onLineChanged func(i int)) int {
	completed := pf.completedLines()

	if len(completed) == 0 {
		return 0
	}

	emptyLine := make([]CellKind, pf.width)
	fill(emptyLine, CellEmpty)
	for _, k := range slices.Backward(completed) {
		copy(pf.field[k], emptyLine)
		onLineChanged(k)
	}

	updated := pf.collapseAbove(completed)
	for _, i := range updated {
		onLineChanged(i)
	}

	return len(completed)
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

func (pf *Playfield) collapseAbove(completed []int) []int {
	updated := []int{}
	step := 0

	for i := len(pf.field) - 1; i >= 1; i -= 1 { // ignore hidden line
		if slices.Contains(completed, i) {
			step++
			continue
		}

		if step == 0 || slices.Equal(pf.field[i], pf.field[i+step]) {
			continue
		}

		pf.field[i], pf.field[i+step] = pf.field[i+step], pf.field[i]
		updated = append(updated, i, i+step)
	}

	return updated
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

func fill[T any](xs []T, x T) {
	for i := range len(xs) {
		xs[i] = x
	}
}
