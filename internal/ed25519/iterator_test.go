package ed25519_test

import (
	"testing"

	"github.com/innix/shrek/internal/ed25519"
)

func BenchmarkKeyIterator_PublicKeyAndNext(b *testing.B) {
	it, err := ed25519.NewKeyIterator(nil)
	if err != nil {
		b.Fatalf("could not create key iterator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = it.PublicKey()
		if !it.Next() {
			b.Fatal("benchmark ran so fast it searched the entire address space, whew")
		}
	}
}
