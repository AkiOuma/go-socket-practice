package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("listeing from 0.0.0.0:8080")

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		go handler(conn)
	}
}

func handler(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	for {
		peek, err := reader.Peek(4)
		if err != nil {
			if err != io.EOF {
				log.Println(err.Error())
				break
			} else {
				log.Println("ending.")
			}
			break
		}
		buffer := bytes.NewBuffer(peek)
		var size int32
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			log.Println(err)
		}
		if int32(reader.Buffered()) < size+4 {
			continue
		}
		data := make([]byte, size+4)
		if _, err := reader.Read(data); err != nil {
			log.Println(err.Error())
			continue
		}
		log.Printf("recevie : %v", string(data[4:]))
	}
}
