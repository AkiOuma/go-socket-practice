package pkg

import (
	"log"
	"testing"
)

func TestDecoder(t *testing.T) {
	content := "hello world"
	version := 1
	operation := 1
	sequence := 0
	pack := NewPack(version, operation, sequence, []byte(content))

	request := Encoder(pack)
	res, err := Decoder(request)

	// validate decode result
	if err != nil ||
		pack.Length != headerSize()+len(content) ||
		pack.HeaderLength != headerSize() ||
		pack.ProtocolVersion != version ||
		pack.OperationCode != operation ||
		pack.Seq != sequence ||
		string(pack.Content) != content {
		t.Error("not pass")
	}
	log.Printf("%#v", res)
}
