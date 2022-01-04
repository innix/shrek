package ed25519

import (
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

func scalarAdd(dst *scalar.Scalar, v uint64) error {
	var dstb [32]byte

	if err := dst.ToBytes(dstb[:]); err != nil {
		return err
	}

	scalarAddBytes(&dstb, v)

	if _, err := dst.SetBits(dstb[:]); err != nil {
		return err
	}

	return nil
}

func scalarAddBytes(dst *[32]byte, v uint64) {
	var carry uint32

	for i := 0; i < 32; i++ {
		carry += uint32(dst[i]) + uint32(v&0xFF)
		dst[i] = byte(carry & 0xFF)
		carry >>= 8

		v >>= 8
	}
}
