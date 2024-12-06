package x19sing

import "testing"

func TestCipher(t *testing.T) {
	cipher, err := New("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	encryptedData, err := cipher.Encrypt("hello world")
	if err != nil {
		t.Fatal(err)
	}
	decryptedData, err := cipher.Decrypt(encryptedData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decryptedData)
}
