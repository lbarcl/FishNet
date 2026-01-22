package tcp

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
)

var writerPool = sync.Pool{
	New: func() interface{} {
		// Initialize with io.Discard as a placeholder
		return gzip.NewWriter(io.Discard)
	},
}

var readerPool = sync.Pool{
	New: func() interface{} {
		// Initialize with an empty reader placeholder
		return new(gzip.Reader)
	},
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func gunzipFrame(in []byte, maxOut uint32) ([]byte, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("empty gzip frame")
	}

	reader := readerPool.Get().(*gzip.Reader)
	defer readerPool.Put(reader)

	if err := reader.Reset(bytes.NewReader(in)); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if _, err := out.ReadFrom(io.LimitReader(reader, int64(maxOut)+1)); err != nil {
		return nil, err
	}

	if uint32(out.Len()) > maxOut {
		return nil, fmt.Errorf("gunzip overflow: decompressed=%d max=%d", out.Len(), maxOut)
	}

	return out.Bytes(), nil
}

func gzipFrame(in []byte, maxOut uint32) ([]byte, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("empty gzip frame")
	}

	writer := writerPool.Get().(*gzip.Writer)
	defer writerPool.Put(writer)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	writer.Reset(buf)
	if _, err := writer.Write(in); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	if uint32(buf.Len()) > maxOut {
		return nil, fmt.Errorf("gzip overflow: compressed=%d max=%d", buf.Len(), maxOut)
	}

	return buf.Bytes(), nil
}

func newUID() (string, error) {
	b := make([]byte, 16) // 128-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hasFlag(flags, f FrameFlags) bool { return flags&f != 0 }
