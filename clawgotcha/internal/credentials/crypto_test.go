package credentials

import (
	"bytes"
	"testing"
)

func TestParseMasterKey_Raw32(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	got, err := ParseMasterKey(string(key))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, key) {
		t.Fatalf("raw mismatch")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte("x"), 32)
	pt := []byte(`{"pat":"ghp_test"}`)
	sealed, err := Encrypt(pt, key)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Decrypt(sealed, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(pt) {
		t.Fatalf("got %q", out)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := bytes.Repeat([]byte("a"), 32)
	sealed, _ := Encrypt([]byte("secret"), key)
	wrong := bytes.Repeat([]byte("b"), 32)
	_, err := Decrypt(sealed, wrong)
	if err == nil {
		t.Fatal("expected error")
	}
}
