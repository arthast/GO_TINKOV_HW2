package game

import (
	"math/big"
	"math/rand/v2"
)

// Game хранит состояние игры и генератор случайных плиток с заданным seed
type Game struct {
	State      State
	random     *rand.Rand
	tiles      []Tile
	thresholds []big.Int
	total      *big.Int
}

// New проверяет настройки кампании и создаёт две плитки в разных ячейках
func New(cfg Config, seed uint64) (*Game, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	g := &Game{
		State:      State{MovesLimit: cfg.MovesLimit},
		random:     rand.New(rand.NewPCG(seed, 0)),
		tiles:      make([]Tile, len(cfg.SpawnRules)),
		thresholds: make([]big.Int, len(cfg.SpawnRules)),
		total:      big.NewInt(1),
	}
	for _, rule := range cfg.SpawnRules {
		denominator := rule.Probability.Denom()
		divisor := new(big.Int).GCD(nil, nil, g.total, denominator)
		g.total.Quo(g.total, divisor).Mul(g.total, denominator)
	}
	var cumulative big.Int
	for i, rule := range cfg.SpawnRules {
		weight := new(big.Int).Quo(g.total, rule.Probability.Denom())
		weight.Mul(weight, rule.Probability.Num())
		cumulative.Add(&cumulative, weight)
		g.tiles[i] = rule.Cashback
		g.thresholds[i].Set(&cumulative)
	}
	g.spawn()
	g.spawn()
	return g, nil
}

func (g *Game) spawn() {
	var free [16][2]int
	count := 0
	for row, cells := range g.State.Board {
		for column, tile := range cells {
			if tile == 0 {
				free[count] = [2]int{row, column}
				count++
			}
		}
	}
	cell := free[g.random.IntN(count)]
	g.State.Board[cell[0]][cell[1]] = g.tileFor(randomBelow(g.total, g.random))
}

func (g *Game) tileFor(ticket *big.Int) Tile {
	for i := range g.thresholds {
		if ticket.Cmp(&g.thresholds[i]) < 0 {
			return g.tiles[i]
		}
	}
	return g.tiles[len(g.tiles)-1]
}

func randomBelow(limit *big.Int, source *rand.Rand) *big.Int {
	bits := limit.BitLen()
	bytes := make([]byte, (bits+7)/8)
	var ticket big.Int
	for {
		for i := range bytes {
			bytes[i] = byte(source.Uint64() & 0xff)
		}
		bytes[0] &= byte(0xff >> (len(bytes)*8 - bits))
		ticket.SetBytes(bytes)
		if ticket.Cmp(limit) < 0 {
			return &ticket
		}
	}
}
