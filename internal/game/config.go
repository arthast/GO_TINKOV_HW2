package game

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// SpawnRule задаёт значение новой плитки и точную вероятность
type SpawnRule struct {
	Cashback    Tile
	Probability *big.Rat
}

// Reward связывает значение плитки с уровнем награды
type Reward struct {
	Cashback Tile
	Tier     string
}

// Config содержит правила кампании
type Config struct {
	MovesLimit           int
	SpawnRules           []SpawnRule
	MaxRewardCashback    Tile
	MaxRewardAllowedRate *big.Rat
	Rewards              []Reward
}

// LoadConfig читает JSON и проверяет правила до начала игры
func LoadConfig(reader io.Reader) (Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var raw struct {
		MovesLimit int `json:"movesLimit"`
		SpawnRules []struct {
			Cashback    string `json:"cashback"`
			Probability string `json:"probability"`
		} `json:"spawnRules"`
		MaxRewardCashback    string `json:"maxRewardCashback"`
		MaxRewardAllowedRate string `json:"maxRewardAllowedRate"`
		Rewards              []struct {
			Cashback string `json:"cashback"`
			Tier     string `json:"tier"`
		} `json:"rewards"`
	}
	if err = json.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	config := Config{MovesLimit: raw.MovesLimit}
	config.MaxRewardCashback, err = parseCashback(raw.MaxRewardCashback, "maxRewardCashback")
	if err != nil {
		return Config{}, err
	}
	config.MaxRewardAllowedRate, err = parseRate(raw.MaxRewardAllowedRate, "maxRewardAllowedRate")
	if err != nil {
		return Config{}, err
	}
	for i, rule := range raw.SpawnRules {
		cashback, parseErr := parseCashback(rule.Cashback, fmt.Sprintf("spawnRules[%d].cashback", i))
		if parseErr != nil {
			return Config{}, parseErr
		}
		probability, parseErr := parseRate(rule.Probability, fmt.Sprintf("spawnRules[%d].probability", i))
		if parseErr != nil {
			return Config{}, parseErr
		}
		config.SpawnRules = append(config.SpawnRules, SpawnRule{Cashback: cashback, Probability: probability})
	}
	for i, reward := range raw.Rewards {
		cashback, parseErr := parseCashback(reward.Cashback, fmt.Sprintf("rewards[%d].cashback", i))
		if parseErr != nil {
			return Config{}, parseErr
		}
		config.Rewards = append(config.Rewards, Reward{Cashback: cashback, Tier: reward.Tier})
	}
	if err = config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

// Validate проверяет обязательные параметры и полную таблицу наград
func (config Config) Validate() error {
	if config.MovesLimit <= 0 {
		return fmt.Errorf("movesLimit must be positive")
	}
	if len(config.SpawnRules) == 0 {
		return fmt.Errorf("spawnRules must not be empty")
	}
	one := big.NewRat(1, 1)
	sum := new(big.Rat)
	var maxSpawn Tile
	for i, rule := range config.SpawnRules {
		if !validCashback(rule.Cashback) {
			return fmt.Errorf("spawnRules[%d].cashback must be a positive power of two", i)
		}
		if rule.Probability == nil || rule.Probability.Sign() <= 0 || rule.Probability.Cmp(one) > 0 {
			return fmt.Errorf("spawnRules[%d].probability must be in (0, 1]", i)
		}
		sum.Add(sum, rule.Probability)
		maxSpawn = max(maxSpawn, rule.Cashback)
	}
	if sum.Cmp(one) != 0 {
		return fmt.Errorf("spawnRules probabilities must sum to 1, got %s", sum)
	}
	// Появляется не более movesLimit+2 плиток, каждый ход добавляет к счёту не больше суммы их значений
	limit := big.NewInt(int64(config.MovesLimit))
	bound := new(big.Int).Add(limit, big.NewInt(2))
	bound.Mul(bound, big.NewInt(int64(maxSpawn))).Mul(bound, limit)
	if bound.Cmp(big.NewInt(math.MaxInt)) > 0 {
		return fmt.Errorf("movesLimit and spawn cashback may overflow tile or score")
	}
	if !validCashback(config.MaxRewardCashback) {
		return fmt.Errorf("maxRewardCashback must be a positive power of two")
	}
	if config.MaxRewardAllowedRate == nil || config.MaxRewardAllowedRate.Sign() < 0 || config.MaxRewardAllowedRate.Cmp(one) > 0 {
		return fmt.Errorf("maxRewardAllowedRate must be in [0, 1]")
	}

	seen := make(map[Tile]bool)
	for i, reward := range config.Rewards {
		if !validCashback(reward.Cashback) || reward.Cashback > config.MaxRewardCashback {
			return fmt.Errorf("rewards[%d].cashback must be a positive power of two up to maxRewardCashback", i)
		}
		if strings.TrimSpace(reward.Tier) == "" {
			return fmt.Errorf("rewards[%d].tier must not be empty", i)
		}
		if seen[reward.Cashback] {
			return fmt.Errorf("duplicate reward for cashback %d", reward.Cashback)
		}
		seen[reward.Cashback] = true
	}
	for cashback := Tile(1); ; cashback *= 2 {
		if !seen[cashback] {
			return fmt.Errorf("missing reward for cashback %d", cashback)
		}
		if cashback == config.MaxRewardCashback {
			break
		}
	}
	return nil
}

func parseCashback(value, field string) (Tile, error) {
	cashback, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", field, err)
	}
	return Tile(cashback), nil
}

func parseRate(value, field string) (*big.Rat, error) {
	rate, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, fmt.Errorf("%s must be a rational number", field)
	}
	return rate, nil
}

func validCashback(cashback Tile) bool {
	return cashback > 0 && cashback&(cashback-1) == 0
}
