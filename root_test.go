package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCanWriteToUnixSocket(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/greeting" {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("Hello, world!"))
		}
	}))

	socket := fmt.Sprintf("%s/%d.sock", os.TempDir(), os.Getpid())
	lis, err := net.Listen("unix", socket)

	server.Listener = lis
	go server.Start()

	if err != nil {
		t.Errorf("Could not listen on unix socket: %s\n", err)
		t.FailNow()
	}

	outBuf := &bytes.Buffer{}
	cmd := &Runner{
		Input:  bytes.NewBufferString("GET /greeting HTTP/1.0\r\n\r\n"),
		Output: outBuf,
	}

	cmd.WriteToSocket("unix", socket, "", nil)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(outBuf.Bytes())), nil)
	if err != nil {
		t.Errorf("Could not read response: %s\n", err)
		t.FailNow()
	}

	if resp.StatusCode != http.StatusOK ||
		strings.Contains(outBuf.String(), "Hello, world!") == false {
		t.Errorf("Unexpected output: %s", outBuf.String())
		t.Fail()
	}
}

func TestCanWriteToTCPSocket(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/greeting" {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("Hello, world!"))
		}
	}))

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Errorf("Could not listen on tcp socket: %s\n", err)
		t.FailNow()
	}

	server.Listener = lis
	go server.Start()

	if err != nil {
		t.Errorf("Could not listen on tcp socket: %s\n", err)
		t.FailNow()
	}

	outBuf := &bytes.Buffer{}
	cmd := &Runner{
		Input:  bytes.NewBufferString("GET /greeting HTTP/1.0\r\n\r\n"),
		Output: outBuf,
	}

	cmd.WriteToSocket("tcp", lis.Addr().String(), "", nil)

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(outBuf.Bytes())), nil)
	if err != nil {
		t.Errorf("Could not read response: %s\n", err)
		t.FailNow()
	}

	if resp.StatusCode != http.StatusOK ||
		strings.Contains(outBuf.String(), "Hello, world!") == false {
		t.Errorf("Unexpected output: %s", outBuf.String())
		t.Fail()
	}
}

func TestWriteFrame(t *testing.T) {
	hs := `GET /raw HTTP/1.1
Host: ws.postman-echo.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
Sec-WebSocket-Version: 13`

	full := make([]byte, 0)
	full = append(full, []byte(hs)...)
	full = append(full, []byte("\n\n====")...)
	full = append(full, []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f}...)
	os.WriteFile("hello.bin", full, 0700)
}
