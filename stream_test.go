package hole

import (
	"bytes"
	"io"
	"testing"
)

func TestReadStream(t *testing.T) {
	rs := NewReadStream()
	var buffer = make([]byte, 1024)
	var data = []byte("This is a payload.")
	rs.FeedData(data)
	var n int
	var err error
	if n, err = rs.Read(buffer); err != nil {
		t.Fatalf("Read Error: %s", err)
	}

	if n != len(data) {
		t.Fatalf("Payload length not match.")
	}

	if !bytes.Equal(buffer[:n], data) {
		t.Fatalf("Payload: except: %x, got: %x", data, buffer[:n])
	}

	buffer = make([]byte, 4)
	rs.FeedData(data)

	if n, err = rs.Read(buffer); err != nil {
		t.Fatalf("Read Error: %s", err)
	}

	if n != len(buffer) {
		t.Fatalf("Payload length: expect %d, got: %d.", len(buffer), n)
	}

	if !bytes.Equal(buffer, data[:n]) {
		t.Fatalf("Payload: except: %x, got: %x", data[:n], buffer)
	}

	rs.FeedEOF()
	if n, err = rs.Read(buffer); err == nil {
		t.Fatalf("Read: except %s, got: nil", io.EOF)
	}
}
