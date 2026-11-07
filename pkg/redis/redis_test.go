package redis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRESPEncoding(t *testing.T) {
	encoded := EncodeCommand("SET", "kunci", "nilai", "EX", 60)
	expected := "*5\r\n$3\r\nSET\r\n$5\r\nkunci\r\n$5\r\nnilai\r\n$2\r\nEX\r\n$2\r\n60\r\n"
	if string(encoded) != expected {
		t.Fatalf("unexpected encoding:\nGot:  %q\nWant: %q", string(encoded), expected)
	}
}

func TestRESPDecoding(t *testing.T) {
	// Simple string
	r := bufio.NewReader(strings.NewReader("+PONG\r\n"))
	val, err := DecodeResponse(r)
	if err != nil || val != "PONG" {
		t.Fatalf("simple string failed: val=%v, err=%v", val, err)
	}

	// Error
	r = bufio.NewReader(strings.NewReader("-ERR unknown command\r\n"))
	_, err = DecodeResponse(r)
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("error decode failed: %v", err)
	}

	// Integer
	r = bufio.NewReader(strings.NewReader(":1024\r\n"))
	val, err = DecodeResponse(r)
	if err != nil || val != int64(1024) {
		t.Fatalf("integer decode failed: val=%v, err=%v", val, err)
	}

	// Bulk string
	r = bufio.NewReader(strings.NewReader("$5\r\nhello\r\n"))
	val, err = DecodeResponse(r)
	if err != nil || val != "hello" {
		t.Fatalf("bulk string decode failed: val=%v, err=%v", val, err)
	}

	// Nil bulk string
	r = bufio.NewReader(strings.NewReader("$-1\r\n"))
	_, err = DecodeResponse(r)
	if !errors.Is(err, ErrNil) {
		t.Fatalf("expected ErrNil, got: %v", err)
	}

	// Array
	r = bufio.NewReader(strings.NewReader("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"))
	val, err = DecodeResponse(r)
	if err != nil {
		t.Fatalf("array decode failed: %v", err)
	}
	arr, ok := val.([]any)
	if !ok || len(arr) != 2 || arr[0] != "foo" || arr[1] != "bar" {
		t.Fatalf("unexpected array: %v", val)
	}
}

// startMockRedis creates an in-memory TCP server that speaks basic RESP commands for unit testing.
func startMockRedis(t *testing.T) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	store := make(map[string]string)
	lists := make(map[string][]string)

	done := make(chan struct{})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}

			go func(c net.Conn) {
				defer c.Close()
				reader := bufio.NewReader(c)

				for {
					reply, err := DecodeResponse(reader)
					if err != nil {
						return
					}

					args, ok := reply.([]any)
					if !ok || len(args) == 0 {
						return
					}

					cmd := strings.ToUpper(fmt.Sprintf("%v", args[0]))
					switch cmd {
					case "PING":
						if len(args) > 1 {
							_, _ = c.Write([]byte(fmt.Sprintf("$%d\r\n%v\r\n", len(fmt.Sprintf("%v", args[1])), args[1])))
						} else {
							_, _ = c.Write([]byte("+PONG\r\n"))
						}

					case "SET":
						if len(args) >= 3 {
							key := fmt.Sprintf("%v", args[1])
							val := fmt.Sprintf("%v", args[2])
							store[key] = val
							_, _ = c.Write([]byte("+OK\r\n"))
						}

					case "GET":
						if len(args) >= 2 {
							key := fmt.Sprintf("%v", args[1])
							if val, ok := store[key]; ok {
								_, _ = c.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
							} else {
								_, _ = c.Write([]byte("$-1\r\n"))
							}
						}

					case "DEL":
						count := 0
						for i := 1; i < len(args); i++ {
							key := fmt.Sprintf("%v", args[i])
							if _, ok := store[key]; ok {
								delete(store, key)
								count++
							}
							if _, ok := lists[key]; ok {
								delete(lists, key)
								count++
							}
						}
						_, _ = c.Write([]byte(fmt.Sprintf(":%d\r\n", count)))

					case "EXISTS":
						count := 0
						for i := 1; i < len(args); i++ {
							key := fmt.Sprintf("%v", args[i])
							if _, ok := store[key]; ok {
								count++
							}
						}
						_, _ = c.Write([]byte(fmt.Sprintf(":%d\r\n", count)))

					case "FLUSHDB":
						store = make(map[string]string)
						lists = make(map[string][]string)
						_, _ = c.Write([]byte("+OK\r\n"))

					case "LPUSH":
						if len(args) >= 3 {
							key := fmt.Sprintf("%v", args[1])
							for i := 2; i < len(args); i++ {
								lists[key] = append([]string{fmt.Sprintf("%v", args[i])}, lists[key]...)
							}
							_, _ = c.Write([]byte(fmt.Sprintf(":%d\r\n", len(lists[key]))))
						}

					case "RPOP":
						if len(args) >= 2 {
							key := fmt.Sprintf("%v", args[1])
							list := lists[key]
							if len(list) == 0 {
								_, _ = c.Write([]byte("$-1\r\n"))
							} else {
								val := list[len(list)-1]
								lists[key] = list[:len(list)-1]
								_, _ = c.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
							}
						}

					case "BRPOP":
						if len(args) >= 3 {
							key := fmt.Sprintf("%v", args[1])
							list := lists[key]
							if len(list) > 0 {
								val := list[len(list)-1]
								lists[key] = list[:len(list)-1]
								var b bytes.Buffer
								b.WriteString("*2\r\n")
								b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(key), key))
								b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
								_, _ = c.Write(b.Bytes())
							} else {
								_, _ = c.Write([]byte("*-1\r\n"))
							}
						}

					default:
						_, _ = c.Write([]byte("-ERR unknown command\r\n"))
					}
				}
			}(conn)
		}
	}()

	cleanup := func() {
		close(done)
		_ = ln.Close()
	}

	return ln.Addr().String(), cleanup
}

func TestRedisClientMockServer(t *testing.T) {
	addr, cleanup := startMockRedis(t)
	defer cleanup()

	client := NewClient(Options{
		Addr: addr,
	})
	defer client.Close()

	// 1. PING
	pong, err := client.Ping()
	if err != nil || pong != "PONG" {
		t.Fatalf("ping failed: pong=%q, err=%v", pong, err)
	}

	// 2. SET & GET
	err = client.Set("sapaan", "NgapakIn", 10*time.Second)
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}

	val, err := client.Get("sapaan")
	if err != nil || val != "NgapakIn" {
		t.Fatalf("get failed: val=%q, err=%v", val, err)
	}

	// 3. GET Non-existent key
	_, err = client.Get("ora_ana")
	if !errors.Is(err, ErrNil) {
		t.Fatalf("expected ErrNil, got: %v", err)
	}

	// 4. DEL
	deleted, err := client.Del("sapaan")
	if err != nil || deleted != 1 {
		t.Fatalf("del failed: deleted=%d, err=%v", deleted, err)
	}

	// 5. LPUSH & RPOP
	lenVal, err := client.LPush("antrean", "gawean1", "gawean2")
	if err != nil || lenVal != 2 {
		t.Fatalf("lpush failed: len=%d, err=%v", lenVal, err)
	}

	popped, err := client.RPop("antrean")
	if err != nil || popped != "gawean1" {
		t.Fatalf("rpop failed: popped=%q, err=%v", popped, err)
	}

	// 6. BRPOP
	brpop, err := client.BRPop(1*time.Second, "antrean")
	if err != nil || len(brpop) != 2 || brpop[1] != "gawean2" {
		t.Fatalf("brpop failed: brpop=%v, err=%v", brpop, err)
	}

	// 7. FLUSHDB
	err = client.FlushDB()
	if err != nil {
		t.Fatalf("flushdb failed: %v", err)
	}
}

