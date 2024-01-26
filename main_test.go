package main

import (
	"testing"
)

func Benchmark_readKeysWithPipe(b *testing.B) {
	client := createRedisClient()
	setKeysToRedis(client, keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if gotFailedValues := readKeysWithPipe(client, keys); gotFailedValues > 0 {
			b.Errorf("failed to read from redis")
		}
	}
}

func Benchmark_readKeys(b *testing.B) {
	client := createRedisClient()
	setKeysToRedis(client, keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if gotFailedValues := readKeys(client, keys); gotFailedValues > 0 {
			b.Errorf("failed to read from redis")
		}
	}
}

func Benchmark_readKeysWithGoRoutines(b *testing.B) {
	client := createRedisClient()
	setKeysToRedis(client, keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if gotFailedValues := readKeysWithGoRoutines(client, keys); gotFailedValues > 0 {
			b.Errorf("failed to read from redis")
		}
	}
}
