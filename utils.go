package hole

import (
    "io"
    "fmt"
    "bytes"
    "errors"
)

// Split the message payload
var NULL_CHAR = []byte("\x00\x01")

// Framing:
// In order to handle framing in Send/Recieve, as these give frame
// boundaries we use a very simple 4 bytes header.
func MakeHeader(data []byte) ([]byte, error) {
    header := make([]byte, 4)

    length := uint32(len(data))

    if length > 0x7fffffff {
        return nil, errors.New("Data to large")
    }

    header[0] = byte((length >> 24) & 0xff)
    header[1] = byte((length >> 16) & 0xff)
    header[2] = byte((length >> 8) & 0xff)
    header[3] = byte((length >> 0) & 0xff)

    return header, nil
}


func ParseHeader(header []byte) (uint32) {
    length := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
    length = length & ^uint32(0x80000000)

    return length
}

func EncodePacket(sessionId, data []byte) []byte {
    var buffer = bytes.NewBuffer(nil)
    buffer.Write(sessionId)
    buffer.Write(NULL_CHAR)
    buffer.Write(data)
    return buffer.Bytes()
}

func DecodePacket(payload []byte) (sessionId, data []byte) {
    parts := bytes.SplitN(payload, NULL_CHAR, 2)
    var err = fmt.Sprintf("InvalId %v\n", payload)
    if len(parts) == 1 {
        panic(err)
    }
    return parts[0], parts[1]
}

func PipeThenClose(src io.Reader, dst io.WriteCloser) {
    var buf = make([]byte, 1024)
    defer dst.Close()
    for {
        n, err := src.Read(buf)
        if n > 0 {
            if _, err = dst.Write(buf[:n]); err != nil {
                break
            }
        }
        if err != nil {
            break
        }
    }
}
