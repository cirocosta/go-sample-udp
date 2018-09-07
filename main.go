package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
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
func server(ctx context.Context, address string) (err error) {
	// ListenPacket provides us a wrapper around ListenUDP so that
	// we don't need to call `net.ResolveUDPAddr` and then subsequentially
	// perform a `ListenUDP` with the UDP address.
	//
	// The returned value (PacketConn) is pretty much the same as the one
	// from ListenUDP (UDPConn) - the only difference is that `Packet*`
	// methods and interfaces are more broad, also covering `ip`.
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return
	}

	// `Close`ing the packet "connection" means cleaning the data structures
	// allocated for holding information about the listening socket.
	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		// By reading from the connection into the buffer we block until there's
		// new content in the socket that we're listening for new packets.
		//
		// Whenever new packets arrive, `buffer` gets filled and we can continue
		// the execution.
		n, addr, err := pc.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("packet-received: bytes=%d from=%s msg=%s\n",
			n, addr.String(), string(buffer[:n]))

		// Q: could this thing ever block?
		n, err = pc.WriteTo(buffer[:n], addr)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("packet-written: bytes=%d to=%s\n", n, addr.String())
	}()

	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return
}

// client wraps the whole functionality of a UDP client that sends
// a message and waits for a response coming back from the server
// that it initially targetted.
func client(ctx context.Context, address string) (err error) {
	// CC: explain why this is resolution is needed.
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	// Q: What is DialUDP really doing?
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}

	defer conn.Close()

	// Q: Is is possible for `write` to block?
	n, err := conn.Write([]byte("hello from the client"))
	if err != nil {
		return
	}

	fmt.Printf("packet-written: bytes=%d\n", n)

	doneChan := make(chan error, 1)

	go func() {
		// Q: How can we make sure that we're reading all that
		//    we want? e.g., what's the best way of making sure
		//    that we were able to consume the whole msg that is
		//    already in the queue?
		buffer := make([]byte, maxBufferSize)

		// Q: How can we set a timeout for this?
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("packet-received: bytes=%d from=%s msg=%s\n",
			n, addr.String(), string(buffer[:n]))

		doneChan <- nil
	}()

	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return
}

func main() {
	flag.Parse()

	var (
		err     error
		address = fmt.Sprintf("%s:%d", *host, *port)
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	if *isServer {
		fmt.Println("running as a server on " + address)
		for {
			err = server(ctx, address)
			if err != nil {
				if err == context.Canceled {
					return
				}

				panic(err)
			}
		}
		return
	}

	fmt.Println("sending to " + address)
	err = client(ctx, address)
	if err != nil && err != context.Canceled {
		panic(err)
	}
}
