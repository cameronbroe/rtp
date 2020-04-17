package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"
)

var h264buffer *bytes.Buffer

func readFileIntoBuffer(path string) (*bytes.Buffer, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(buf), nil
}

const MAX_BUFFER_SIZE = 1024

func main() {
	if len(os.Args) < 2 {
		_ = fmt.Errorf("please give us a file to stream")
		os.Exit(1)
	}

	h264buffer, err := readFileIntoBuffer(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("Length of read buffer: ", h264buffer.Len())

	server, err := net.ListenPacket("udp", "127.0.0.1:16000")
	defer server.Close()
	if err != nil {
		panic(err)
	}

	doneChan := make(chan error, 1)
	buffer := make([]byte, MAX_BUFFER_SIZE)

	go func() {
		for {
			n, addr, err := server.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("packet-received: bytes=%d from %s\n", n, addr.String())

			deadline := time.Now().Add(time.Second)
			err = server.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			for _, b := range h264buffer.Bytes() {
				amount, err := server.WriteTo([]byte{b}, addr)
				if err != nil {
					doneChan <- err
					return
				}
				fmt.Printf("packet-written: bytes=%d to=%s\n", amount, addr.String())
			}
		}
	}()

	select {
	case err = <-doneChan:
		panic(err)
	}
}
