package main

import (
	"flag"
	"fmt"
	"net"
)

var (
	isServer = flag.Bool("server", false, "whether it should be run as a server")
	port     = flag.Uint("port", 1337, "port to send to or receive from")
	host     = flag.String("host", "127.0.0.1", "address to send to or receive from")
)

const maxBufferSize = 1024

// TODO - make this thing an echo server
func server(address string) {
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		panic(err)
	}

	defer pc.Close()

	buffer := make([]byte, maxBufferSize)

	n, addr, err := pc.ReadFrom(buffer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-received: bytes=%d from=%s\n", n, addr.String())

	n, err = pc.WriteTo(buffer[:n], addr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-written: bytes=%d to=%s\n", n, addr.String())
}

func client(address string) {
	// when would this thing fail?
	conn, err := net.Dial("udp", address)
	if err != nil {
		panic(err)
	}

	// why do we need to close?
	defer conn.Close()

	n, err := conn.Write([]byte("hello from the client"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-written: bytes=%d\n", n)

	buffer := make([]byte, maxBufferSize)
	n, err = conn.Read(buffer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-received: bytes=%d\n", n)
}

func main() {
	flag.Parse()

	address := fmt.Sprintf("%s:%d", *host, *port)

	if *isServer {
		fmt.Println("running as a server on " + address)
		server(address)
		return
	}

	fmt.Println("sending to " + address)
	client(address)
}
