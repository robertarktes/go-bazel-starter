package cache

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

type CacheEntry struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         []byte            `json:"body"`
	CachedAt     time.Time         `json:"cached_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	RequestCount int64             `json:"request_count"`
}

type RateLimitInfo struct {
	Count   int64         `json:"count"`
	ResetAt time.Time     `json:"reset_at"`
	Window  time.Duration `json:"window"`
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

func (rc *RedisCache) GetCachedResponse(ctx context.Context, url string) (*CacheEntry, error) {
	key := rc.generateCacheKey(url)

	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get from Redis: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	if time.Now().After(entry.ExpiresAt) {
		rc.client.Del(ctx, key)
		return nil, nil
	}

	rc.client.HIncrBy(ctx, key+":stats", "requests", 1)

	return &entry, nil
}

func (rc *RedisCache) CacheResponse(ctx context.Context, url string, resp *http.Response, ttl time.Duration) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	entry := CacheEntry{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		Body:         body,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		RequestCount: 0,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	key := rc.generateCacheKey(url)
	return rc.client.Set(ctx, key, data, ttl).Err()
}

func (rc *RedisCache) CheckRateLimit(ctx context.Context, url string, limit int, window time.Duration) (bool, error) {
	key := rc.generateRateLimitKey(url)

	count, err := rc.client.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	if count >= int64(limit) {
		return true, nil
	}

	pipe := rc.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to update rate limit: %w", err)
	}

	return false, nil
}

func (rc *RedisCache) GetCacheStats(ctx context.Context, url string) (map[string]interface{}, error) {
	key := rc.generateCacheKey(url)

	stats, err := rc.client.HGetAll(ctx, key+":stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache entry: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	result := map[string]interface{}{
		"url":           url,
		"cached_at":     entry.CachedAt,
		"expires_at":    entry.ExpiresAt,
		"status_code":   entry.StatusCode,
		"body_size":     len(entry.Body),
		"headers_count": len(entry.Headers),
	}

	for k, v := range stats {
		result[k] = v
	}

	return result, nil
}

func (rc *RedisCache) ClearCache(ctx context.Context) error {
	pattern := "cache:*"
	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := rc.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	return iter.Err()
}

func (rc *RedisCache) generateCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("cache:%s", hex.EncodeToString(hash[:]))
}

func (rc *RedisCache) generateRateLimitKey(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("ratelimit:%s", hex.EncodeToString(hash[:]))
}
