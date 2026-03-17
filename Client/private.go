package Client

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/lbarcl/fishnet-go/repo"
	"github.com/valyala/bytebufferpool"
)

func (c *Client) handleFrame(flags repo.FrameFlags, payload *bytebufferpool.ByteBuffer) {
	defer c.bufferPool.Put(payload)

}

func (c *Client) sendFrame(flags repo.FrameFlags, payload []byte) error {
	if len(payload) > int(c.settings.MaxFrameBytes) {
		return fmt.Errorf("payload size exceeds maximum allowed: %d", len(payload))
	}

	frame := make([]byte, 5+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	frame[4] = byte(flags)

	copy(frame[5:], payload)

	if _, err := c.conn.Write(frame); err != nil {
		return err
	}

	return nil
}

func (c *Client) listen() {
	headerBuf := make([]byte, 5)

	for {
		if _, err := io.ReadFull(c.conn, headerBuf); err != nil {
			if c.ctx.Err() == nil {
				c.onErrorFunc(err)
			}
			return
		}

		payloadSize := binary.BigEndian.Uint32(headerBuf[:4])
		if payloadSize > c.settings.MaxFrameBytes {
			c.onErrorFunc(fmt.Errorf("payload size %d exceeds max", payloadSize))
			return
		}

		flags := repo.FrameFlags(headerBuf[4])
		payload := c.bufferPool.Get() // Consider sync.Pool for large payloads
		if _, err := io.CopyN(payload, c.conn, int64(payloadSize)); err != nil {
			if c.ctx.Err() == nil {
				c.onErrorFunc(err)
			}
			return
		}

		c.handleFrame(flags, payload)
	}
}
