package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	after := time.After(5 * time.Second)
	iter := 0
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	// sending message
	for {
		select {
		case <-after:
			log.Println("time out")
			return
		default:
			for i := 0; i < 10; i++ {
				content := "hello[" + strconv.Itoa(iter) + "]"
				// size := byteSize(content)
				data, err := addLengthField([]byte(content))
				if err != nil {
					log.Fatal(err.Error())
				}
				_, err = conn.Write(data)
				if err != nil {
					log.Fatal(err.Error())
				}
				iter++
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// add length field
func addLengthField(content []byte) ([]byte, error) {
	size := len(content)
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, int32(size)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, content); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
