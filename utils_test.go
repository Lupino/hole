package hole

import (
    "fmt"
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
