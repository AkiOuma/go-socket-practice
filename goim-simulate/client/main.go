package main

import (
	"log"
	"net"
	"socket-practice/goim-simulate/pkg"
	"strconv"
	"time"
)

func main() {
	seq := 0
	version := 1
	code := 1
	timeout := time.After(5 * time.Second)
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()
	for {
		select {
		case <-timeout:
			log.Println("timeout")
			return
		default:
			for i := 0; i < 5; i++ {
				data := pkg.Encoder(pkg.NewPack(version, code, seq, []byte("hello"+strconv.Itoa(seq))))
				seq++
				conn.Write(data)
			}
			time.Sleep(time.Second)
		}
	}
}
