package x19sing

import (
	"encoding/binary"
	"errors"
)

const (
	// KeyLength is the fixed length of the key
	KeyLength = 32
)

type Cipher struct {
	key       []byte
	paddedKey []byte
	keyInt64  []int64
}

// New creates a new Cipher instance with the provided key.
func New(key string) (*Cipher, error) {
	if len(key) != KeyLength {
		return nil, errors.New("the key is not 32 bits")
	}
	c := &Cipher{key: []byte(key)}
	// Step 1: Pad the key
	c.paddedKey = c.padKey()

	// Step 2: Convert the padded key to int64 array
	c.keyInt64 = c.bytesToInt64Arr(c.paddedKey)
	return c, nil
}

// Decrypt decrypts the given encrypted data string.
func (c *Cipher) Decrypt(encryptedData string) (string, error) {
	if len(encryptedData) < 64 {
		return "", errors.New("encrypted data must be at least 64 characters long")
	}
	if len(encryptedData)%16 != 0 {
		return "", errors.New("encrypted data length must be a multiple of 16")
	}
	// Step 1: Convert the encrypted data to int64 array
	encryptedDataInt64 := c.stringToInt64Array(encryptedData)

	// Step 2: Decrypt the data
	decryptedDataInt64 := c.decrypt(encryptedDataInt64)

	// Step 3: Convert the decrypted data back to a string
	decryptedData := c.int64ArrayToString(decryptedDataInt64)
	return decryptedData, nil
}

// Encrypt encrypts the given plain text data string.
func (c *Cipher) Encrypt(plainData string) (string, error) {
	if len(plainData) == 0 {
		return "", errors.New("plain data cannot be empty")
	}
	// Step 1: Pad the input data to the correct length
	paddedData := c.padData([]byte(plainData))

	// Step 2: Convert the padded data into an int64 array
	dataInt64 := c.bytesToInt64Arr(paddedData)

	// Step 3: Encrypt the data using the padded key
	encryptedDataInt64 := c.encrypt(dataInt64)

	// Step 4: Convert the encrypted data back into a string
	encryptedData := c.int64ArrayToString(encryptedDataInt64)
	return encryptedData, nil
}

// decryptData performs the decryption process on the data using the provided key.
func (c *Cipher) decrypt(data []int64) []int64 {
	num := len(data)
	if num < 1 {
		return data
	}
	num2, num3, num4 := data[num-1], data[0], int64(6+52/num)
	for num5 := num4 * 2654435769; num5 != 0; num5 -= 2654435769 {
		num6 := num5 >> 2 & 3
		var num7 int64
		for num7 = int64(num - 1); num7 > 0; num7-- {
			num2 = data[num7-1]
			data[num7] -= (num2>>5 ^ num3<<2) + (num3>>3 ^ num2<<4) ^ ((num5 ^ num3) + (c.keyInt64[(num7&3^num6)] ^ num2))
			num3 = data[num7]
		}
		num2 = data[num-1]
		data[0] -= (num2>>5 ^ num3<<2) + (num3>>3 ^ num2<<4) ^ ((num5 ^ num3) + (c.keyInt64[(num7&3^num6)] ^ num2))
		num3 = data[0]
	}
	return data
}

// encryptData performs the encryption process on the data using the provided key.
func (c *Cipher) encrypt(data []int64) []int64 {
	num := len(data)
	if num < 1 {
		return data
	}
	num2, num3, num4, num5 := data[num-1], data[0], int64(0), int64(6+52/num)
	for ; num5 > 0; num5-- {
		num4 += int64(2654435769)
		num7 := (num4 >> 2) & 3
		var num8 int64
		for num8 = 0; num8 < int64(num-1); num8++ {
			num3 = data[num8+1]
			data[num8] += (num2>>5 ^ num3<<2) + (num3>>3 ^ num2<<4) ^ ((num4 ^ num3) + (c.keyInt64[(num8&3)^num7] ^ num2))
			num2 = data[num8]
		}
		num3 = data[0]
		data[num-1] += (num2>>5 ^ num3<<2) + (num3>>3 ^ num2<<4) ^ ((num4 ^ num3) + (c.keyInt64[(num8&3)^num7] ^ num2))
		num2 = data[num-1]
	}
	return data
}

// parseHexToInt64 converts a hex string to an int64 value.
func (c *Cipher) parseHexToInt64(hexStr string) int64 {
	var result int64
	for i := 0; i < len(hexStr); i++ {
		value := c.hexCharToValue(hexStr[i])
		if value == -1 {
			return 0
		}
		result = result*16 + int64(value)
	}
	return result
}

// hexCharToValue converts a single hex character to its integer value.
func (c *Cipher) hexCharToValue(b byte) int {
	if '0' <= b && b <= '9' {
		return int(b - '0')
	} else if 'a' <= b && b <= 'f' {
		return int(b-'a') + 10
	} else if 'A' <= b && b <= 'F' {
		return int(b-'A') + 10
	}
	return -1
}

// stringToInt64Array converts a hex string to an array of int64 values.
func (c *Cipher) stringToInt64Array(hexStr string) []int64 {
	numChunks := len(hexStr) / 16
	array := make([]int64, numChunks)
	for i := 0; i < numChunks; i++ {
		array[i] = c.parseHexToInt64(hexStr[i*16 : i*16+16])
	}
	return array
}

// int64ArrayToString converts an array of int64 values back to a hex string.
func (c *Cipher) int64ArrayToString(data []int64) string {
	var result []byte
	for _, v := range data {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
		result = append(result, b...)
	}
	return string(c.trimTrailingZeros(result))
}

// padKey ensures the key has a length of 32 bytes.
func (c *Cipher) padKey() []byte {
	if len(c.key) > KeyLength {
		return c.key
	}
	paddedKey := make([]byte, (len(c.key)+KeyLength-1)/KeyLength*KeyLength)
	copy(paddedKey, c.key)
	return paddedKey
}

// padData ensures the plain text data has a length that is a multiple of 8 bytes.
func (c *Cipher) padData(data []byte) []byte {
	if len(data)%8 == 0 {
		return data
	}
	paddedData := make([]byte, (len(data)+7)/8*8)
	copy(paddedData, data)
	return paddedData
}

// bytesToInt64Arr converts a byte slice to an array of int64 values.
func (c *Cipher) bytesToInt64Arr(b []byte) []int64 {
	numElements := (len(b) + 7) / 8
	arr := make([]int64, numElements)
	for i := 0; i < numElements-1; i++ {
		arr[i] = int64(binary.LittleEndian.Uint64(b[i*8:]))
	}
	lastBytes := make([]byte, 8)
	copy(lastBytes, b[(numElements-1)*8:])
	arr[numElements-1] = int64(binary.LittleEndian.Uint64(lastBytes))
	return arr
}

// trimTrailingZeros removes trailing zero bytes from a byte slice.
func (c *Cipher) trimTrailingZeros(data []byte) []byte {
	for len(data) > 0 && data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	return data
}
