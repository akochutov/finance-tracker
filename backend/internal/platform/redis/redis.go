package redis

import (
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func New(url string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	opts.DialTimeout = 1 * time.Second
	opts.ReadTimeout = 300 * time.Millisecond
	opts.WriteTimeout = 300 * time.Millisecond
	opts.MaxRetries = -1

	rdb := goredis.NewClient(opts)

	return rdb, nil
}
