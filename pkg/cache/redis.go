package cache

import (
	"encoding/json"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/redis"
)

// RedisCache provides Redis-backed persistent caching with TTL.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new RedisCache instance with the given client.
func NewRedisCache(client *redis.Client, prefix ...string) *RedisCache {
	p := "larapak:cache:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &RedisCache{
		client: client,
		prefix: p,
	}
}

// NewRedisCacheFromEnv initializes RedisCache using environment variables.
func NewRedisCacheFromEnv() *RedisCache {
	return NewRedisCache(redis.NewFromEnv())
}

func (r *RedisCache) formatKey(key string) string {
	return r.prefix + key
}

// Set stores a key-value pair in Redis with an optional TTL duration.
func (r *RedisCache) Set(key string, value any, duration time.Duration) {
	if r.client == nil {
		return
	}
	formattedKey := r.formatKey(key)

	var payload string
	switch v := value.(type) {
	case string:
		payload = v
	case []byte:
		payload = string(v)
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return
		}
		payload = string(bytes)
	}

	_ = r.client.Set(formattedKey, payload, duration)
}

// Get retrieves a key from Redis. Returns false if missing or expired.
func (r *RedisCache) Get(key string) (any, bool) {
	if r.client == nil {
		return nil, false
	}
	val, err := r.client.Get(r.formatKey(key))
	if err != nil {
		return nil, false
	}
	return val, true
}

// GetString retrieves a key specifically as a string.
func (r *RedisCache) GetString(key string) (string, bool) {
	if r.client == nil {
		return "", false
	}
	val, err := r.client.Get(r.formatKey(key))
	if err != nil {
		return "", false
	}
	return val, true
}

// Delete removes a key from Redis.
func (r *RedisCache) Delete(key string) {
	if r.client == nil {
		return
	}
	_, _ = r.client.Del(r.formatKey(key))
}

// Flush clears all keys in the currently selected database.
func (r *RedisCache) Flush() {
	if r.client == nil {
		return
	}
	_ = r.client.FlushDB()
}

// Client returns the underlying raw Redis client.
func (r *RedisCache) Client() *redis.Client {
	return r.client
}
