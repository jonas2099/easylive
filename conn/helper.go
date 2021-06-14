package conn

import (
	"github.com/Monibuca/engine/v2/util/bits/pio"
	"io"
)

func (c *Conn) getRealCSID(csid uint32) (uint32, error) {
	var data []byte
	switch csid {
	default: // Chunk basic header 1
	case 0: // Chunk basic header 2
		if _, err := io.ReadFull(c.bufReader, data[:1]); err != nil {
			return 0, err
		}
		return uint32(data[0]) + 64, nil
	case 1: // Chunk basic header 3
		if _, err := io.ReadFull(c.bufReader, data[:2]); err != nil {
			return 0, err
		}
		return uint32(pio.U16BE(data)) + 64, nil
	}
	return csid, nil
}

func (c *Conn) WriteUintBE(v uint32, n int) error {
	for i := 0; i < n; i++ {
		b := byte(v>>uint32((n-i-1)<<3)) & 0xff
		if err := c.bufWriter.WriteByte(b); err != nil {
			return err
		}
	}
	return nil
}

func (c *Conn) WriteUintLE(v uint32, n int) error {
	for i := 0; i < n; i++ {
		b := byte(v) & 0xff
		if err := c.bufWriter.WriteByte(b); err != nil {
			return err
		}
		v = v >> 8
	}
	return nil
}
