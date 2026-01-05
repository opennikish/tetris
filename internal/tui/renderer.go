package tui

import (
	"strings"

	"github.com/opennikish/tetris/internal/game"
	"github.com/opennikish/tetris/internal/terminal"
)

const BorderOffset = 2

type PlayfieldRenderer struct {
	term    *terminal.Terminal
	offsetX int
	offsetY int
}

func NewPlayfieldRenderer(term *terminal.Terminal, offsetX, offsetY int) *PlayfieldRenderer {
	return &PlayfieldRenderer{
		term:    term,
		offsetX: offsetX,
		offsetY: offsetY,
	}
}

func (r *PlayfieldRenderer) Draw(playfield *game.Playfield) {
	r.term.Clear()
	r.term.SetCursor(r.offsetY+1, r.offsetX+1)

	leftBorder := "<!"
	rightBorder := "!>"

	r.term.Println(strings.Repeat(" ", playfield.Width()*2+len(leftBorder)+len(rightBorder))) // todo: remove it & add offsetY, now's it's only for testing

	pfLine := make([]game.CellKind, playfield.Width())
	for i := range playfield.Height() {
		r.term.MoveCursorRight(r.offsetX)

		r.term.Print(leftBorder)

		playfield.CopyLine(i, pfLine)
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

func (r *PlayfieldRenderer) DrawPlayfieldLine(line []game.CellKind) {
	for _, ck := range line {
		r.renderCell(ck)
	}
}

func (r *PlayfieldRenderer) RedrawPlayfieldLine(i int, line []game.CellKind) {
	r.term.SetCursor(r.offsetY+i+1+1, r.offsetX+BorderOffset+1) // extra +1 for line because of drawing empty line
	for _, ck := range line {
		r.renderCell(ck)
	}
}

func (r *PlayfieldRenderer) DrawTetro(tetro *game.Tetromino, ck game.CellKind) {
	for _, p := range tetro.Points {
		r.term.SetCursor(r.offsetY+p.Y+1, r.offsetX+BorderOffset+p.X*2+1)
		r.renderCell(ck)
	}
}

// todo: accept mapper func in struct
func (r *PlayfieldRenderer) renderCell(ck game.CellKind) {
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
