package backend

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/secretbox"
)

var (
	ErrFailedToDecrypt = errors.New("failed to decrypt")
)

type Crypt struct {
	PSK string
}

type CryptRequest struct {
	ID   string
	Data []byte
	Pin  string
}

func (c Crypt) Encrypt(req CryptRequest) ([]byte, error) {
	naclKey := c.getNaclKey(c.PSK, req.ID, req.Pin)

	nonce := new([24]byte)
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	out := make([]byte, 24)
	copy(out, nonce[:])

	sealed := secretbox.Seal(out, req.Data, nonce, naclKey)
	return []byte(base64.StdEncoding.EncodeToString(sealed)), nil
}

func (c Crypt) Decrypt(req CryptRequest) ([]byte, error) {
	naclKey := c.getNaclKey(c.PSK, req.ID, req.Pin)

	sealed, err := base64.StdEncoding.DecodeString(string(req.Data))
	if err != nil {
		return nil, err
	}

	nonce := new([24]byte)
	copy(nonce[:], sealed[:24])

	decrypted, ok := secretbox.Open(nil, sealed[24:], nonce, naclKey)
	if !ok {
		return nil, ErrFailedToDecrypt
	}

	return decrypted, nil
}

func (c *Crypt) getNaclKey(keys ...string) *[32]byte {
	keyHashed := blake2b.Sum512([]byte(strings.Join(keys, ",")))

	bytes := bytes.NewBuffer([]byte{})

	for i := 0; i < len(keyHashed) && bytes.Len() < 32; i++ {
		if i%2 == 0 {
			bytes.WriteByte(keyHashed[i])
		}
	}

	for bytes.Len() < 32 {
		bytes.WriteByte(0xff)
	}

	naclKey := new([32]byte)
	copy(naclKey[:], bytes.Bytes())

	return naclKey
}
