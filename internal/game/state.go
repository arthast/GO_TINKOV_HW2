// Package game implements the campaign configuration and Cashback Merge rules.
package game

import "fmt"

// Tile - целочисленный процент кешбэка; ноль обозначает пустую ячейку
type Tile int

// Board - игровое поле размером четыре на четыре
type Board [4][4]Tile

// State содержит поле и счётчики одной игры
type State struct {
	Board      Board
	MovesLimit int
	MovesUsed  int
	Score      int
	Finished   bool
}

// String возвращает поле и счётчики в виде текста
func (s State) String() string {
	var output []byte
	for _, row := range s.Board {
		for column, tile := range row {
			if column > 0 {
				output = append(output, ' ')
			}
			if tile == 0 {
				output = append(output, '.')
			} else {
				output = fmt.Appendf(output, "%d%%", tile)
			}
		}
		output = append(output, '\n')
	}
	output = fmt.Appendf(output, "Moves: %d/%d | Score: %d\n", s.MovesUsed, s.MovesLimit, s.Score)
	return string(output)
}
