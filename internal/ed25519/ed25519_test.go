package ed25519_test

import (
	"testing"

	"github.com/innix/shrek/internal/ed25519"
)

func BenchmarkGenerateNewKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ed25519.GenerateKey(nil)
		if err != nil {
			b.Fatalf("key pair generator errored unexpectedly during benchmark: %v", err)
		}
	}
}
