// Package pkcs12 provides some implementations of PKCS#12.
//
// This implementation is distilled from https://tools.ietf.org/html/rfc7292 and referenced documents.
// It is intended for decoding P12/PFX-stored certificate+key for use with the crypto/tls package.
package pkcs12

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

type pfxPdu struct {
	Version  int
	AuthSafe contentInfo
	MacData  macData `asn1:"optional"`
}

type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"tag:0,explicit,optional"`
}

var (
	oidDataContentType          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidEncryptedDataContentType = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 6}

	localKeyId = []byte{0x01, 0x00, 0x00, 0x00}
)

type encryptedData struct {
	Version              int
	EncryptedContentInfo encryptedContentInfo
}

type encryptedContentInfo struct {
	ContentType                asn1.ObjectIdentifier
	ContentEncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedContent           []byte `asn1:"tag:0,optional"`
}

func (i encryptedContentInfo) GetAlgorithm() pkix.AlgorithmIdentifier {
	return i.ContentEncryptionAlgorithm
}
func (i encryptedContentInfo) GetData() []byte { return i.EncryptedContent }

type safeBag struct {
	Id         asn1.ObjectIdentifier
	Value      asn1.RawValue     `asn1:"tag:0,explicit"`
	Attributes []pkcs12Attribute `asn1:"set,optional"`
}

type pkcs12Attribute struct {
	Id    asn1.ObjectIdentifier
	Value asn1.RawValue `ans1:"set"`
}

type encryptedPrivateKeyInfo struct {
	AlgorithmIdentifier pkix.AlgorithmIdentifier
	EncryptedData       []byte
}

func (i encryptedPrivateKeyInfo) GetAlgorithm() pkix.AlgorithmIdentifier { return i.AlgorithmIdentifier }
func (i encryptedPrivateKeyInfo) GetData() []byte                        { return i.EncryptedData }

// PEM block types
const (
	CertificateType = "CERTIFICATE"
	PrivateKeyType  = "PRIVATE KEY"
)

// unmarshal calls asn1.Unmarshal, but also returns an error if there is any
// trailing data after unmarshaling.
func unmarshal(in []byte, out interface{}) error {
	trailing, err := asn1.Unmarshal(in, out)
	if err != nil {
		return err
	}
	if len(trailing) != 0 {
		return errors.New("pkcs12: trailing data found")
	}
	return nil
}

// ConvertToPEM converts all "safe bags" contained in pfxData to PEM blocks.
func ConvertToPEM(pfxData []byte, password string) (blocks []*pem.Block, err error) {
	p, err := bmpString(password)

	defer func() { // clear out BMP version of the password before we return
		for i := 0; i < len(p); i++ {
			p[i] = 0
		}
	}()

	if err != nil {
		return nil, ErrIncorrectPassword
	}

	bags, p, err := getSafeContents(pfxData, p)

	blocks = make([]*pem.Block, 0, 2)
	for _, bag := range bags {
		var block *pem.Block
		block, err = convertBag(&bag, p)
		if err != nil {
			return
		}
		blocks = append(blocks, block)
	}

	return
}

func convertBag(bag *safeBag, password []byte) (*pem.Block, error) {
	b := new(pem.Block)

	for _, attribute := range bag.Attributes {
		k, v, err := convertAttribute(&attribute)
		if err != nil {
			return nil, err
		}
		if b.Headers == nil {
			b.Headers = make(map[string]string)
		}
		b.Headers[k] = v
	}

	switch {
	case bag.Id.Equal(oidCertBagType):
		b.Type = CertificateType
		certsData, err := decodeCertBag(bag.Value.Bytes)
		if err != nil {
			return nil, err
		}
		b.Bytes = certsData
	case bag.Id.Equal(oidPkcs8ShroudedKeyBagType):
		b.Type = PrivateKeyType

		key, err := decodePkcs8ShroudedKeyBag(bag.Value.Bytes, password)
		if err != nil {
			return nil, err
		}

		switch key := key.(type) {
		case *rsa.PrivateKey:
			b.Bytes = x509.MarshalPKCS1PrivateKey(key)
		case *ecdsa.PrivateKey:
			b.Bytes, err = x509.MarshalECPrivateKey(key)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	default:
		return nil, errors.New("don't know how to convert a safe bag of type " + bag.Id.String())
	}
	return b, nil
}

var (
	oidFriendlyName     = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 20}
	oidLocalKeyID       = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 21}
	oidMicrosoftCSPName = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 17, 1}
)

var attributeNameByOID = map[string]string{
	oidFriendlyName.String():     "friendlyName",
	oidLocalKeyID.String():       "localKeyId",
	oidMicrosoftCSPName.String(): "Microsoft CSP Name", // openssl-compatible
}

func convertAttribute(attribute *pkcs12Attribute) (key, value string, err error) {
	key = attributeNameByOID[attribute.Id.String()]
	switch {
	case attribute.Id.Equal(oidMicrosoftCSPName):
		fallthrough
	case attribute.Id.Equal(oidFriendlyName):
		if _, err = asn1.Unmarshal(attribute.Value.Bytes, &attribute.Value); err != nil {
			return
		}
		if value, err = decodeBMPString(attribute.Value.Bytes); err != nil {
			return
		}
	case attribute.Id.Equal(oidLocalKeyID):
		id := new([]byte)
		if _, err = asn1.Unmarshal(attribute.Value.Bytes, id); err != nil {
			return
		}
		value = fmt.Sprintf("% x", *id)
	default:
		err = errors.New("don't know how to handle attribute with OID " + attribute.Id.String())
		return
	}

	return key, value, nil
}

// Decode extracts a certificate and private key from pfxData.
// This function assumes that there is only one certificate and only one private key in the pfxData.
func Decode(pfxData []byte, utf8Password string) (privateKey interface{}, certificate *x509.Certificate, err error) {
	p, err := bmpString(utf8Password)
	defer func() { // clear out BMP version of the password before we return
		for i := 0; i < len(p); i++ {
			p[i] = 0
		}
	}()

	if err != nil {
		return nil, nil, err
	}
	bags, p, err := getSafeContents(pfxData, p)
	if err != nil {
		return nil, nil, err
	}

	if len(bags) != 2 {
		err = errors.New("expected exactly two safe bags in the PFX PDU")
		return
	}

	for _, bag := range bags {
		switch {
		case bag.Id.Equal(oidCertBagType):
			certsData, err := decodeCertBag(bag.Value.Bytes)
			if err != nil {
				return nil, nil, err
			}
			certs, err := x509.ParseCertificates(certsData)
			if err != nil {
				return nil, nil, err
			}
			if len(certs) != 1 {
				err = errors.New("expected exactly one certificate in the certBag")
				return nil, nil, err
			}
			certificate = certs[0]
		case bag.Id.Equal(oidPkcs8ShroudedKeyBagType):
			if privateKey, err = decodePkcs8ShroudedKeyBag(bag.Value.Bytes, p); err != nil {
				return nil, nil, err
			}
		}
	}

	if certificate == nil {
		return nil, nil, errors.New("certificate missing")
	}
	if privateKey == nil {
		return nil, nil, errors.New("private key missing")
	}

	return
}

func getSafeContents(p12Data, password []byte) (bags []safeBag, actualPassword []byte, err error) {
	pfx := new(pfxPdu)
	if _, err = asn1.Unmarshal(p12Data, pfx); err != nil {
		return nil, nil, fmt.Errorf("error reading P12 data: %v", err)
	}

	if pfx.Version != 3 {
		return nil, nil, NotImplementedError("can only decode v3 PFX PDU's")
	}

	if !pfx.AuthSafe.ContentType.Equal(oidDataContentType) {
		return nil, nil, NotImplementedError("only password-protected PFX is implemented")
	}

	// unmarshal the explicit bytes in the content for type 'data'
	if _, err = asn1.Unmarshal(pfx.AuthSafe.Content.Bytes, &pfx.AuthSafe.Content); err != nil {
		return nil, nil, err
	}

	actualPassword = password
	password = nil
	if len(pfx.MacData.Mac.Algorithm.Algorithm) > 0 {
		if err = verifyMac(&pfx.MacData, pfx.AuthSafe.Content.Bytes, actualPassword); err != nil {
			if err == ErrIncorrectPassword && bytes.Compare(actualPassword, []byte{0, 0}) == 0 {
				// some implementations use an empty byte array for the empty string password
				// try one more time with empty-empty password
				actualPassword = []byte{}
				err = verifyMac(&pfx.MacData, pfx.AuthSafe.Content.Bytes, actualPassword)
			}
		}
		if err != nil {
			return
		}
	}

	var authenticatedSafe []contentInfo
	if _, err = asn1.Unmarshal(pfx.AuthSafe.Content.Bytes, &authenticatedSafe); err != nil {
		return
	}

	if len(authenticatedSafe) != 2 {
		return nil, nil, NotImplementedError("expected exactly two items in the authenticated safe")
	}

	for _, ci := range authenticatedSafe {
		var data []byte
		switch {
		case ci.ContentType.Equal(oidDataContentType):
			if _, err = asn1.Unmarshal(ci.Content.Bytes, &data); err != nil {
				return
			}
		case ci.ContentType.Equal(oidEncryptedDataContentType):
			var encryptedData encryptedData
			if _, err = asn1.Unmarshal(ci.Content.Bytes, &encryptedData); err != nil {
				return
			}
			if encryptedData.Version != 0 {
				return nil, nil, NotImplementedError("only version 0 of EncryptedData is supported")
			}
			if data, err = pbDecrypt(encryptedData.EncryptedContentInfo, actualPassword); err != nil {
				return
			}
		default:
			return nil, nil, NotImplementedError("only data and encryptedData content types are supported in authenticated safe")
		}

		var safeContents []safeBag
		if _, err = asn1.Unmarshal(data, &safeContents); err != nil {
			return
		}
		bags = append(bags, safeContents...)
	}
	return
}

// == Encode ====================================

func getLocalKeyId(id []byte) (attribute pkcs12Attribute, err error) {
	octetString := asn1.RawValue{Tag: 4, Class: 0, IsCompound: false, Bytes: id}
	bytes, err := asn1.Marshal(octetString)
	if err != nil {
		return
	}

	attribute = pkcs12Attribute{
		Id:    oidLocalKeyID,
		Value: asn1.RawValue{Tag: 17, Class: 0, IsCompound: true, Bytes: bytes},
	}

	return attribute, nil
}

func convertToRawVal(val interface{}) (raw asn1.RawValue, err error) {
	bytes, err := asn1.Marshal(val)
	if err != nil {
		return
	}

	_, err = asn1.Unmarshal(bytes, &raw)
	return raw, nil
}

func makeSafeBags(oid asn1.ObjectIdentifier, value []byte) ([]safeBag, error) {
	attribute, err := getLocalKeyId(localKeyId)

	if err != nil {
		return nil, EncodeError("local key id: " + err.Error())
	}

	bag := make([]safeBag, 1)
	bag[0] = safeBag{
		Id:         oid,
		Value:      asn1.RawValue{Tag: 0, Class: 2, IsCompound: true, Bytes: value},
		Attributes: []pkcs12Attribute{attribute},
	}

	return bag, nil
}

func makeCertBagContentInfo(derBytes []byte) (*contentInfo, error) {
	certBag1 := certBag{
		Id:   oidCertTypeX509Certificate,
		Data: derBytes,
	}

	bytes, err := asn1.Marshal(certBag1)
	if err != nil {
		return nil, EncodeError("encoding cert bag: " + err.Error())
	}

	certSafeBags, err := makeSafeBags(oidCertBagType, bytes)
	if err != nil {
		return nil, EncodeError("safe bags: " + err.Error())
	}

	return makeContentInfo(certSafeBags)
}

func makeShroudedKeyBagContentInfo(privateKey interface{}, password []byte) (*contentInfo, error) {
	shroudedKeyBagBytes, err := encodePkcs8ShroudedKeyBag(privateKey, password)
	if err != nil {
		return nil, EncodeError("encode PKCS#8 shrouded key bag: " + err.Error())
	}

	safeBags, err := makeSafeBags(oidPkcs8ShroudedKeyBagType, shroudedKeyBagBytes)
	if err != nil {
		return nil, EncodeError("safe bags: " + err.Error())
	}

	return makeContentInfo(safeBags)
}

func makeContentInfo(val interface{}) (*contentInfo, error) {
	fullBytes, err := asn1.Marshal(val)
	if err != nil {
		return nil, EncodeError("contentInfo raw value marshal: " + err.Error())
	}

	octetStringVal := asn1.RawValue{Tag: 4, Class: 0, IsCompound: false, Bytes: fullBytes}
	octetStringFullBytes, err := asn1.Marshal(octetStringVal)
	if err != nil {
		return nil, EncodeError("raw contentInfo to octet string: " + err.Error())
	}

	contentInfo := contentInfo{ContentType: oidDataContentType}
	contentInfo.Content = asn1.RawValue{Tag: 0, Class: 2, IsCompound: true, Bytes: octetStringFullBytes}

	return &contentInfo, nil
}

func makeContentInfos(derBytes []byte, privateKey interface{}, password []byte) ([]contentInfo, error) {
	shroudedKeyContentInfo, err := makeShroudedKeyBagContentInfo(privateKey, password)
	if err != nil {
		return nil, EncodeError("shrouded key content info: " + err.Error())
	}

	certBagContentInfo, err := makeCertBagContentInfo(derBytes)
	if err != nil {
		return nil, EncodeError("cert bag content info: " + err.Error())
	}

	contentInfos := make([]contentInfo, 2)
	contentInfos[0] = *shroudedKeyContentInfo
	contentInfos[1] = *certBagContentInfo

	return contentInfos, nil
}

func makeSalt(saltByteCount int) ([]byte, error) {
	salt := make([]byte, saltByteCount)
	_, err := io.ReadFull(rand.Reader, salt)
	return salt, err
}

// Encode converts a certificate and a private key to the PKCS#12 byte stream format.
//
// derBytes is a DER encoded certificate.
// privateKey is an RSA
func Encode(derBytes []byte, privateKey interface{}, password string) (pfxBytes []byte, err error) {
	secret, err := bmpString(password)
	if err != nil {
		return nil, ErrIncorrectPassword
	}

	contentInfos, err := makeContentInfos(derBytes, privateKey, secret)
	if err != nil {
		return nil, err
	}

	// Marhsal []contentInfo so we can re-constitute the byte stream that will
	// be suitable for computing the MAC
	bytes, err := asn1.Marshal(contentInfos)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an asn1.RawValue so, we can compute the MAC against the .Bytes
	var contentInfosRaw asn1.RawValue
	err = unmarshal(bytes, &contentInfosRaw)
	if err != nil {
		return nil, err
	}

	authSafeContentInfo, err := makeContentInfo(contentInfosRaw)
	if err != nil {
		return nil, EncodeError("authSafe content info: " + err.Error())
	}

	salt, err := makeSalt(pbeSaltSizeBytes)
	if err != nil {
		return nil, EncodeError("salt value: " + err.Error())
	}

	// Compute the MAC for marshaled bytes of contentInfos, which includes the
	// cert bag, and the shrouded key bag.
	digest := computeMac(contentInfosRaw.FullBytes, pbeIterationCount, salt, secret)

	pfx := pfxPdu{
		Version:  3,
		AuthSafe: *authSafeContentInfo,
		MacData: macData{
			Iterations: pbeIterationCount,
			MacSalt:    salt,
			Mac: digestInfo{
				Algorithm: pkix.AlgorithmIdentifier{
					Algorithm: oidSha1Algorithm,
				},
				Digest: digest,
			},
		},
	}

	bytes, err = asn1.Marshal(pfx)
	if err != nil {
		return nil, EncodeError("marshal PFX PDU: " + err.Error())
	}

	return bytes, err
}
