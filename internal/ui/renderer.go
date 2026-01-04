package ui

import (
	"strings"

	"github.com/opennikish/tetris/internal/game"
	"github.com/opennikish/tetris/internal/terminal"
)

const BorderOffset = 2

type PlayfieldRender struct {
	term    *terminal.Terminal
	offsetX int
	offsetY int
}

func NewPlayfieldRenderer(term *terminal.Terminal, offsetX, offsetY int) *PlayfieldRender {
	return &PlayfieldRender{
		term:    term,
		offsetX: offsetX,
		offsetY: offsetY,
	}
}

func (r *PlayfieldRender) Draw(playfield *game.Playfield) {
	r.term.Clearscreen()
	r.term.SetCursor(r.offsetY+1, r.offsetX+1)

	leftBorder := "<!"
	rightBorder := "!>"

	r.term.Println(strings.Repeat(" ", playfield.Width()*2+len(leftBorder)+len(rightBorder)))

	pfLine := make([]game.CellKind, playfield.Width())
	for i := range playfield.Height() {
		r.term.MoveCursorRight(r.offsetX)

		r.term.Print(leftBorder)

		playfield.CopyLine(i+1, pfLine)
		r.DrawPlayfieldLine(pfLine)

		r.term.Println(rightBorder)
	}

	r.term.MoveCursorRight(r.offsetX)
	r.term.Print(leftBorder)
	r.term.Print(strings.Repeat("==", playfield.Width()))
	r.term.Println(rightBorder)

	r.term.MoveCursorRight(r.offsetX)
	r.term.Print(leftBorder)
	r.term.Print(strings.Repeat(`\/`, playfield.Width()))
	r.term.Println(rightBorder)
}

func (r *PlayfieldRender) DrawPlayfieldLine(line []game.CellKind) {
	for _, ck := range line {
		r.renderCell(ck)
	}
}

func (r *PlayfieldRender) RedrawPlayfieldLine(i int, line []game.CellKind) {
	r.term.SetCursor(r.offsetY+i+1, r.offsetX+BorderOffset+1)
	for _, ck := range line {
		r.renderCell(ck)
	}
}

// todo: consider keep only DrawTetro with mapper func
func (r *PlayfieldRender) ClearTetro(tetro *game.Tetromino, playfield *game.Playfield) {
	for _, p := range tetro.Points {
		r.term.SetCursor(r.offsetY+p.Y+1, r.offsetX+BorderOffset+p.X*2+1)
		r.renderCell(playfield.Cell(p.Y, p.X))
	}
}

func (r *PlayfieldRender) DrawTetro(tetro *game.Tetromino) {
	for _, p := range tetro.Points {
		r.term.SetCursor(r.offsetY+p.Y+1, r.offsetX+BorderOffset+p.X*2+1)
		r.renderCell(game.CellBlock)
	}
}

// todo: accept mapper func in struct
func (r *PlayfieldRender) renderCell(ck game.CellKind) {
	switch ck {
	case game.CellBlock:
		r.term.Printf("%c%c", '[', ']')
	case game.CellEmpty:
		r.term.Printf("%c%c", ' ', '.')
	case game.CellHidden:
		r.term.Printf("%c%c", ' ', ' ')
	default:
		r.term.Printf("%c%c", '?', '?')
	}
}
