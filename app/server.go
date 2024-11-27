package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	for {
		buf := make ([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
		}

		if n < 2 {
			fmt.Println("Breaking...")
			break
		}

		fmt.Println("read n bytes: ", n)
		fmt.Println("received: ", string(buf[:n]))

		switch buf[0] {
			case byte('*'):
				fmt.Println("received array")
				handleArray(buf, n, conn)
			default:
				fmt.Println("received unsupported request")
		}
	}
}

func handleArray(buf []byte, n int, conn net.Conn) {
	arrayLength, _ := strconv.Atoi(string(buf[1]))
	cmdLength, _ := strconv.Atoi(string(buf[5]))
	command := buf[8:8+cmdLength]

	switch {
		case bytes.EqualFold(command, []byte("ping")):
			fmt.Println("matched ping")
			conn.Write([]byte("+PONG\r\n"))
			if arrayLength > 1 {
				buf = append([]byte{}, buf[0:5]...)
				buf = append(buf, buf[5+cmdLength+2:]...)
				handleArray(buf, n, conn)
			}
		case bytes.EqualFold(command, []byte("echo")):
			fmt.Println("matched echo")
			conn.Write(buf[7+int(cmdLength+3):n])
	}
}
