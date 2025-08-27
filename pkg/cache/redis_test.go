package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_Integration(t *testing.T) {
	// Skip if Redis is not available
	redisAddr := "localhost:6379"
	cache, err := NewRedisCache(redisAddr, "", 0)
	if err != nil {
		t.Skipf("Redis not available at %s: %v", redisAddr, err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("cache response and retrieve", func(t *testing.T) {
		// Create a test HTTP response
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)
		resp.WriteString("Hello, World!")

		httpResp := resp.Result()
		defer httpResp.Body.Close()

		// Cache the response
		ttl := 1 * time.Minute
		err := cache.CacheResponse(ctx, "https://example.com", httpResp, ttl)
		require.NoError(t, err)

		// Retrieve from cache
		cached, err := cache.GetCachedResponse(ctx, "https://example.com")
		require.NoError(t, err)
		require.NotNil(t, cached)

		assert.Equal(t, http.StatusOK, cached.StatusCode)
		assert.Equal(t, "Hello, World!", string(cached.Body))
		assert.True(t, cached.ExpiresAt.After(time.Now()))
	})

	t.Run("rate limiting", func(t *testing.T) {
		url := "https://api.example.com"
		limit := 3
		window := 1 * time.Minute

		// Check rate limit multiple times
		for i := 0; i < limit; i++ {
			limited, err := cache.CheckRateLimit(ctx, url, limit, window)
			require.NoError(t, err)
			assert.False(t, limited)
		}

		// Next request should be rate limited
		limited, err := cache.CheckRateLimit(ctx, url, limit, window)
		require.NoError(t, err)
		assert.True(t, limited)
	})

	t.Run("cache statistics", func(t *testing.T) {
		url := "https://stats.example.com"

		// Create and cache a response
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)
		resp.WriteString("Stats data")

		httpResp := resp.Result()
		defer httpResp.Body.Close()

		err := cache.CacheResponse(ctx, url, httpResp, 1*time.Minute)
		require.NoError(t, err)

		// Get cached response to increment stats
		_, err = cache.GetCachedResponse(ctx, url)
		require.NoError(t, err)

		// Get statistics
		stats, err := cache.GetCacheStats(ctx, url)
		require.NoError(t, err)

		assert.Equal(t, url, stats["url"])
		assert.Equal(t, http.StatusOK, stats["status_code"])
		assert.Equal(t, 11, stats["body_size"]) // "Stats data" length
		assert.Equal(t, 0, stats["headers_count"])
	})

	t.Run("clear cache", func(t *testing.T) {
		// Cache a response
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)
		resp.WriteString("Clear me")

		httpResp := resp.Result()
		defer httpResp.Body.Close()

		err := cache.CacheResponse(ctx, "https://clear.example.com", httpResp, 1*time.Minute)
		require.NoError(t, err)

		// Verify it's cached
		cached, err := cache.GetCachedResponse(ctx, "https://clear.example.com")
		require.NoError(t, err)
		require.NotNil(t, cached)

		// Clear cache
		err = cache.ClearCache(ctx)
		require.NoError(t, err)

		// Verify it's cleared
		cached, err = cache.GetCachedResponse(ctx, "https://clear.example.com")
		require.NoError(t, err)
		assert.Nil(t, cached)
	})
}

func TestRedisCache_KeyGeneration(t *testing.T) {
	cache := &RedisCache{}

	t.Run("cache key generation", func(t *testing.T) {
		key1 := cache.generateCacheKey("https://example.com")
		key2 := cache.generateCacheKey("https://example.com")
		key3 := cache.generateCacheKey("https://different.com")

		assert.Equal(t, key1, key2)
		assert.NotEqual(t, key1, key3)
		assert.Contains(t, key1, "cache:")
	})

	t.Run("rate limit key generation", func(t *testing.T) {
		key1 := cache.generateRateLimitKey("https://api.example.com")
		key2 := cache.generateRateLimitKey("https://api.example.com")
		key3 := cache.generateRateLimitKey("https://different-api.com")

		assert.Equal(t, key1, key2)
		assert.NotEqual(t, key1, key3)
		assert.Contains(t, key1, "ratelimit:")
	})
}
