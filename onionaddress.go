package shrek

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"io"
	"io/fs"
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

const (
	publicKeyFileName = "hs_ed25519_public_key"
	secretKeyFileName = "hs_ed25519_secret_key"
	hostNameFileName  = "hostname"

	publicKeyFileHeader = "== ed25519v1-public: type0 ==\x00\x00\x00"
	secretKeyFileHeader = "== ed25519v1-secret: type0 ==\x00\x00\x00"
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

	pkFile := filepath.Join(dir, publicKeyFileName)
	pkData := append([]byte(publicKeyFileHeader), addr.PublicKey...)
	if err := os.WriteFile(pkFile, pkData, fileMode); err != nil {
		return fmt.Errorf("shrek: could not save public key to file: %w", err)
	}

	skFile := filepath.Join(dir, secretKeyFileName)
	skData := append([]byte(secretKeyFileHeader), addr.SecretKey...)
	if err := os.WriteFile(skFile, skData, fileMode); err != nil {
		return fmt.Errorf("shrek: could not save secret key to file: %w", err)
	}

	hnFile := filepath.Join(dir, hostNameFileName)
	hnData := []byte(hostname)
	if err := os.WriteFile(hnFile, hnData, fileMode); err != nil {
		return fmt.Errorf("shrek: could not save onion hostname to file: %w", err)
	}

	return nil
}

// ReadOnionAddress reads the public key and secret key from the files in the given
// directory, then it parses the keys from the files inside the directory and validates
// that they are valid keys to use as an onion address.
//
// The provided directory must be one created either by the SaveOnionAddress function
// or any other program that outputs the keys in the same format. The directory must
// contain the following files:
//
//   hs_ed25519_public_key
//   hs_ed25519_secret_key
//
func ReadOnionAddress(dir string) (*OnionAddress, error) {
	// Check dir exists.
	if fi, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("shrek: directory not found: %q", dir)
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("shrek: path is not a directory: %q", dir)
	}

	return ReadOnionAddressFS(os.DirFS(dir))
}

// ReadOnionAddressFS does the same thing as ReadOnionAddress. The only difference is
// that it accepts an fs.FS to abstract away the underlying file system.
func ReadOnionAddressFS(fsys fs.FS) (*OnionAddress, error) {
	// Read public key from file and validate contents.
	pkData, err := fs.ReadFile(fsys, publicKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("shrek: reading public key file: %w", err)
	}
	if l := len(pkData); l != len(publicKeyFileHeader)+ed25519.PublicKeySize {
		return nil, fmt.Errorf("shrek: public key file has wrong length: %d", l)
	}

	// Read private key from file and validate contents.
	skData, err := fs.ReadFile(fsys, secretKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("shrek: reading secret key file: %w", err)
	}
	if l := len(skData); l != len(secretKeyFileHeader)+ed25519.PrivateKeySize {
		return nil, fmt.Errorf("shrek: secret key file has wrong length: %d", l)
	}

	kp := &ed25519.KeyPair{
		PublicKey:  ed25519.PublicKey(pkData[len(publicKeyFileHeader):]),
		PrivateKey: ed25519.PrivateKey(skData[len(secretKeyFileHeader):]),
	}

	// Validate keys match.
	if err := kp.Validate(); err != nil {
		return nil, fmt.Errorf("shrek: keys in directory do not match: %w", err)
	}

	return &OnionAddress{
		PublicKey: kp.PublicKey,
		SecretKey: kp.PrivateKey,
	}, nil
}
