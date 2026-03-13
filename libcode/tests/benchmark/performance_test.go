// Package benchmark provides performance benchmarks for libcode
package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gemone/libcode/internal/ai"
	"github.com/gemone/libcode/internal/ai/provider"
	"github.com/gemone/libcode/internal/logger"
	"github.com/gemone/libcode/internal/storage"
	"github.com/stretchr/testify/require"
)

// BenchmarkAIStreaming benchmarks AI streaming performance (TTFB)
func BenchmarkAIStreaming(b *testing.B) {
	if os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("ANTHROPIC_API_KEY") == "" {
		b.Skip("No API key configured")
	}

	// Initialize logger
	_ = logger.Init("WARN", "")

	// Get provider
	providerRegistry := provider.DefaultRegistry()
	providers := providerRegistry.List()
	if len(providers) == 0 {
		b.Skip("No AI provider available")
	}

	var aiProvider ai.Provider
	for _, p := range providers {
		aiProvider = p
		break
	}

	ctx := context.Background()
	req := &ai.Request{
		Messages: []ai.Message{
			{Role: "user", Content: "Count to 10"},
		},
		MaxTokens: 50,
	}

	b.ResetTimer()
	b.ReportMetric(0, "ms/TTFB")

	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Measure TTFB with streaming
		chunks, err := aiProvider.Stream(ctx, req)
		if err != nil {
			b.Fatalf("Stream failed: %v", err)
		}

		firstChunk := true
		for chunk := range chunks {
			if firstChunk {
				ttfb := time.Since(start)
				b.ReportMetric(float64(ttfb.Milliseconds()), "ms/TTFB")
				firstChunk = false
			}
			if chunk.Error != nil {
				b.Fatalf("Chunk error: %v", chunk.Error)
			}
		}
	}
}

// BenchmarkDatabaseOperations benchmarks database performance
func BenchmarkDatabaseOperations(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench.db")

	db, err := storage.Open(dbPath)
	require.NoError(b, err)
	defer db.Close()

	// Benchmark session creation
	b.Run("CreateSession", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			session := &storage.Session{
				ID:        "bench-session",
				Title:     "Benchmark Session",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Model:     "gpt-4",
				Messages: []storage.Message{
					{
						ID:        "msg-1",
						Role:      "user",
						Content:   "Benchmark test message",
						Timestamp: time.Now(),
					},
				},
			}
			_ = db.CreateSession(session)
		}
	})

	// Benchmark session retrieval
	b.Run("GetSession", func(b *testing.B) {
		// Create a session first
		session := &storage.Session{
			ID:        "get-bench-session",
			Title:     "Get Benchmark Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Model:     "gpt-4",
			Messages: []storage.Message{
				{
					ID:        "msg-1",
					Role:      "user",
					Content:   "Benchmark test message",
					Timestamp: time.Now(),
				},
			},
		}
		_ = db.CreateSession(session)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = db.GetSession(session.ID)
		}
	})

	// Benchmark session listing
	b.Run("ListSessions", func(b *testing.B) {
		// Create multiple sessions
		for i := 0; i < 100; i++ {
			session := &storage.Session{
				ID:        "list-bench-session",
				Title:     "List Benchmark Session",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Model:     "gpt-4",
			}
			_ = db.CreateSession(session)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = db.ListSessions(100)
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "memory.db")

	db, err := storage.Open(dbPath)
	require.NoError(b, err)
	defer db.Close()

	b.Run("LargeSession", func(b *testing.B) {
		// Create a session with many messages
		messages := make([]storage.Message, 1000)
		for i := 0; i < 1000; i++ {
			messages[i] = storage.Message{
				ID:        "msg-" + string(rune(i)),
				Role:      "user",
				Content:   "This is a test message with some content for memory benchmarking purposes.",
				Timestamp: time.Now(),
			}
		}

		session := &storage.Session{
			ID:        "memory-bench-session",
			Title:     "Memory Benchmark Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Model:     "gpt-4",
			Messages:  messages,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = db.CreateSession(session)
			_, _ = db.GetSession(session.ID)
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent operation performance
func BenchmarkConcurrentOperations(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent.db")

	db, err := storage.Open(dbPath)
	require.NoError(b, err)
	defer db.Close()

	b.Run("ConcurrentReads", func(b *testing.B) {
		// Create a test session
		session := &storage.Session{
			ID:        "concurrent-read-session",
			Title:     "Concurrent Read Test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Model:     "gpt-4",
		}
		_ = db.CreateSession(session)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = db.GetSession(session.ID)
			}
		})
	})

	b.Run("ConcurrentWrites", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				session := &storage.Session{
					ID:        "concurrent-write-session",
					Title:     "Concurrent Write Test",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Model:     "gpt-4",
				}
				_ = db.CreateSession(session)
			}
		})
	})
}
