package rotate

import (
	"io"
	"sync"
)

// ICipher 循环移位加解密算法
type ICipher interface {
	Encrypt([]byte) []byte
	Decrypt([]byte) []byte
	EncryptReader(io.Reader, io.Writer) error
	DecryptReader(io.Reader, io.Writer) error
}

// cipher represents a Caesar cipher with rotation key
type cipher struct {
	key    []byte // 密钥
	keyLen int    // 密钥长度
}

func New(key []byte) ICipher {
	return &cipher{
		key:    key,
		keyLen: len(key),
	}
}

// Encrypt 加密
func (c *cipher) Encrypt(data []byte) []byte {
	return c.process(data, true)
}

// Decrypt 解密
func (c *cipher) Decrypt(data []byte) []byte {
	return c.process(data, false)
}

// EncryptReader 加密
func (c *cipher) EncryptReader(r io.Reader, w io.Writer) error {
	return c.stream(r, w, true)
}

// DecryptReader 解密
func (c *cipher) DecryptReader(r io.Reader, w io.Writer) error {
	return c.stream(r, w, false)
}

func (c *cipher) stream(r io.Reader, w io.Writer, encrypt bool) error {
	if c.key == nil {
		_, err := io.Copy(w, r)
		return err
	}
	var buf = make([]byte, c.keyLen*15)
	var err error
	for {
		var n int
		n, err = r.Read(buf)
		if n > 0 {
			_, err = w.Write(c.process(buf[:n], encrypt))
			if err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func (c *cipher) process(data []byte, encrypt bool) []byte {
	if c.key == nil {
		return data
	}
	dataLen := len(data)
	result := make([]byte, dataLen)
	var wg sync.WaitGroup

	numChunks := (dataLen + c.keyLen - 1) / c.keyLen
	for i := 0; i < numChunks; i++ {
		start := i * c.keyLen
		end := start + c.keyLen
		if end > dataLen {
			end = dataLen
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			if encrypt {
				for j := start; j < end; j++ {
					result[j] = data[j] ^ c.key[j%c.keyLen]
				}
				c.shiftRight(result[start:end])
			} else {
				c.shiftLeft(data[start:end])
				for j := start; j < end; j++ {
					result[j] = data[j] ^ c.key[j%c.keyLen]
				}
			}
		}(start, end)
	}
	wg.Wait()
	return result
}

// 右移
func (c *cipher) shiftRight(data []byte) {
	for i := 0; i < len(data); i++ {
		shift := int(c.key[i%c.keyLen]) % 8
		data[i] = (data[i] >> shift) | (data[i] << (8 - shift))
	}
}

// 左移
func (c *cipher) shiftLeft(data []byte) {
	for i := 0; i < len(data); i++ {
		shift := int(c.key[i%c.keyLen]) % 8
		data[i] = (data[i] << shift) | (data[i] >> (8 - shift))
	}
}
