package hole

import (
	"bytes"
	"fmt"
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
	fmt.Printf("%v\n", header)
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
