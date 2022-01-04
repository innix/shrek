package shrek

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/innix/shrek/internal/ed25519"
)

func MineOnionHostName(ctx context.Context, rand io.Reader, m Matcher) (*OnionAddress, error) {
	hostname := make([]byte, EncodedPublicKeySize)

	it, err := ed25519.NewKeyIterator(rand)
	if err != nil {
		return nil, fmt.Errorf("could not create key iterator: %w", err)
	}

	for more := true; ctx.Err() == nil; more = it.Next() {
		if !more {
			return nil, errors.New("searched entire address space and no match was found")
		}

		addr := &OnionAddress{
			PublicKey: it.PublicKey(),

			// The private key is not needed to generate the hostname. So to avoid pointless
			// computation, we wait until a match has been found first.
			SecretKey: nil,
		}

		// The approximate encoder only generates the first 51 bytes of the hostname accurately;
		// the last 5 bytes are wrong. But it is much faster, so it is used first then the exact
		// encoder is used if a match is found here.
		addr.HostNameApprox(hostname)

		// Check if approximate hostname matches.
		if !m.MatchApprox(hostname) {
			continue
		}

		// Generate full hostname, so we can check for exact match. Generating the full address
		// on every iteration is avoided because it's much slower than the approx.
		addr.HostName(hostname)

		// Check if exact hostname matches.
		if !m.Match(hostname) {
			continue
		}

		// Compute private key after a match has been found.
		sk, err := it.PrivateKey()
		if err != nil {
			return nil, fmt.Errorf("could not compute private key: %w", err)
		}
		addr.SecretKey = sk

		// Sanity check keys retrieved from iterator.
		kp := &ed25519.KeyPair{PublicKey: addr.PublicKey, PrivateKey: addr.SecretKey}
		if err := kp.Validate(); err != nil {
			return nil, fmt.Errorf("key validation failed: %w", err)
		}

		return addr, nil
	}

	return nil, ctx.Err()
}
