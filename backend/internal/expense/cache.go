package expense

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	suggestionsKey = "fintrack:v1:expense:suggestions"
	cacheOpTimeout = 200 * time.Millisecond
)

type SuggestionCache interface {
	Get(ctx context.Context) ([]Suggestion, bool, error)
	Set(ctx context.Context, s []Suggestion) error
	Invalidate(ctx context.Context) error
}

type RedisSuggestionCache struct {
	rdb *goredis.Client
	ttl time.Duration
}

func NewRedisSuggestionCache(rdb *goredis.Client, ttl time.Duration) *RedisSuggestionCache {
	return &RedisSuggestionCache{rdb: rdb, ttl: ttl}
}

var _ SuggestionCache = (*RedisSuggestionCache)(nil)

func (c *RedisSuggestionCache) Get(ctx context.Context) ([]Suggestion, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheOpTimeout)
	defer cancel()

	data, err := c.rdb.Get(ctx, suggestionsKey).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("redis get: %w", err)
	}

	var list []Suggestion
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, false, fmt.Errorf("unmarshal suggestions: %w", err)
	}

	return list, true, nil
}

func (c *RedisSuggestionCache) Set(ctx context.Context, s []Suggestion) error {
	ctx, cancel := context.WithTimeout(ctx, cacheOpTimeout)
	defer cancel()

	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal suggestions: %w", err)
	}

	if err := c.rdb.Set(ctx, suggestionsKey, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (c *RedisSuggestionCache) Invalidate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, cacheOpTimeout)
	defer cancel()

	if err := c.rdb.Del(ctx, suggestionsKey).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}
