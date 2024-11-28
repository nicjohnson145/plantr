package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type KeyOpts struct {
	Size *int
}

func GenerateKeyPair(opts *KeyOpts) (string, string, error) {
	size := 4096
	if opts != nil && opts.Size != nil {
		size = *opts.Size
	}

	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return "", "", fmt.Errorf("error generating key: %w", err)
	}

	public := key.Public()

	privPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(public.(*rsa.PublicKey)),
		},
	)

	return string(pubPEM), string(privPEM), nil
}

func EncodePrivateKey(key *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	))
}

func EncodePublicKey(key *rsa.PublicKey) string {
	return string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(key),
		},
	))
}

func DecodePublicKey(key string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(key))
	out, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}

	return out, nil
}

func DecodePrivateKey(key string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(key))
	out, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}

	return out, nil
}

func EncryptValue(msg string, key string) (string, error) {
	pubKey, err := DecodePublicKey(key)
	if err != nil {
		return "", fmt.Errorf("error decoding as public key: %w", err)
	}
	cipher, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, []byte(msg), nil)
	if err != nil {
		return "", fmt.Errorf("error encrypting: %w", err)
	}

	return base64.StdEncoding.EncodeToString(cipher), nil
}

func DecryptValue(cipher string, key string) (string, error) {
	privKey, err := DecodePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("error decoding as private key: %w", err)
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", fmt.Errorf("error decoding cipher from base64: %w", err)
	}

	msg, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, cipherBytes, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting: %w", err)
	}

	return string(msg), nil
}
