package tus

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
)

// StreamingChecksumReader 流式校验和验证器
type StreamingChecksumReader struct {
	reader      io.Reader
	hasher      hash.Hash
	expectedSum string
	algorithm   string
	validated   bool
	eof         bool
}

func NewStreamingChecksumReader(reader io.Reader, algorithm, expectedChecksum string) (*StreamingChecksumReader, error) {
	var hasher hash.Hash
	switch algorithm {
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	case "md5":
		hasher = md5.New()
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return &StreamingChecksumReader{
		reader:      reader,
		hasher:      hasher,
		expectedSum: expectedChecksum,
		algorithm:   algorithm,
	}, nil
}

func (scr *StreamingChecksumReader) Read(p []byte) (int, error) {
	n, err := scr.reader.Read(p)
	if n > 0 {
		// 边读边计算哈希
		scr.hasher.Write(p[:n])
	}

	// 当读取完成时验证校验和
	if err == io.EOF && !scr.validated {
		scr.eof = true
		if validateErr := scr.validateChecksum(); validateErr != nil {
			return n, validateErr // 返回校验和错误
		}
		scr.validated = true
	}

	return n, err
}

func (scr *StreamingChecksumReader) validateChecksum() error {
	calculatedSum := base64.StdEncoding.EncodeToString(scr.hasher.Sum(nil))
	if calculatedSum != scr.expectedSum {
		return fmt.Errorf("checksum verification failed: expected %s, got %s",
			scr.expectedSum, calculatedSum)
	}
	return nil
}
