package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

		s := bufio.NewScanner(conn)
		for s.Scan() {
			line := s.Text()
			fmt.Println("received: ", line)

			if line == "PING" {
				conn.Write([]byte("+PONG\r\n"))
			}
		}
	}
}
