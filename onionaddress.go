package shrek

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/innix/shrek/internal/ed25519"
	"golang.org/x/crypto/sha3"
)

const (
	// EncodedPublicKeyApproxSize is the size, in bytes, of the public key when
	// encoded using the approximate encoder.
	EncodedPublicKeyApproxSize = 51

	// EncodedPublicKeySize is the size, in bytes, of the public key when encoded
	// using the real encoder.
	EncodedPublicKeySize = 56
)

var b32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

type OnionAddress struct {
	PublicKey ed25519.PublicKey
	SecretKey ed25519.PrivateKey
}

// HostName returns the .onion address representation of the public key stored in
// the OnionAddress. The .onion TLD is not included.
func (addr *OnionAddress) HostName(hostname []byte) {
	const version = 3

	if l := len(hostname); l != EncodedPublicKeySize {
		panic(fmt.Sprintf("bad buffer length: %d", l))
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

// HostNameString returns the .onion address representation of the public key stored
// in the OnionAddress as a string. Unlike HostName and HostNameApprox, this method
// does include the .onion TLD in the returned hostname.
func (addr *OnionAddress) HostNameString() string {
	hostname := make([]byte, EncodedPublicKeySize)
	addr.HostName(hostname)

	return fmt.Sprintf("%s.onion", hostname)
}

// HostNameApprox returns an approximate .onion address representation of the public
// key stored in the OnionAddress. The start of the address is accurate, the last few
// characters at the end are not. The .onion TLD is not included.
func (addr *OnionAddress) HostNameApprox(hostname []byte) {
	if l := len(hostname); l != EncodedPublicKeySize {
		panic(fmt.Sprintf("bad buffer length: %d", l))
	}

	b32.Encode(hostname, addr.PublicKey)
}

func GenerateOnionAddress(rand io.Reader) (*OnionAddress, error) {
	kp, err := ed25519.GenerateKey(rand)
	if err != nil {
		return nil, fmt.Errorf("shrek: could not generate onion address: %w", err)
	}

	return &OnionAddress{
		PublicKey: kp.PublicKey,
		SecretKey: kp.PrivateKey,
	}, nil
}

// SaveOnionAddress saves the hostname, public key, and secret key from the given
// OnionAddress to the destination directory. It creates a sub-directory named after
// the hostname in the destination directory, then it creates 3 files inside the
// created sub-directory:
//
//   hs_ed25519_public_key
//   hs_ed25519_secret_key
//   hostname
//
func SaveOnionAddress(dir string, addr *OnionAddress) error {
	const (
		dirMode  = 0o700
		fileMode = 0o600
	)

	hostname := addr.HostNameString()
	dir = filepath.Join(dir, hostname)

	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("shrek: could not create directories: %w", err)
	}

	pk := addr.PublicKey
	pkFile := filepath.Join(dir, "hs_ed25519_public_key")
	pkData := append([]byte("== ed25519v1-public: type0 ==\x00\x00\x00"), pk...)
	if err := os.WriteFile(pkFile, pkData, fileMode); err != nil {
		return fmt.Errorf("shrek: could not save public key to file: %w", err)
	}

	skFile := filepath.Join(dir, "hs_ed25519_secret_key")
	skData := append([]byte("== ed25519v1-secret: type0 ==\x00\x00\x00"), addr.SecretKey...)
	if err := os.WriteFile(skFile, skData, fileMode); err != nil {
		return fmt.Errorf("shrek: could not save secret key to file: %w", err)
	}

	hnFile := filepath.Join(dir, "hostname")
	hnData := []byte(hostname)
	if err := os.WriteFile(hnFile, hnData, fileMode); err != nil {
		return fmt.Errorf("shrek: could not save onion hostname to file: %w", err)
	}

	return nil
}
