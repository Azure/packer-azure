package pkcs12

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/x509/pkix"
	"encoding/asn1"
)

type macData struct {
	Mac        digestInfo
	MacSalt    []byte
	Iterations int `asn1:"optional,default:1"`
}

// from PKCS#7:
type digestInfo struct {
	Algorithm pkix.AlgorithmIdentifier
	Digest    []byte
}

const (
	sha1Algorithm = "SHA-1"
)

var (
	oidSha1Algorithm = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
)

func verifyMac(macData *macData, message, password []byte) error {
	if !macData.Mac.Algorithm.Algorithm.Equal(oidSha1Algorithm) {
		return NotImplementedError("unknown digest algorithm: " + macData.Mac.Algorithm.Algorithm.String())
	}

	expectedMAC := computeMac(message, macData.Iterations, macData.MacSalt, password)

	if !hmac.Equal(macData.Mac.Digest, expectedMAC) {
		return ErrIncorrectPassword
	}
	return nil
}

func computeMac(message []byte, iterations int, salt, password []byte) []byte {
	key := pbkdf(sha1Sum, 20, 64, salt, password, iterations, 3, 20)

	mac := hmac.New(sha1.New, key)
	mac.Write(message)

	return mac.Sum(nil)
}
