package ed25519

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"

	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

const (
	// PublicKeySize is the size, in bytes, of public keys as used in this package.
	PublicKeySize = 32

	// PrivateKeySize is the size, in bytes, of private keys as used in this package.
	PrivateKeySize = 64

	// SeedSize is the size, in bytes, of private key seeds.
	SeedSize = 32
)

// PrivateKey is the type of Ed25519 private keys.
type PrivateKey []byte

// PublicKey is the type of Ed25519 public keys.
type PublicKey []byte

// KeyPair is a type with both Ed25519 keys.
type KeyPair struct {
	// PublicKey is the public key of the Ed25519 key pair.
	PublicKey PublicKey

	// PrivateKey is the private key of the Ed25519 key pair.
	PrivateKey PrivateKey
}

// Validate performs sanity checks to ensure that the public and private keys match.
func (kp *KeyPair) Validate() error {
	pk, err := getPublicKeyFromPrivateKey(kp.PrivateKey)
	if err != nil {
		return fmt.Errorf("could not compute public key from private key: %w", err)
	}

	if !bytes.Equal(kp.PublicKey, pk) {
		return errors.New("keys do not match")
	}

	return nil
}

func GenerateKey(rand io.Reader) (*KeyPair, error) {
	if rand == nil {
		rand = cryptorand.Reader
	}

	seed := make([]byte, SeedSize)
	if _, err := io.ReadFull(rand, seed); err != nil {
		return nil, err
	}

	sk := make([]byte, PrivateKeySize)
	newKeyFromSeed(sk, seed)

	// Private key does not contain the public key in this implementation, so we
	// need to compute it instead.
	pk, err := getPublicKeyFromPrivateKey(sk)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PublicKey:  pk,
		PrivateKey: sk,
	}, nil
}

func newKeyFromSeed(sk, seed []byte) {
	if l := len(seed); l != SeedSize {
		panic(fmt.Sprintf("bad seed length: %d", l))
	}

	digest := sha512.Sum512(seed)
	clampSecretKey(&digest)
	copy(sk, digest[:])
}

func getPublicKeyFromPrivateKey(sk []byte) ([]byte, error) {
	if l := len(sk); l != PrivateKeySize {
		panic(fmt.Errorf("bad private key length: %d", len(sk)))
	}

	sc, err := scalar.NewFromBits(sk[:scalar.ScalarSize])
	if err != nil {
		return nil, err
	}

	pk := curve.NewCompressedEdwardsY()
	pk.SetEdwardsPoint(curve.NewEdwardsPoint().MulBasepoint(curve.ED25519_BASEPOINT_TABLE, sc))

	return pk[:], nil
}

func clampSecretKey(sk *[64]byte) {
	sk[0] &= 248
	sk[31] &= 63
	sk[31] |= 64
}
