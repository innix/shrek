package shrek

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ed25519voi "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"golang.org/x/crypto/sha3"
)

const (
	// EncodedPublicKeyApproxSize is the size, in bytes, of the public key when encoded using the approximate encoder.
	EncodedPublicKeyApproxSize = 51

	// EncodedPublicKeySize is the size, in bytes, of the public key when encoded using the real encoder.
	EncodedPublicKeySize = 56
)

var b32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")

type OnionAddress struct {
	PublicKey ed25519.PublicKey
	SecretKey ed25519.PrivateKey
}

func (addr *OnionAddress) HostName(hostname []byte) {
	const version = 3

	if len(hostname) != EncodedPublicKeySize {
		panic(fmt.Sprintf("Hostname buffer must have len of %d", EncodedPublicKeySize))
	}

	// checksum = sha3_sum256(".onion checksum" + public_key + version)
	var checksumBuf bytes.Buffer
	checksumBuf.Write([]byte(".onion checksum"))
	checksumBuf.Write(addr.PublicKey)
	checksumBuf.Write([]byte{version})
	checksum := sha3.Sum256(checksumBuf.Bytes())

	// onion_addr = base32_encode(public_key + checksum + version)
	var onionAddrBuf bytes.Buffer
	onionAddrBuf.Write(addr.PublicKey)
	onionAddrBuf.Write(checksum[:2])
	onionAddrBuf.Write([]byte{version})

	b32.Encode(hostname, onionAddrBuf.Bytes())
}

func (addr *OnionAddress) HostNameString() string {
	hostname := make([]byte, EncodedPublicKeySize)
	addr.HostName(hostname)

	return fmt.Sprintf("%s.onion", hostname)
}

func (addr *OnionAddress) HostNameApprox(hostname []byte) {
	if len(hostname) != EncodedPublicKeySize {
		panic(fmt.Sprintf("Hostname buffer must have len of %d", EncodedPublicKeySize))
	}

	b32.Encode(hostname, addr.PublicKey)
}

func GenerateOnionAddress(rand io.Reader) (*OnionAddress, error) {
	publicKey, secretKey, err := ed25519voi.GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	return &OnionAddress{
		PublicKey: ed25519.PublicKey(publicKey),
		SecretKey: ed25519.PrivateKey(secretKey),
	}, nil
}

func GenerateOnionAddressSlow(rand io.Reader) (*OnionAddress, error) {
	publicKey, secretKey, err := ed25519.GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	return &OnionAddress{
		PublicKey: publicKey,
		SecretKey: secretKey,
	}, nil
}

func SaveOnionAddress(dir string, addr *OnionAddress) error {
	const (
		dirMode  = 0o700
		fileMode = 0o600
	)

	hostname := addr.HostNameString()
	dir = filepath.Join(dir, hostname)

	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("could not create directories: %w", err)
	}

	pk := addr.PublicKey
	pkFile := filepath.Join(dir, "hs_ed25519_public_key")
	pkData := append([]byte("== ed25519v1-public: type0 ==\x00\x00\x00"), pk...)
	if err := os.WriteFile(pkFile, pkData, fileMode); err != nil {
		return fmt.Errorf("could not save public key to file: %w", err)
	}

	sk := expandSecretKey(addr.SecretKey)
	skFile := filepath.Join(dir, "hs_ed25519_secret_key")
	skData := append([]byte("== ed25519v1-secret: type0 ==\x00\x00\x00"), sk[:]...)
	if err := os.WriteFile(skFile, skData, fileMode); err != nil {
		return fmt.Errorf("could not save secret key to file: %w", err)
	}

	hnFile := filepath.Join(dir, "hostname")
	hnData := []byte(hostname)
	if err := os.WriteFile(hnFile, hnData, fileMode); err != nil {
		return fmt.Errorf("could not save onion hostname to file: %w", err)
	}

	return nil
}

func expandSecretKey(secretKey ed25519.PrivateKey) [64]byte {
	h := sha512.Sum512(secretKey[:32])
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64

	return h
}
