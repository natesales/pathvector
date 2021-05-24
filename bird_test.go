package main

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"net"
	"testing"
)

//func init() {
//	log.Infof("Starting listener")
//	l, err := net.Listen("unix", "/tmp/wireframe-bird-test.sock")
//	if err != nil {
//		panic(err)
//	}
//	defer l.Close()
//
//	for {
//		conn, err := l.Accept()
//		if err != nil {
//			panic(err)
//		}
//
//		go func(c net.Conn) {
//			io.Copy(c, c)
//			c.Close()
//		}(conn)
//	}
//}
//
//func TestBirdConfig(t *testing.T) {
//	if err := runBirdCommand("foo", "/tmp/wireframe-bird-test.sock"); err != nil {
//		t.Error(err)
//	}
//}

func TestConn(t *testing.T) {
	go func() {
		if err := runBirdCommand("bird command test\n", "test.sock"); err != nil {
			t.Error(err)
		}
	}()

	log.Println("Starting fake BIRD socket server")
	l, err := net.Listen("unix", "test.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		if _, err := conn.Write([]byte("0001 Hello World")); err != nil {
			t.Error(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf[:])
		if err != nil {
			t.Error(err)
		}
		if string(buf[:n]) != "bird command test\n" {
			t.Errorf("expected 'bird command test' got %s", string(buf[:n]))
		}

		if _, err := conn.Write(bytes.Repeat([]byte("A"), 2048)); err != nil {
			t.Error(err)
		}

		return
	}
}
