package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

type Request struct {
	req []byte
	cmd []byte
	args [][]byte
	key string
	value []byte
}

var store = make(map[string][]byte)

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
			break
		}

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
	req := parseArray(buf)

	switch {
	case bytes.EqualFold(req.cmd, []byte("ping")):
		fmt.Println("matched ping")
		conn.Write([]byte("+PONG\r\n"))
	case bytes.EqualFold(req.cmd, []byte("echo")):
		fmt.Println("matched echo")
		message := bytes.Join(req.args, []byte("\r\n"))
		message = append(message, []byte("\r\n")...)
		conn.Write(message)
	case bytes.EqualFold(req.cmd, []byte("set")):
		fmt.Println("matched set")
		store[string(req.key)] = req.value
		conn.Write([]byte("+OK\r\n"))
	case bytes.EqualFold(req.cmd, []byte("get")):
		fmt.Println("matched get")
		value, ok := store[string(req.key)]
		if ok {
			message := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
			conn.Write([]byte(message))
		} else {
			conn.Write([]byte("$-1\r\n"))
		}
	}
}

func parseArray(req []byte) Request {
	parts := bytes.Split(req, []byte("\r\n"))
	cmd := parts[2]
	var key string
	var value []byte
	var args [][]byte

	if bytes.EqualFold(cmd, []byte("echo")) {
		args = parts[3:]
		if args[len(args)-1][0] == 0 {
			args = args[:len(args)-1]
		}
	} else if bytes.EqualFold(cmd, []byte("set")) {
		key = string(parts[4])
		value = parts[6]
	} else if bytes.EqualFold(cmd, []byte("get")) {
		key = string(parts[4])
	}

	return Request{
		req: req,
		cmd: cmd,
		key: key,
		value: value,
		args: args,
	}
}
