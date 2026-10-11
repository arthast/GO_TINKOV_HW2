package game

import (
	"math/big"
	"strconv"
	"strings"
	"testing"
)

const configJSONFixture = `{
	"movesLimit": 100,
	"spawnRules": [
		{"cashback": "1", "probability": "0.1"},
		{"cashback": "2", "probability": "0.2"},
		{"cashback": "8", "probability": "0.7"}
	],
	"maxRewardCashback": "4",
	"maxRewardAllowedRate": "0.2",
	"rewards": [
		{"cashback": "1", "tier": "basic"},
		{"cashback": "2", "tier": "plus"},
		{"cashback": "4", "tier": "max"}
	]
}`

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig(strings.NewReader(configJSONFixture))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if config.MovesLimit != 100 || config.MaxRewardCashback != 4 {
		t.Fatalf("unexpected campaign settings: %+v", config)
	}
	if config.MaxRewardAllowedRate == nil || config.MaxRewardAllowedRate.Cmp(big.NewRat(1, 5)) != 0 {
		t.Fatalf("allowed rate: got %v, want 1/5", config.MaxRewardAllowedRate)
	}

	wantRules := []SpawnRule{
		{Cashback: 1, Probability: big.NewRat(1, 10)},
		{Cashback: 2, Probability: big.NewRat(1, 5)},
		{Cashback: 8, Probability: big.NewRat(7, 10)},
	}
	if len(config.SpawnRules) != len(wantRules) {
		t.Fatalf("spawn rules: got %d, want %d", len(config.SpawnRules), len(wantRules))
	}
	for i, rule := range config.SpawnRules {
		if rule.Cashback != wantRules[i].Cashback || rule.Probability == nil || rule.Probability.Cmp(wantRules[i].Probability) != 0 {
			t.Errorf("spawn rule %d: got %+v, want %+v", i, rule, wantRules[i])
		}
	}

	wantRewards := []Reward{{Cashback: 1, Tier: "basic"}, {Cashback: 2, Tier: "plus"}, {Cashback: 4, Tier: "max"}}
	if len(config.Rewards) != len(wantRewards) {
		t.Fatalf("rewards: got %d, want %d", len(config.Rewards), len(wantRewards))
	}
	for i, reward := range config.Rewards {
		if reward != wantRewards[i] {
			t.Errorf("reward %d: got %+v, want %+v", i, reward, wantRewards[i])
		}
	}
}

func TestLoadConfigRejectsInvalid(t *testing.T) {
	tests := []struct {
		name string
		old  string
		new  string
	}{
		{name: "malformed JSON", old: `"movesLimit": 100`, new: `"movesLimit":`},
		{name: "nonpositive moves limit", old: `"movesLimit": 100`, new: `"movesLimit": 0`},
		{name: "empty spawn rules", old: `"spawnRules": [`, new: `"spawnRules": [], "unused": [`},
		{name: "zero probability", old: `"probability": "0.1"`, new: `"probability": "0"`},
		{name: "negative probability", old: `"probability": "0.1"`, new: `"probability": "-0.1"`},
		{name: "probability above one", old: `"probability": "0.1"`, new: `"probability": "1.1"`},
		{name: "malformed probability", old: `"probability": "0.1"`, new: `"probability": "bad"`},
		{name: "probability sum", old: `"probability": "0.7"`, new: `"probability": "0.6"`},
		{name: "empty cashback", old: `"cashback": "1"`, new: `"cashback": "0"`},
		{name: "nonpower cashback", old: `"cashback": "1"`, new: `"cashback": "3"`},
		{name: "fractional cashback", old: `"cashback": "1"`, new: `"cashback": "1.5"`},
		{name: "numeric JSON cashback", old: `"cashback": "1"`, new: `"cashback": 1`},
		{name: "invalid max cashback", old: `"maxRewardCashback": "4"`, new: `"maxRewardCashback": "3"`},
		{name: "missing max cashback", old: `"maxRewardCashback": "4",`, new: ""},
		{name: "missing allowed rate", old: `"maxRewardAllowedRate": "0.2",`, new: ""},
		{name: "negative allowed rate", old: `"maxRewardAllowedRate": "0.2"`, new: `"maxRewardAllowedRate": "-0.1"`},
		{name: "allowed rate above one", old: `"maxRewardAllowedRate": "0.2"`, new: `"maxRewardAllowedRate": "1.1"`},
		{name: "malformed allowed rate", old: `"maxRewardAllowedRate": "0.2"`, new: `"maxRewardAllowedRate": "bad"`},
		{name: "incomplete rewards", old: `{"cashback": "2", "tier": "plus"},`, new: ""},
		{name: "duplicate reward", old: `{"cashback": "2", "tier": "plus"}`, new: `{"cashback": "1", "tier": "plus"}`},
		{name: "empty reward tier", old: `"tier": "basic"`, new: `"tier": ""`},
		{name: "invalid reward cashback", old: `{"cashback": "2", "tier": "plus"}`, new: `{"cashback": "3", "tier": "plus"}`},
		{name: "reward above maximum", old: `{"cashback": "4", "tier": "max"}`, new: `{"cashback": "8", "tier": "max"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.Replace(configJSONFixture, tt.old, tt.new, 1)
			if _, err := LoadConfig(strings.NewReader(input)); err == nil {
				t.Fatal("LoadConfig accepted invalid configuration")
			}
		})
	}
}

func TestLoadConfigAllowsRateEndpoints(t *testing.T) {
	for _, rate := range []string{"0", "1"} {
		t.Run(rate, func(t *testing.T) {
			input := strings.Replace(configJSONFixture, `"maxRewardAllowedRate": "0.2"`, `"maxRewardAllowedRate": "`+rate+`"`, 1)
			if _, err := LoadConfig(strings.NewReader(input)); err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
		})
	}
}

func TestValidateRejectsOverflowRisk(t *testing.T) {
	cfg := testConfig(t)
	cfg.MovesLimit = 1
	cfg.SpawnRules = []SpawnRule{{
		Cashback: Tile(1 << (strconv.IntSize - 2)), Probability: big.NewRat(1, 1),
	}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("config permitting tile and score overflow must be rejected")
	}
}
