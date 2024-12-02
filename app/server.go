package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var dir string
var dbfilename string

type Request struct {
	Req []byte
	Cmd []byte
	SubCmd []byte
	Args [][]byte
	Key string
	Value []byte
	ExpirationMilli int64
}

type Value struct {
	Value []byte
	ExpirationMilli int64
}

var store = make(map[string]*Value)

func main() {
	flag.StringVar(&dir, "dir", "/tmp/redis-files", "directory where the rdb snapshot file is located")
	flag.StringVar(&dbfilename, "dbfilename", "dump.rdb", "the name of the rdb file locaterd in dir")
	flag.Parse()

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
	case bytes.EqualFold(req.Cmd, []byte("ping")):
		fmt.Println("matched ping")
		conn.Write([]byte("+PONG\r\n"))
	case bytes.EqualFold(req.Cmd, []byte("echo")):
		fmt.Println("matched echo")
		message := bytes.Join(req.Args, []byte("\r\n"))
		message = append(message, []byte("\r\n")...)
		conn.Write(message)
	case bytes.EqualFold(req.Cmd, []byte("set")):
		fmt.Println("matched set")
		store[req.Key] = &Value{Value: req.Value}
		if req.ExpirationMilli > 0 {
			store[req.Key].ExpirationMilli = req.ExpirationMilli
		}
		conn.Write([]byte("+OK\r\n"))
	case bytes.EqualFold(req.Cmd, []byte("get")):
		fmt.Println("matched get")
		value, ok := store[req.Key]
		if ok {
			if (value.ExpirationMilli != 0 && value.ExpirationMilli < time.Now().UnixMilli()) {
				delete(store, req.Key)
				conn.Write([]byte("$-1\r\n"))
				return
			}
			message := fmt.Sprintf("$%d\r\n%s\r\n", len(value.Value), value.Value)
			conn.Write([]byte(message))
		} else {
			conn.Write([]byte("$-1\r\n"))
		}
	case bytes.EqualFold(req.Cmd, []byte("config")):
		fmt.Println("matched config")
		parameterReq := req.Args[1]
		var parameterResp []byte
		if bytes.EqualFold(parameterReq, []byte("dir")) {
			parameterResp = []byte(dir)
		} else if bytes.EqualFold(parameterReq, []byte("dbfilename")) {
			parameterResp = []byte(dbfilename)
		}
		message := createRespArrayOfBulkStrings([][]byte{parameterReq, parameterResp})
		conn.Write((message))
	}
}

func parseArray(req []byte) Request {
	parts := bytes.Split(req, []byte("\r\n"))
	cmd := parts[2]
	var subCmd []byte
	var key string
	var value []byte
	var args [][]byte
	var expirationMilli int64

	if bytes.EqualFold(cmd, []byte("echo")) {
		args = parts[3:]
		if args[len(args)-1][0] == 0 {
			args = args[:len(args)-1]
		}
	} else if bytes.EqualFold(cmd, []byte("set")) {
		key = string(parts[4])
		value = parts[6]
		if len(parts) >= 10 && bytes.EqualFold(parts[8], []byte("px")) {
			ms, _ := strconv.Atoi(string(parts[10]))
			unixMs := time.Now().UnixMilli()
			expirationMilli = unixMs + int64(ms)
		}
	} else if bytes.EqualFold(cmd, []byte("get")) {
		key = string(parts[4])
	} else if bytes.EqualFold(cmd, []byte("config")) {
		subCmd = parts[4]
		args = parts[5:]
		if args[len(args)-1][0] == 0 {
			args = args[:len(args)-1]
		}
	}

	return Request{
		Req: req,
		Cmd: cmd,
		SubCmd: subCmd,
		Key: key,
		Value: value,
		Args: args,
		ExpirationMilli: expirationMilli,
	}
}

func createRespArrayOfBulkStrings(bulkStrings [][]byte) []byte {
	dataTypeAndCount := fmt.Sprintf("*%d\r\n", len(bulkStrings))
	var stringPairs string
	for i := 0; i < len(bulkStrings); i++ {
		stringPairs += fmt.Sprintf("$%d\r\n%s\r\n", len(bulkStrings[i]), bulkStrings[i])
	}
	resp := dataTypeAndCount + stringPairs

	return []byte(resp)
}
