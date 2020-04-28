package main

import (
	"log"
	"net"
)

type Message struct {
	MessageType uint8
	Payload uint32
	Timestamp uint32
	StreamID uint32
}

const (
	PORT = ":1935"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	proxy := NewNetworkProxy(conn)
	err := Handshake(proxy)
	if err != nil {
		log.Fatal(err.Error())
	}

	return
}

func main() {
	listen, e := net.Listen("tcp", PORT)
	if e != nil {
		log.Fatal(e.Error())
	}

	for {
		conn, e := listen.Accept()
		if e != nil {
			log.Fatal(e.Error())
		}

		go handleConnection(conn)
	}
}