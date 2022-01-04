package ed25519

import (
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type keyIterator struct {
	kp      *KeyPair
	eightPt *curve.EdwardsPoint

	pt *curve.EdwardsPoint
	sc *scalar.Scalar

	counter uint64
}

// NewKeyIterator creates and initializes a new Ed25519 key iterator.
// The iterator is NOT thread safe; you must create a separate iterator for
// each worker instead of sharing a single instance.
func NewKeyIterator(rand io.Reader) (*keyIterator, error) {
	eightPt := curve.NewEdwardsPoint()
	eightPt = eightPt.MulBasepoint(curve.ED25519_BASEPOINT_TABLE, scalar.NewFromUint64(8))

	it := &keyIterator{
		eightPt: eightPt,
	}
	if _, err := it.init(rand); err != nil {
		return nil, err
	}

	return it, nil
}

func (it *keyIterator) Next() bool {
	const maxCounter = math.MaxUint64 - 8

	if it.counter > uint64(maxCounter) {
		return false
	}

	it.pt = it.pt.Add(it.pt, it.eightPt)
	it.counter += 8

	return true
}

func (it *keyIterator) PublicKey() PublicKey {
	var pk curve.CompressedEdwardsY
	pk.SetEdwardsPoint(it.pt)

	return pk[:]
}

func (it *keyIterator) PrivateKey() (PrivateKey, error) {
	sc := scalar.New().Set(it.sc)

	if it.counter > 0 {
		scalarAdd(sc, it.counter)
	}

	sk := make([]byte, PrivateKeySize)
	if err := sc.ToBytes(sk[:scalar.ScalarSize]); err != nil {
		panic(err)
	}
	copy(sk[scalar.ScalarSize:], it.kp.PrivateKey[scalar.ScalarSize:])

	// Sanity check.
	if !((sk[0] & 248) == sk[0]) || !(((sk[31] & 63) | 64) == sk[31]) {
		return nil, errors.New("sanity check on private key failed")
	}

	return sk, nil
}

func (it *keyIterator) init(rand io.Reader) (*KeyPair, error) {
	kp, err := GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	// Parse private key.
	sk, err := scalar.NewFromBits(kp.PrivateKey[:scalar.ScalarSize])
	if err != nil {
		return nil, fmt.Errorf("ed25519: could not parse scalar from private key: %w", err)
	}

	// Parse public key.
	cpt, err := curve.NewCompressedEdwardsYFromBytes(kp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("ed25519: could not parse point from public key: %w", err)
	}
	pk := curve.NewEdwardsPoint()
	if _, err := pk.SetCompressedY(cpt); err != nil {
		return nil, fmt.Errorf("ed25519: could not decompress point from public key: %w", err)
	}

	// Cache data so it can be used later.
	it.kp = kp
	it.sc = sk
	it.pt = pk

	// Reset counter.
	it.counter = 0

	return kp, nil
}
