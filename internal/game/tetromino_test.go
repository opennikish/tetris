package game

import (
	"strings"
	"testing"
)

func TestTMoveDown(t *testing.T) {
	tetro := NewTTetro()
	t.Logf("curr:\n%s\n", toString(tetro))

	tetro.MoveVert(1)
	expected := strings.TrimSpace(`
..........
...###....
....#.....
..........
`)

	eq(t, expected, toString(tetro))
}

func TestTMoveRight(t *testing.T) {
	tetro := NewTTetro()
	t.Logf("curr:\n%s\n", toString(tetro))

	tetro.MoveHoriz(1)
	expected := strings.TrimSpace(`
....###...
.....#....
..........
..........
`)

	eq(t, expected, toString(tetro))
}

func TestTMoveLeft(t *testing.T) {
	tetro := NewTTetro()
	t.Logf("curr:\n%s\n", toString(tetro))

	tetro.MoveHoriz(-1)
	expected := strings.TrimSpace(`
..###.....
...#......
..........
..........
`)

	eq(t, expected, toString(tetro))
}

func TestTRotate(t *testing.T) {
	tetro := NewTTetro()
	tetro.MoveVert(1)
	expected := []string{
		strings.TrimSpace(`
....#.....
....##....
....#.....
..........
`),
		strings.TrimSpace(`
....#.....
...###....
..........
..........
`),
		strings.TrimSpace(`
....#.....
...##.....
....#.....
..........
`),
		strings.TrimSpace(`
..........
...###....
....#.....
..........
`),
	}

	for _, exp := range expected {
		t.Logf("\ncurr: %v\n", tetro.Points)
		tetro.Rotate()
		eq(t, exp, toString(tetro))
	}
}

func eq[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Fatalf("\n--- expected:\n%v\n--- got:\n%v", expected, actual)
	}
}

func toString(tetro *Tetromino) string {
	field := []byte(strings.TrimSpace(`
..........
..........
..........
..........
`))

	for _, p := range tetro.Points {
		i := 11*p.Y + p.X // 11 is field width + \n
		field[i] = '#'
	}

	return string(field)
}
