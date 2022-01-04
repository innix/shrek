package ed25519_test

import (
	"testing"

	"github.com/innix/shrek/internal/ed25519"
)

func BenchmarkKeyIterator_PublicKeyAndNext(b *testing.B) {
	it, err := ed25519.NewKeyIterator(nil)
	if err != nil {
		b.Fatalf("Could not create key iterator: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_ = it.PublicKey()
		if err != nil {
			b.Fatal(err)
		}
		it.Next()
	}
}
