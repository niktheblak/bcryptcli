package scrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"

	stdscrypt "golang.org/x/crypto/scrypt"
)

const (
	defaultN      = 32768
	defaultR      = 8
	defaultP      = 1
	defaultKeyLen = 32
	saltLen       = 8
)

var (
	ErrHashTooShort              = errors.New("hashedSecret too short to be a scrypted password")
	ErrMismatchedHashAndPassword = errors.New("hashedPassword is not the hash of the given password")
)

var Encoding = base64.RawURLEncoding

func DecodeSaltAndHash(encodedHash []byte) (salt []byte, hash []byte, err error) {
	decoded := make([]byte, Encoding.DecodedLen(len(encodedHash)))
	n, err := Encoding.Decode(decoded, hash)
	if err != nil {
		return
	}
	decoded = decoded[:n]
	if len(decoded) < saltLen {
		err = ErrHashTooShort
		return
	}
	salt = decoded[:saltLen]
	hash = decoded[saltLen:]
	return
}

func CompareHashAndPassword(hashedPassword, password []byte) error {
	salt, key, err := DecodeSaltAndHash(hashedPassword)
	if err != nil {
		return err
	}
	expected, err := HashWithSalt(salt, password)
	if err != nil {
		return err
	}
	res := subtle.ConstantTimeCompare(key, expected)
	if res == 1 {
		return nil
	}
	return ErrMismatchedHashAndPassword
}

func HashWithSalt(salt, password []byte) ([]byte, error) {
	return stdscrypt.Key(password, salt, defaultN, defaultR, defaultP, defaultKeyLen)
}

func Hash(password []byte) (salt []byte, hash []byte, err error) {
	salt = make([]byte, saltLen)
	_, err = rand.Read(salt)
	if err != nil {
		return
	}
	hash, err = HashWithSalt(salt, password)
	return
}

func GenerateFromPassword(password []byte) (encodedHash string, err error) {
	var buf bytes.Buffer
	enc := base64.NewEncoder(Encoding, &buf)
	salt, hash, err := Hash(password)
	if err != nil {
		return
	}
	enc.Write(salt)
	enc.Write(hash)
	enc.Close()
	encodedHash = buf.String()
	return
}