package game

import (
	"math/big"
	"math/rand/v2"
	"strings"
	"testing"
)

func testConfig(t *testing.T) Config {
	t.Helper()
	cfg, err := LoadConfig(strings.NewReader(`{
		"movesLimit": 100,
		"spawnRules": [
			{"cashback": "1", "probability": "0.85"},
			{"cashback": "2", "probability": "0.15"}
		],
		"maxRewardCashback": "32", "maxRewardAllowedRate": "0.20",
		"rewards": [
			{"cashback": "1", "tier": "basic"},
			{"cashback": "2", "tier": "bronze"},
			{"cashback": "4", "tier": "silver"},
			{"cashback": "8", "tier": "gold"},
			{"cashback": "16", "tier": "platinum"},
			{"cashback": "32", "tier": "maximum"}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestNew(t *testing.T) {
	cfg := testConfig(t)
	first, err := New(cfg, 341124)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(cfg, 341124)
	if err != nil {
		t.Fatal(err)
	}
	if first.State != second.State {
		t.Fatal("same config and seed must give the same state")
	}
	count := 0
	for _, row := range first.State.Board {
		for _, tile := range row {
			if tile != 0 {
				count++
				if tile != 1 && tile != 2 {
					t.Fatalf("unexpected initial tile: %d", tile)
				}
			}
		}
	}
	if count != 2 || first.State.MovesLimit != cfg.MovesLimit ||
		first.State.MovesUsed != 0 || first.State.Score != 0 || first.State.Finished {
		t.Fatalf("incorrect initial state: %+v", first.State)
	}
	cfg.MovesLimit = 0
	if _, err := New(cfg, 341124); err == nil {
		t.Fatal("invalid config must be rejected before creating a game")
	}
}

func TestStateString(t *testing.T) {
	state := State{Board: Board{{1, 0, 2, 0}}, MovesLimit: 100}
	want := "1% . 2% .\n. . . .\n. . . .\n. . . .\nMoves: 0/100 | Score: 0\n"
	if got := state.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestSpawnWeightBoundaries(t *testing.T) {
	g, err := New(testConfig(t), 341124)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		ticket int64
		want   Tile
	}{{0, 1}, {16, 1}, {17, 2}, {19, 2}} {
		if got := g.tileFor(big.NewInt(tc.ticket)); got != tc.want {
			t.Errorf("ticket %d: got %d, want %d", tc.ticket, got, tc.want)
		}
	}
}

type sequenceSource struct {
	values []uint64
	index  int
}

func (s *sequenceSource) Uint64() uint64 {
	v := s.values[s.index]
	s.index++
	return v
}

func TestRandomTicketRejectsOutOfRange(t *testing.T) {
	source := &sequenceSource{values: []uint64{31, 19}}
	got := randomBelow(big.NewInt(20), rand.New(source))
	if got.Int64() != 19 || source.index != 2 {
		t.Fatalf("ticket = %s, draws = %d; want 19 after two draws", got, source.index)
	}
}
