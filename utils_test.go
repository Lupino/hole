package hole

import (
	"bytes"
	"github.com/satori/go.uuid"
	"testing"
)

func TestHeader(t *testing.T) {
	var data = []byte("data")
	var length = uint32(len(data))
	var header, err = MakeHeader(data)
	if err != nil {
		t.Fatal(err)
	}
	var lengthGot = ParseHeader(header)

	if lengthGot != length {
		t.Fatalf("Header: except: %d, got: %d", length, lengthGot)
	}
}

func TestEncodeAndDecodePacket(t *testing.T) {
	sessionId := uuid.NewV4().Bytes()
	data := []byte("This is a payload.")

	payload := EncodePacket(sessionId, data)

	sid, d := DecodePacket(payload)

	if !bytes.Equal(sid, sessionId) {
		t.Fatalf("SessionId: except: %x, got: %x", sessionId, sid)
	}

	if !bytes.Equal(d, data) {
		t.Fatalf("Payload: except: %x, got: %x", data, d)
	}
}

type writer struct {
	t      *testing.T
	except []byte
}

func (w *writer) Write(buf []byte) (n int, err error) {
	if !bytes.Equal(w.except, buf) {
		w.t.Fatalf("Payload: except: %x, got: %x", w.except, buf)
	}
	return len(buf), nil
}

func (w *writer) Close() error {
	return nil
}

func TestPipeThenClose(t *testing.T) {
	r := NewReadStream()
	data := []byte("This is a payload.")

	w := new(writer)
	w.t = t
	w.except = data

	go PipeThenClose(r, w)

	r.FeedData(data)
	r.FeedEOF()
}
