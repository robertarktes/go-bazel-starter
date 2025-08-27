package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/robertarktes/go-bazel-starter/pkg/cache"
	"github.com/robertarktes/go-bazel-starter/pkg/httpx"
	"github.com/robertarktes/go-bazel-starter/pkg/retry"
)

func main() {
	url := flag.String("url", "https://example.com", "URL to fetch")
	maxRetries := flag.Int("retries", 3, "Maximum retries")
	useCache := flag.Bool("cache", false, "Use Redis cache")
	redisAddr := flag.String("redis", "localhost:6379", "Redis server address")
	rateLimit := flag.Int("rate-limit", 10, "Rate limit per minute")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var redisCache *cache.RedisCache
	if *useCache {
		var err error
		redisCache, err = cache.NewRedisCache(*redisAddr, "", 0)
		if err != nil {
			fmt.Printf("Warning: Failed to connect to Redis at %s: %v\n", *redisAddr, err)
			fmt.Println("Continuing without cache...")
		} else {
			defer redisCache.Close()
			fmt.Printf("Connected to Redis at %s\n", *redisAddr)
		}
	}

	client := httpx.NewClient(
		httpx.WithTimeout(5*time.Second),
		httpx.WithRetries(*maxRetries, retry.WithExponentialBackoff(500*time.Millisecond, 2.0)),
		httpx.WithRequestHook(func(req *http.Request) {
			fmt.Printf("Request: %s %s\n", req.Method, req.URL)
		}),
		httpx.WithResponseHook(func(resp *http.Response, latency time.Duration) {
			fmt.Printf("Response: %d (latency: %v)\n", resp.StatusCode, latency)
		}),
	)

	if redisCache != nil {
		limited, err := redisCache.CheckRateLimit(ctx, *url, *rateLimit, time.Minute)
		if err != nil {
			fmt.Printf("Warning: Rate limit check failed: %v\n", err)
		} else if limited {
			fmt.Printf("Rate limited: %s (limit: %d requests per minute)\n", *url, *rateLimit)
			os.Exit(1)
		}
	}

	if redisCache != nil {
		if cached, err := redisCache.GetCachedResponse(ctx, *url); err == nil && cached != nil {
			fmt.Printf("Cache HIT for %s\n", *url)
			fmt.Printf("Cached response: %d (body size: %d bytes)\n",
				cached.StatusCode, len(cached.Body))
			fmt.Printf("Cached at: %s, expires at: %s\n",
				cached.CachedAt.Format(time.RFC3339),
				cached.ExpiresAt.Format(time.RFC3339))

			if stats, err := redisCache.GetCacheStats(ctx, *url); err == nil {
				fmt.Printf("Cache stats: %+v\n", stats)
			}
			return
		} else if err != nil {
			fmt.Printf("Cache check failed: %v\n", err)
		} else {
			fmt.Printf("Cache MISS for %s\n", *url)
		}
	}

	start := time.Now()
	resp, err := client.Get(ctx, *url)
	latency := time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d, Latency: %v\n", resp.StatusCode, latency)

	if redisCache != nil {
		ttl := 5 * time.Minute // Cache for 5 minutes
		if err := redisCache.CacheResponse(ctx, *url, resp, ttl); err != nil {
			fmt.Printf("Warning: Failed to cache response: %v\n", err)
		} else {
			fmt.Printf("Response cached with TTL: %v\n", ttl)
		}
	}
}
