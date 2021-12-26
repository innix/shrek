package shrek

import (
	"context"
	"fmt"
	"io"
)

func MineOnionHostName(ctx context.Context, rand io.Reader, m Matcher) (*OnionAddress, error) {
	hostname := make([]byte, EncodedPublicKeySize)

	for ctx.Err() == nil {
		addr, err := GenerateOnionAddress(rand)
		if err != nil {
			return nil, fmt.Errorf("could not generate key pair: %w", err)
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

		return addr, nil
	}

	return nil, ctx.Err()
}
