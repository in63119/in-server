package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestDecryptOpenSSLCompatible(t *testing.T) {
	cipherText := "U2FsdGVkX1+r86z3OtRl+pkXmqsTpb/+vXazcXDOayI="

	plain, err := Decrypt(cipherText, "mysecret")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plain != "hello world" {
		t.Fatalf("unexpected plaintext: %q", plain)
	}
}

func TestEVPBytesToKeyMatchesOpenSSL(t *testing.T) {
	salt := []byte{0xab, 0xf3, 0xac, 0xf7, 0x3a, 0xd4, 0x65, 0xfa}
	wantKey, _ := hex.DecodeString("376950CA734B1B72FDCD8F809FD9DEE6E12D3F8FEDB5DDDDF6F6D1AAECC173BA")
	wantIV, _ := hex.DecodeString("63A6CE42D65165AEF6304DF274FD21AF")

	key, iv := evpBytesToKey([]byte("mysecret"), salt, 32, 16)

	if !bytes.Equal(key, wantKey) {
		t.Fatalf("unexpected key: %x", key)
	}
	if !bytes.Equal(iv, wantIV) {
		t.Fatalf("unexpected iv: %x", iv)
	}
}

func TestEVPBytesToKeyMD5Legacy(t *testing.T) {
	salt := []byte{0xab, 0xf3, 0xac, 0xf7, 0x3a, 0xd4, 0x65, 0xfa}
	wantKey, _ := hex.DecodeString("0936EEF8F283AACBA2DD86F3FBE5A09AB2C974D33C0468D62262537D40EA6284")
	wantIV, _ := hex.DecodeString("CF4988CC9F8919C594B1A8EB595CF391")

	key, iv := evpBytesToKeyMD5([]byte("mysecret"), salt, 32, 16)

	if !bytes.Equal(key, wantKey) {
		t.Fatalf("unexpected key: %x", key)
	}
	if !bytes.Equal(iv, wantIV) {
		t.Fatalf("unexpected iv: %x", iv)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	secret := "s3cr3t!"
	message := "owner-private-key-1234"

	cipherText, err := Encrypt(message, secret)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	out, err := Decrypt(cipherText, secret)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if out != message {
		t.Fatalf("round trip mismatch: %q != %q", out, message)
	}
}
