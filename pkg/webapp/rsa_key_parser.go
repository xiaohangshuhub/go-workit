package webapp

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"math/big"
)

// parseRSAPublicKey parses a base64-encoded RSA public key and returns a *rsa.PublicKey.
func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}

	e := 0
	if len(eBytes) == 3 {
		e = int(binary.BigEndian.Uint32(append([]byte{0}, eBytes...)))
	} else if len(eBytes) == 4 {
		e = int(binary.BigEndian.Uint32(eBytes))
	} else {
		// fallback
		e = int(big.NewInt(0).SetBytes(eBytes).Int64())
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}

	return pubKey, nil
}
