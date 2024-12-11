package encrypt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

// Define a constant secret key (replace with your actual key)
var secretKey [32]byte

func init() {
	// Generate a random secret key (32 bytes)
	_, err := io.ReadFull(rand.Reader, secretKey[:])
	if err != nil {
		panic(err)
	}
}

// EncryptWithNaCl encrypts plaintext using NaCl secretbox and returns a Base64-encoded ciphertext.
func EncryptWithNaCl(plaintext []byte) (string, error) {
	// Generate a random nonce (24 bytes)
	var nonce [24]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return "", err
	}

	// Encrypt the data using secretbox
	encrypted := secretbox.Seal(nonce[:], plaintext, &nonce, &secretKey)

	// Encode the ciphertext as Base64
	encodedCiphertext := base64.StdEncoding.EncodeToString(encrypted)
	return encodedCiphertext, nil
}

// DecryptWithNaCl decrypts a Base64-encoded ciphertext using NaCl secretbox.
func DecryptWithNaCl(encodedCiphertext string) ([]byte, error) {
	// Decode the Base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	// Nonce size is 24 bytes
	if len(ciphertext) < 24 {
		return nil, errors.New("ciphertext is too short")
	}

	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])

	// Decrypt the data using secretbox
	decrypted, ok := secretbox.Open(nil, ciphertext[24:], &nonce, &secretKey)
	if !ok {
		return nil, errors.New("decryption failed")
	}

	return decrypted, nil
}
