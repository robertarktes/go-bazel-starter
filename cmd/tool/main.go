package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/robertarktes/go-bazel-starter/pkg/httpx"
	"github.com/robertarktes/go-bazel-starter/pkg/retry"
)

func main() {
	url := flag.String("url", "https://example.com", "URL to fetch")
	maxRetries := flag.Int("retries", 3, "Maximum retries")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	start := time.Now()
	resp, err := client.Get(ctx, *url)
	latency := time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d, Latency: %v\n", resp.StatusCode, latency)
}
