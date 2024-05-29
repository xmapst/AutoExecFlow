package rotate

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

// 生成随机字节数组
func generateRandomBytes(size int) []byte {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// BenchmarkEncrypt 压力测试加密
func BenchmarkEncrypt(b *testing.B) {
	key := generateRandomBytes(128) // 128 字节密钥
	_cipher := New(key)
	data := generateRandomBytes(1024 * 1024 * 10) // 10MB 数据

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_cipher.Encrypt(data)
	}
}

// BenchmarkDecrypt 压力测试解密
func BenchmarkDecrypt(b *testing.B) {
	key := generateRandomBytes(128) // 128 字节密钥
	_cipher := New(key)
	data := generateRandomBytes(1024 * 1024 * 10) // 10MB 数据
	encryptedData := bytes.Buffer{}
	err := _cipher.EncryptReader(bytes.NewReader(data), &encryptedData)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = _cipher.DecryptReader(&encryptedData, io.Discard)
	}
}

// BenchmarkEncryptStream 压力测试流式加密
func BenchmarkEncryptStream(b *testing.B) {
	key := generateRandomBytes(128) // 128 字节密钥
	_cipher := New(key)
	data := bytes.NewReader(generateRandomBytes(1024 * 1024 * 10)) // 10MB 数据
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = _cipher.EncryptReader(data, io.Discard)
	}
}

// BenchmarkEncryptStream 压力测试流式加密
func BenchmarkDecryptStream(b *testing.B) {
	key := generateRandomBytes(128) // 128 字节密钥
	_cipher := New(key)
	data := bytes.NewReader(generateRandomBytes(1024 * 1024 * 10)) // 10MB 数据
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = _cipher.EncryptReader(data, io.Discard)
	}
}

// TestEncryptDecrypt 单元测试加密和解密
func TestEncryptDecrypt(t *testing.T) {
	key := []byte("mysecretkey") // 固定密钥
	_cipher := New(key)

	tests := []struct {
		name string
		data []byte
	}{
		{"EmptyData", []byte{}},
		{"SmallData", []byte("Hello, World!")},
		{"LargeData", generateRandomBytes(1024 * 1024 * 10)}, // 10MB 数据
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptedData := _cipher.Encrypt(tt.data)
			decryptedData := _cipher.Decrypt(encryptedData)

			if !bytes.Equal(tt.data, decryptedData) {
				t.Errorf("decrypted data does not match original, got %v, want %v", decryptedData, tt.data)
			}
		})
	}
}

// TestEncryptDecrypt 单元测试流式加密和解密
func TestEncryptDecryptStream(t *testing.T) {
	key := []byte("mysecretkey") // 固定密钥
	_cipher := New(key)

	tests := []struct {
		name string
		data []byte
	}{
		{"EmptyData", []byte{}},
		{"SmallData", []byte("Hello, World!")},
		{"LargeData", generateRandomBytes(1024 * 1024 * 10)}, // 10MB 数据
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptedData := bytes.Buffer{}
			err := _cipher.EncryptReader(bytes.NewReader(tt.data), &encryptedData)
			if err != nil {
				t.Error(err)
			}
			decryptedData := bytes.Buffer{}
			err = _cipher.DecryptReader(&encryptedData, &decryptedData)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(tt.data, decryptedData.Bytes()) {
				t.Errorf("decrypted data does not match original, got %v, want %v", decryptedData, tt.data)
			}
		})
	}
}

// TestKeyChangeEffectiveness 测试密钥变化的影响
func TestKeyChangeEffectiveness(t *testing.T) {
	key1 := []byte("mysecretkey1")
	key2 := []byte("mysecretkey2")
	cipher1 := New(key1)
	cipher2 := New(key2)

	data := []byte("Sensitive Data")
	encryptedData1 := cipher1.Encrypt(data)
	encryptedData2 := cipher2.Encrypt(data)

	if bytes.Equal(encryptedData1, encryptedData2) {
		t.Errorf("encryption with different keys should produce different results, got %v and %v", encryptedData1, encryptedData2)
	}
}
