package queue

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/IlhamHattaManggala/larapak-go/pkg/redis"
)

func TestMemoryQueue(t *testing.T) {
	q := NewMemoryQueue()

	executed := false
	job := &FunctionJob{
		Name: "test-job",
		Fn: func() error {
			executed = true
			return nil
		},
	}

	if err := q.Push(job); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	popped, ok := q.Pop()
	if !ok || popped == nil {
		t.Fatalf("expected job from pop")
	}

	if err := popped.Handle(); err != nil || !executed {
		t.Fatalf("job execution failed: %v", err)
	}
}

// startMockRedisQueueServer creates an in-memory mock Redis server for queue tests.
func startMockRedisQueueServer(t *testing.T) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	var mu sync.Mutex
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
					case "LPUSH":
						if len(args) >= 3 {
							key := fmt.Sprintf("%v", args[1])
							mu.Lock()
							for i := 2; i < len(args); i++ {
								lists[key] = append([]string{fmt.Sprintf("%v", args[i])}, lists[key]...)
							}
							count := len(lists[key])
							mu.Unlock()
							_, _ = c.Write([]byte(fmt.Sprintf(":%d\r\n", count)))
						}

					case "RPOP":
						if len(args) >= 2 {
							key := fmt.Sprintf("%v", args[1])
							mu.Lock()
							list := lists[key]
							if len(list) == 0 {
								mu.Unlock()
								_, _ = c.Write([]byte("$-1\r\n"))
							} else {
								val := list[len(list)-1]
								lists[key] = list[:len(list)-1]
								mu.Unlock()
								_, _ = c.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
							}
						}

					case "BRPOP":
						if len(args) >= 3 {
							key := fmt.Sprintf("%v", args[1])
							mu.Lock()
							list := lists[key]
							if len(list) > 0 {
								val := list[len(list)-1]
								lists[key] = list[:len(list)-1]
								mu.Unlock()

								var b bytes.Buffer
								b.WriteString("*2\r\n")
								b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(key), key))
								b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
								_, _ = c.Write(b.Bytes())
							} else {
								mu.Unlock()
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

	return ln.Addr().String(), func() {
		close(done)
		_ = ln.Close()
	}
}

type emailJob struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Executed  bool   `json:"executed"`
}

func (e *emailJob) JobName() string {
	return "send-email"
}

func (e *emailJob) Handle() error {
	e.Executed = true
	return nil
}

func TestRedisQueue(t *testing.T) {
	addr, cleanup := startMockRedisQueueServer(t)
	defer cleanup()

	RegisterJob("send-email", func() Job {
		return &emailJob{}
	})

	client := redis.NewClient(redis.Options{Addr: addr})
	defer client.Close()

	rq := NewRedisQueue(client)

	job := &emailJob{
		Recipient: "bocah@tegal.dev",
		Subject:   "Halo Larapak!",
	}

	if err := rq.Push(job); err != nil {
		t.Fatalf("redis push failed: %v", err)
	}

	popped, ok := rq.Pop()
	if !ok || popped == nil {
		t.Fatalf("expected job from redis queue")
	}

	ej, ok := popped.(*emailJob)
	if !ok || ej.Recipient != "bocah@tegal.dev" {
		t.Fatalf("unexpected unmarshaled job: %+v", popped)
	}

	if err := ej.Handle(); err != nil || !ej.Executed {
		t.Fatalf("job execution failed")
	}
}
