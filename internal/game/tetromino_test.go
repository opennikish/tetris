package game

import (
	"testing"
)

func TestTMoveDown(t *testing.T) {
	tetro := NewTTetro()

	tetro.MoveVert(3)
	expected := [4]Point{
		{4, 3},
		{3, 4},
		{4, 4},
		{5, 4},
	}

	eq(t, expected, tetro.Points)
}

func TestTMoveRight(t *testing.T) {
	tetro := NewTTetro()

	tetro.MoveHoriz(3)
	expected := [4]Point{
		{7, 0},
		{6, 1},
		{7, 1},
		{8, 1},
	}

	eq(t, expected, tetro.Points)
}

func TestTMoveLeft(t *testing.T) {
	tetro := NewTTetro()

	tetro.MoveHoriz(-3)
	expected := [4]Point{
		{1, 0},
		{0, 1},
		{1, 1},
		{2, 1},
	}

	eq(t, expected, tetro.Points)
}

func TestTRotate(t *testing.T) {
	tetro := NewTTetro()
	expected := [][4]Point{
		{
			{5, 1},
			{4, 0},
			{4, 1},
			{4, 2},
		},
		{
			{4, 2},
			{5, 1},
			{4, 1},
			{3, 1},
		},
		{
			{3, 1},
			{4, 2},
			{4, 1},
			{4, 0},
		},
		{
			{4, 0},
			{3, 1},
			{4, 1},
			{5, 1},
		},
	}

	for _, points := range expected {
		t.Logf("curr: %v\n", tetro.Points)
		tetro.Rotate()
		eq(t, points, tetro.Points)
	}
}

func eq[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Fatalf("expected: %v got: %v", expected, actual)
	}
}
