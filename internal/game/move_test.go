package game

import (
	"math/big"
	"testing"
)

func TestMoveLeft(t *testing.T) {
	for _, tc := range []struct {
		name  string
		row   [4]Tile
		want  [4]Tile
		score int
	}{
		{"pair", [4]Tile{1, 1, 0, 0}, [4]Tile{2, 0, 0, 0}, 2},
		{"three equal", [4]Tile{1, 1, 1, 0}, [4]Tile{2, 1, 0, 0}, 2},
		{"four equal", [4]Tile{1, 1, 1, 1}, [4]Tile{2, 2, 0, 0}, 4},
		{"two pairs", [4]Tile{2, 2, 4, 4}, [4]Tile{4, 8, 0, 0}, 12},
		{"no double merge", [4]Tile{1, 1, 2, 0}, [4]Tile{2, 2, 0, 0}, 2},
		{"compress before merge", [4]Tile{0, 1, 0, 1}, [4]Tile{2, 0, 0, 0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			board := Board{tc.row}
			got, score := Move(board, Left)
			if want := (Board{tc.want}); got != want || score != tc.score {
				t.Fatalf("Move() = %v, %d; want %v, %d", got, score, want, tc.score)
			}
		})
	}
}

func TestMoveDirections(t *testing.T) {
	board := Board{{0, 1, 0, 0}, {1, 1, 1, 0}, {0, 1, 0, 0}, {}}
	for _, tc := range []struct {
		name      string
		direction Direction
		want      Board
	}{
		{"left", Left, Board{{1, 0, 0, 0}, {2, 1, 0, 0}, {1, 0, 0, 0}, {}}},
		{"right", Right, Board{{0, 0, 0, 1}, {0, 0, 1, 2}, {0, 0, 0, 1}, {}}},
		{"up", Up, Board{{1, 2, 1, 0}, {0, 1, 0, 0}, {}, {}}},
		{"down", Down, Board{{}, {}, {0, 1, 0, 0}, {1, 2, 1, 0}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, score := Move(board, tc.direction)
			if got != tc.want || score != 2 {
				t.Fatalf("Move() = %v, %d; want %v, 2", got, score, tc.want)
			}
		})
	}
}

func TestApplyMoveNoChange(t *testing.T) {
	cfg := testConfig(t)
	first, err := New(cfg, 341124)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(cfg, 341124)
	if err != nil {
		t.Fatal(err)
	}
	state := State{Board: Board{{1, 2, 0, 0}}, MovesLimit: cfg.MovesLimit, MovesUsed: 3, Score: 7}
	first.State = state
	second.State = state
	result := first.ApplyMove(Left)
	if result.Changed || result.ScoreDelta != 0 || first.State != state {
		t.Fatalf("no-op result = %+v, state = %+v; want unchanged %+v", result, first.State, state)
	}
	first.ApplyMove(Right)
	second.ApplyMove(Right)
	if first.State != second.State {
		t.Fatalf("no-op consumed randomness: first = %+v, second = %+v", first.State, second.State)
	}
}

func TestApplyMoveValid(t *testing.T) {
	for _, tc := range []struct {
		name  string
		board Board
		want  Board
		score int
	}{
		{"merge", Board{{2, 2, 4, 4}}, Board{{4, 8, 0, 0}}, 12},
		{"slide", Board{{0, 2, 0, 4}}, Board{{2, 4, 0, 0}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig(t)
			cfg.SpawnRules = []SpawnRule{{Cashback: 1, Probability: big.NewRat(1, 1)}}
			g, err := New(cfg, 341124)
			if err != nil {
				t.Fatal(err)
			}
			g.State.Board = tc.board
			g.State.Score = 7
			result := g.ApplyMove(Left)
			if !result.Changed || result.ScoreDelta != tc.score ||
				g.State.Score != 7+tc.score || g.State.MovesUsed != 1 || g.State.Finished {
				t.Fatalf("result = %+v, state = %+v", result, g.State)
			}
			spawned := 0
			for row, cells := range g.State.Board {
				for column, tile := range cells {
					want := tc.want[row][column]
					if tile == want {
						continue
					}
					if want != 0 || tile != 1 {
						t.Fatalf("cell [%d][%d] = %d; want %d or one spawn in an empty cell", row, column, tile, want)
					}
					spawned++
				}
			}
			if spawned != 1 {
				t.Fatalf("spawned %d tiles; want 1", spawned)
			}
		})
	}
}

func TestApplyMoveLimit(t *testing.T) {
	cfg := testConfig(t)
	cfg.MovesLimit = 1
	g, err := New(cfg, 341124)
	if err != nil {
		t.Fatal(err)
	}
	g.State.Board = Board{{1, 1, 0, 0}}
	result := g.ApplyMove(Left)
	if !result.Changed || result.ScoreDelta != 2 || !g.State.Finished ||
		g.State.MovesUsed != 1 || g.State.Score != 2 {
		t.Fatalf("last move result = %+v, state = %+v", result, g.State)
	}
	state := g.State
	result = g.ApplyMove(Right)
	if result.Changed || result.ScoreDelta != 0 || g.State != state {
		t.Fatalf("finished game changed: result = %+v, state = %+v; want %+v", result, g.State, state)
	}
}

func TestApplyMoveFullBoardAfterSpawn(t *testing.T) {
	cfg := testConfig(t)
	cfg.SpawnRules = []SpawnRule{{Cashback: 1, Probability: big.NewRat(1, 1)}}
	for _, tc := range []struct {
		name       string
		belowSpawn Tile
		finished   bool
	}{
		{"no moves", 64, true},
		{"vertical merge remains", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, err := New(cfg, 341124)
			if err != nil {
				t.Fatal(err)
			}
			g.State.Board = Board{
				{0, 1, 2, 4},
				{8, 16, 32, tc.belowSpawn},
				{128, 256, 512, 1024},
				{2048, 4096, 8192, 16384},
			}
			want := g.State.Board
			want[0] = [4]Tile{1, 2, 4, 1}
			result := g.ApplyMove(Left)
			if !result.Changed || result.ScoreDelta != 0 || g.State.Board != want ||
				g.State.Finished != tc.finished || g.State.MovesUsed != 1 || g.State.Score != 0 {
				t.Fatalf("full board result = %+v, state = %+v; want board %v, finished %t", result, g.State, want, tc.finished)
			}
		})
	}
}

func TestApplyMoveContinuesAfter32(t *testing.T) {
	g, err := New(testConfig(t), 341124)
	if err != nil {
		t.Fatal(err)
	}
	g.State.Board = Board{{16, 16, 0, 0}}
	result := g.ApplyMove(Left)
	if !result.Changed || result.ScoreDelta != 32 || g.State.Board[0][0] != 32 ||
		g.State.Score != 32 || g.State.MovesUsed != 1 || g.State.Finished {
		t.Fatalf("32%% merge result = %+v, state = %+v", result, g.State)
	}
}
