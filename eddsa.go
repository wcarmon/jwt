package jwt

import (
	"crypto/ed25519"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
)

type algEdDSA struct {
	name string
}

func (a *algEdDSA) Name() string {
	return a.name
}

func (a *algEdDSA) Sign(key PrivateKey, headerAndPayload []byte) ([]byte, error) {
	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrInvalidKey
	}

	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidKey
	}

	return ed25519.Sign(privateKey, []byte(headerAndPayload)), nil
}

func (a *algEdDSA) Verify(key PublicKey, headerAndPayload []byte, signature []byte) error {
	publicKey, ok := key.(ed25519.PublicKey)
	if !ok {
		if privateKey, ok := key.(ed25519.PrivateKey); ok {
			publicKey = privateKey.Public().(ed25519.PublicKey)
		} else {
			return ErrInvalidKey
		}
	}

	if len(publicKey) != ed25519.PublicKeySize {
		return ErrInvalidKey
	}

	if !ed25519.Verify(publicKey, headerAndPayload, signature) {
		return ErrTokenSignature
	}

	return nil
}

// Key Helpers.

// MustLoadEdDSA accepts private and public PEM filenames
// and returns a pair of private and public ed25519 keys.
// Pass the returned private key to `Token` (signing) function
// and the public key to the `VerifyToken` function.
//
// It panics on errors.
func MustLoadEdDSA(privateKeyFilename, publicKeyFilename string) (ed25519.PrivateKey, ed25519.PublicKey) {
	privateKey, err := LoadPrivateKeyEdDSA(privateKeyFilename)
	if err != nil {
		panic(err)
	}

	publicKey, err := LoadPublicKeyEdDSA(publicKeyFilename)
	if err != nil {
		panic(err)
	}

	return privateKey, publicKey
}

// LoadPrivateKeyEdDSA accepts a file path of a PEM-encoded ed25519 private key
// and returns the ed25519 private key Go value.
// Pass the returned value to the `Token` (signing) function.
func LoadPrivateKeyEdDSA(filename string) (ed25519.PrivateKey, error) {
	b, err := ReadFile(filename)
	if err != nil {
		return nil, err
	}

	key, err := ParsePrivateKeyEdDSA(b)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// LoadPublicKeyEdDSA accepts a file path of a PEM-encoded ed25519 public key
// and returns the ed25519 public key Go value.
// Pass the returned value to the `VerifyToken` function.
func LoadPublicKeyEdDSA(filename string) (ed25519.PublicKey, error) {
	b, err := ReadFile(filename)
	if err != nil {
		return nil, err
	}

	key, err := ParsePublicKeyEdDSA(b)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ParsePrivateKeyEdDSA decodes and parses the
// PEM-encoded ed25519 private key's raw contents.
// Pass the result to the `Token` (signing) function.
func ParsePrivateKeyEdDSA(key []byte) (ed25519.PrivateKey, error) {
	asn1PrivKey := struct {
		Version          int
		ObjectIdentifier struct {
			ObjectIdentifier asn1.ObjectIdentifier
		}
		PrivateKey []byte
	}{}

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("private key: malformed or missing PEM format (EdDSA)")
	}

	if _, err := asn1.Unmarshal(block.Bytes, &asn1PrivKey); err != nil {
		return nil, err
	}

	privateKey := ed25519.NewKeyFromSeed(asn1PrivKey.PrivateKey[2:])
	return privateKey, nil
}

// ParsePublicKeyEdDSA decodes and parses the
// PEM-encoded ed25519 public key's raw contents.
// Pass the result to the `VerifyToken` function.
func ParsePublicKeyEdDSA(key []byte) (ed25519.PublicKey, error) {
	asn1PubKey := struct {
		OBjectIdentifier struct {
			ObjectIdentifier asn1.ObjectIdentifier
		}
		PublicKey asn1.BitString
	}{}

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("public key: malformed or missing PEM format (EdDSA)")
	}

	if _, err := asn1.Unmarshal(block.Bytes, &asn1PubKey); err != nil {
		return nil, err
	}

	publicKey := ed25519.PublicKey(asn1PubKey.PublicKey.Bytes)
	return publicKey, nil
}