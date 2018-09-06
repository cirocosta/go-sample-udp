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

// maxBufferSize specifies the size of the buffers that
// are used to temporarily hold data from the UDP packets
// that we receive.
const maxBufferSize = 1024

// server wraps all the UDP echo server functionality.
// ps.: the server is capable of answering to a single
// client at a time.
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

	fmt.Printf("packet-received: bytes=%d from=%s msg=%s\n",
		n, addr.String(), string(buffer[:n]))

	n, err = pc.WriteTo(buffer[:n], addr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-written: bytes=%d to=%s\n", n, addr.String())
}

// client wraps the whole functionality of a UDP client that sends
// a message and waits for a response coming back from the server
// that it initially targetted.
func client(address string) {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	n, err := conn.Write([]byte("hello from the client"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-written: bytes=%d\n", n)

	buffer := make([]byte, maxBufferSize)
	n, addr, err := conn.ReadFrom(buffer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("packet-received: bytes=%d from=%s msg=%s\n",
		n, addr.String(), string(buffer[:n]))
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
