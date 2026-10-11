package game

// Direction задаёт направление движения плиток
type Direction uint8

// Поддерживаемые направления движения
const (
	Up Direction = iota
	Down
	Left
	Right
)

// MoveResult сообщает, изменилось ли поле и сколько очков принёс ход
type MoveResult struct {
	Changed    bool
	ScoreDelta int
}

// Move сдвигает и объединяет плитки в копии поля без добавления новых плиток
func Move(board Board, direction Direction) (Board, int) {
	if direction > Right {
		return board, 0
	}
	score := 0
	for line := range 4 {
		var tiles [4]Tile
		for offset := range 4 {
			row, column := position(direction, line, offset)
			tiles[offset] = board[row][column]
		}
		merged, delta := mergeLine(tiles)
		score += delta
		for offset, tile := range merged {
			row, column := position(direction, line, offset)
			board[row][column] = tile
		}
	}
	return board, score
}

func mergeLine(line [4]Tile) ([4]Tile, int) {
	compact := make([]Tile, 0, len(line))
	for _, tile := range line {
		if tile != 0 {
			compact = append(compact, tile)
		}
	}
	var merged [4]Tile
	next, score := 0, 0
	for i := 0; i < len(compact); i++ {
		tile := compact[i]
		if i+1 < len(compact) && tile == compact[i+1] {
			tile *= 2
			score += int(tile)
			i++
		}
		merged[next] = tile
		next++
	}
	return merged, score
}

func position(direction Direction, line, offset int) (int, int) {
	switch direction {
	case Up:
		return offset, line
	case Down:
		return 3 - offset, line
	case Left:
		return line, offset
	case Right:
		return line, 3 - offset
	default:
		return line, offset
	}
}

// ApplyMove обновляет счётчики и добавляет одну плитку, только если поле изменилось
func (g *Game) ApplyMove(direction Direction) MoveResult {
	if g.State.Finished {
		return MoveResult{}
	}
	board, score := Move(g.State.Board, direction)
	if board == g.State.Board {
		return MoveResult{}
	}
	g.State.Board = board
	g.State.MovesUsed++
	g.State.Score += score
	g.spawn()
	g.State.Finished = g.State.MovesUsed >= g.State.MovesLimit || !hasMoves(g.State.Board)
	return MoveResult{Changed: true, ScoreDelta: score}
}

func hasMoves(board Board) bool {
	for _, direction := range []Direction{Up, Down, Left, Right} {
		if moved, _ := Move(board, direction); moved != board {
			return true
		}
	}
	return false
}
