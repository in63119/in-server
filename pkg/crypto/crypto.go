package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
)

const (
	saltedPrefix = "Salted__"
)

func Encrypt(plaintext, secret string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	key, iv := evpBytesToKey([]byte(secret), salt, 32, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	padded := pkcs7Pad([]byte(plaintext), block.BlockSize())
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	out := append([]byte(saltedPrefix), salt...)
	out = append(out, ciphertext...)
	return base64.StdEncoding.EncodeToString(out), nil
}

func Decrypt(encoded, secret string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	if len(raw) < len(saltedPrefix)+8 || string(raw[:len(saltedPrefix)]) != saltedPrefix {
		return "", errors.New("invalid cipher text")
	}

	salt := raw[len(saltedPrefix) : len(saltedPrefix)+8]
	ct := raw[len(saltedPrefix)+8:]

	plain, err := decryptWithParams(ct, []byte(secret), salt, evpBytesToKeySHA256)
	if err == nil {
		return plain, nil
	}

	if legacy, legacyErr := decryptWithParams(ct, []byte(secret), salt, evpBytesToKeyMD5); legacyErr == nil {
		return legacy, nil
	}

	return "", err
}

func decryptWithParams(ct, secret, salt []byte, deriver func([]byte, []byte, int, int) ([]byte, []byte)) (string, error) {
	key, iv := deriver(secret, salt, 32, 16)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	if len(ct)%block.BlockSize() != 0 {
		return "", errors.New("cipher text is not full blocks")
	}

	plaintext := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ct)

	unpadded, err := pkcs7Unpad(plaintext, block.BlockSize())
	if err != nil {
		return "", err
	}
	return string(unpadded), nil
}

func SHA256(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func evpBytesToKey(password, salt []byte, keyLen, ivLen int) ([]byte, []byte) {
	return evpBytesToKeySHA256(password, salt, keyLen, ivLen)
}

func evpBytesToKeySHA256(password, salt []byte, keyLen, ivLen int) ([]byte, []byte) {
	return evpBytesToKeyWithHash(password, salt, keyLen, ivLen, sha256.New)
}

func evpBytesToKeyMD5(password, salt []byte, keyLen, ivLen int) ([]byte, []byte) {
	return evpBytesToKeyWithHash(password, salt, keyLen, ivLen, md5.New)
}

func evpBytesToKeyWithHash(password, salt []byte, keyLen, ivLen int, hasher func() hash.Hash) ([]byte, []byte) {
	var d []byte
	var prev []byte
	for len(d) < keyLen+ivLen {
		hash := hasher()
		hash.Write(prev)
		hash.Write(password)
		hash.Write(salt)
		prev = hash.Sum(nil)
		d = append(d, prev...)
	}
	return d[:keyLen], d[keyLen : keyLen+ivLen]
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	pad := bytesRepeat(byte(padding), padding)
	return append(data, pad...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid pkcs7 data")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, errors.New("invalid pkcs7 padding")
	}
	for i := 0; i < pad; i++ {
		if data[len(data)-1-i] != byte(pad) {
			return nil, errors.New("invalid pkcs7 padding")
		}
	}
	return data[:len(data)-pad], nil
}

func bytesRepeat(b byte, count int) []byte {
	out := make([]byte, count)
	for i := range out {
		out[i] = b
	}
	return out
}
