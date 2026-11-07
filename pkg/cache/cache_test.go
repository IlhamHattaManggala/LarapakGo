package cache

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/redis"
)

func TestMemoryCache(t *testing.T) {
	c := NewMemoryCache()

	c.Set("kunci", "isi", 100*time.Millisecond)
	val, ok := c.Get("kunci")
	if !ok || val != "isi" {
		t.Fatalf("expected 'isi', got %v", val)
	}

	time.Sleep(120 * time.Millisecond)
	_, ok = c.Get("kunci")
	if ok {
		t.Fatalf("expected expired item, but found")
	}

	c.Set("tetep", "selawase", 0)
	c.Delete("tetep")
	_, ok = c.Get("tetep")
	if ok {
		t.Fatalf("expected deleted item, but found")
	}
}

// startMockRedisCacheServer spins up a minimal mock Redis server for cache testing.
func startMockRedisCacheServer(t *testing.T) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	store := make(map[string]string)
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
					reply, err := redis.DecodeResponse(reader)
					if err != nil {
						return
					}

					args, ok := reply.([]any)
					if !ok || len(args) == 0 {
						return
					}

					cmd := strings.ToUpper(fmt.Sprintf("%v", args[0]))
					switch cmd {
					case "SET":
						key := fmt.Sprintf("%v", args[1])
						val := fmt.Sprintf("%v", args[2])
						store[key] = val
						_, _ = c.Write([]byte("+OK\r\n"))

					case "GET":
						key := fmt.Sprintf("%v", args[1])
						if val, ok := store[key]; ok {
							_, _ = c.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
						} else {
							_, _ = c.Write([]byte("$-1\r\n"))
						}

					case "DEL":
						count := 0
						for i := 1; i < len(args); i++ {
							key := fmt.Sprintf("%v", args[i])
							if _, ok := store[key]; ok {
								delete(store, key)
								count++
							}
						}
						_, _ = c.Write([]byte(fmt.Sprintf(":%d\r\n", count)))

					case "FLUSHDB":
						store = make(map[string]string)
						_, _ = c.Write([]byte("+OK\r\n"))
					}
				}
			}(conn)
		}
	}()

	return ln.Addr().String(), func() {
		close(done)
		_ = ln.Close()
	}
}

func TestRedisCache(t *testing.T) {
	addr, cleanup := startMockRedisCacheServer(t)
	defer cleanup()

	client := redis.NewClient(redis.Options{Addr: addr})
	defer client.Close()

	rc := NewRedisCache(client)

	rc.Set("tokoh", "Sadewa", 5*time.Second)

	val, ok := rc.Get("tokoh")
	if !ok || val != "Sadewa" {
		t.Fatalf("expected 'Sadewa', got %v", val)
	}

	rc.Delete("tokoh")
	_, ok = rc.Get("tokoh")
	if ok {
		t.Fatalf("expected deleted item, but found")
	}

	rc.Set("k1", "v1", 0)
	rc.Set("k2", "v2", 0)
	rc.Flush()

	_, ok = rc.Get("k1")
	if ok {
		t.Fatalf("expected flush to clear k1")
	}
}

func TestCacheDefaultEnv(t *testing.T) {
	_ = os.Setenv("CACHE_DRIVER", "memory")
	store := Default()
	if _, ok := store.(*MemoryCache); !ok {
		t.Fatalf("expected MemoryCache instance")
	}
}

