package tcp

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
)

func gunzipFrame(in []byte, maxOut uint32) ([]byte, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("empty gzip frame")
	}

	r, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var out bytes.Buffer
	if _, err := out.ReadFrom(io.LimitReader(r, int64(maxOut)+1)); err != nil {
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

	var out bytes.Buffer
	w := gzip.NewWriter(&out)
	defer w.Close()

	if _, err := w.Write(in); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	if uint32(out.Len()) > maxOut {
		return nil, fmt.Errorf("gzip overflow: compressed=%d max=%d", out.Len(), maxOut)
	}
	return out.Bytes(), nil
}

func newUID() (string, error) {
	b := make([]byte, 16) // 128-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hasFlag(flags, f FrameFlags) bool { return flags&f != 0 }
