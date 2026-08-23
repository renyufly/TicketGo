package order

import (
	"fmt"
	"time"
)

type InventoryStrategy string

const (
	StrategyNaive       InventoryStrategy = "naive"
	StrategyPessimistic InventoryStrategy = "pessimistic"
	StrategyAtomic      InventoryStrategy = "atomic"
	StrategyOptimistic  InventoryStrategy = "optimistic"
)

type ConcurrencyConfig struct {
	Strategy          InventoryStrategy
	NaiveDelay        time.Duration
	LockTimeout       time.Duration
	StatementTimeout  time.Duration
	OptimisticRetries int
	OptimisticBackoff time.Duration
}

func DefaultConcurrencyConfig() ConcurrencyConfig {
	return ConcurrencyConfig{
		Strategy:          StrategyAtomic,
		LockTimeout:       500 * time.Millisecond,
		StatementTimeout:  2500 * time.Millisecond,
		OptimisticRetries: 5,
		OptimisticBackoff: 2 * time.Millisecond,
	}
}

func ParseInventoryStrategy(value string) (InventoryStrategy, error) {
	strategy := InventoryStrategy(value)
	switch strategy {
	case StrategyNaive, StrategyPessimistic, StrategyAtomic, StrategyOptimistic:
		return strategy, nil
	default:
		return "", fmt.Errorf("unsupported inventory strategy %q", value)
	}
}
