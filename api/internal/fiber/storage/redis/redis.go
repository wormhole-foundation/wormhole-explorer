package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Storage interface that is implemented by storage providers
type Storage struct {
	db     redis.UniversalClient
	prefix string
}

// New creates a new redis storage
func New(config ...Config) (*Storage, error) {
	// Set default config
	cfg := configDefault(config...)

	// Create new redis universal client
	var db redis.UniversalClient

	// Parse the URL and update config values accordingly
	if cfg.URL != "" {
		options, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, err
		}

		// Update the config values with the parsed URL values
		cfg.Username = options.Username
		cfg.Password = options.Password
		cfg.Database = options.DB
		cfg.Addrs = []string{options.Addr}
	} else if len(cfg.Addrs) == 0 {
		// Fallback to Host and Port values if Addrs is empty
		cfg.Addrs = []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)}
	}

	db = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:            cfg.Addrs,
		MasterName:       cfg.MasterName,
		ClientName:       cfg.ClientName,
		SentinelUsername: cfg.SentinelUsername,
		SentinelPassword: cfg.SentinelPassword,
		DB:               cfg.Database,
		Username:         cfg.Username,
		Password:         cfg.Password,
		TLSConfig:        cfg.TLSConfig,
		PoolSize:         cfg.PoolSize,
	})

	// Test connection
	if err := db.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	// Empty collection if Clear is true
	if cfg.Reset {
		if err := db.FlushDB(context.Background()).Err(); err != nil {
			return nil, err
		}
	}

	// Create new store
	return &Storage{
		db:     db,
		prefix: cfg.Prefix,
	}, nil
}

// ...

// Get value by key
func (s *Storage) Get(key string) ([]byte, error) {
	if len(key) <= 0 {
		return nil, nil
	}
	val, err := s.db.Get(context.Background(), s.composeKey(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// Set key with value
func (s *Storage) Set(key string, val []byte, exp time.Duration) error {
	if len(key) <= 0 || len(val) <= 0 {
		return nil
	}
	return s.db.Set(context.Background(), s.composeKey(key), val, exp).Err()
}

// Delete key by key
func (s *Storage) Delete(key string) error {
	if len(key) <= 0 {
		return nil
	}
	return s.db.Del(context.Background(), s.composeKey(key)).Err()
}

// Reset all keys
func (s *Storage) Reset() error {
	return s.db.FlushDB(context.Background()).Err()
}

// Close the database
func (s *Storage) Close() error {
	return s.db.Close()
}

// Return database client
func (s *Storage) Conn() redis.UniversalClient {
	return s.db
}

func (s *Storage) composeKey(key string) string {
	if s.prefix != "" {
		return s.prefix + ":" + key
	}
	return key
}
