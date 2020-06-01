package main

import (
	"log"
	"net"
)

const (
	PORT = ":1935"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	proxy := NewNetworkProxy(conn)
	err := Handshake(proxy)
	if err != nil {
		log.Fatal(err)
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