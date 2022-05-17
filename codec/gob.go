package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(header *Header) error {
	return g.dec.Decode(header)
}

func (g *GobCodec) ReadBody(i interface{}) error {
	return g.dec.Decode(i)
}

func (g *GobCodec) Write(h *Header, b interface{}) error {
	defer func() {
		err := g.buf.Flush()
		if err != nil {
			g.Close()
		}
	}()

	err := g.enc.Encode(h)
	if err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}

	err = g.enc.Encode(b)
	if err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}

	return nil
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)

	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
