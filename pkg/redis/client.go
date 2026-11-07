package redis

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// Options holds configuration settings for connecting to a Redis server.
type Options struct {
	Addr         string        // Host and port (e.g., "127.0.0.1:6379")
	Password     string        // Optional auth password
	DB           int           // Target database index (default 0)
	DialTimeout  time.Duration // Network dial timeout (default 5s)
	ReadTimeout  time.Duration // Socket read timeout (default 3s)
	WriteTimeout time.Duration // Socket write timeout (default 3s)
	PoolSize     int           // Max idle connections in pool (default 10)
}

// Client represents a thread-safe Redis client backed by a connection pool.
type Client struct {
	opts    Options
	pool    chan net.Conn
	mu      sync.Mutex
	active  int
	closed  bool
}

// NewClient initializes a new Redis Client with the provided options.
func NewClient(opts Options) *Client {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:6379"
	}
	if opts.DialTimeout <= 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = 3 * time.Second
	}
	if opts.WriteTimeout <= 0 {
		opts.WriteTimeout = 3 * time.Second
	}
	if opts.PoolSize <= 0 {
		opts.PoolSize = 10
	}

	return &Client{
		opts: opts,
		pool: make(chan net.Conn, opts.PoolSize),
	}
}

// NewFromEnv initializes a Redis Client using environment variables.
// Variables: REDIS_HOST (default 127.0.0.1), REDIS_PORT (default 6379),
// REDIS_PASSWORD (optional), REDIS_DB (default 0).
func NewFromEnv() *Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	dbIndex := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if val, err := strconv.Atoi(dbStr); err == nil {
			dbIndex = val
		}
	}

	return NewClient(Options{
		Addr:     net.JoinHostPort(host, port),
		Password: password,
		DB:       dbIndex,
	})
}

// dial establishes and configures a new raw Redis TCP connection.
func (c *Client) dial() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", c.opts.Addr, c.opts.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("redis: connection failed to %s: %w", c.opts.Addr, err)
	}

	// Authenticate if password provided
	if c.opts.Password != "" {
		if err := c.execAuth(conn); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}

	// Select database if not default 0
	if c.opts.DB > 0 {
		if err := c.execSelect(conn); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}

	return conn, nil
}

func (c *Client) execAuth(conn net.Conn) error {
	_ = conn.SetDeadline(time.Now().Add(c.opts.WriteTimeout))
	if _, err := conn.Write(EncodeCommand("AUTH", c.opts.Password)); err != nil {
		return fmt.Errorf("redis: auth write failed: %w", err)
	}
	reader := bufio.NewReader(conn)
	_ = conn.SetDeadline(time.Now().Add(c.opts.ReadTimeout))
	if _, err := DecodeResponse(reader); err != nil {
		return fmt.Errorf("redis: auth failed: %w", err)
	}
	return nil
}

func (c *Client) execSelect(conn net.Conn) error {
	_ = conn.SetDeadline(time.Now().Add(c.opts.WriteTimeout))
	if _, err := conn.Write(EncodeCommand("SELECT", c.opts.DB)); err != nil {
		return fmt.Errorf("redis: select db write failed: %w", err)
	}
	reader := bufio.NewReader(conn)
	_ = conn.SetDeadline(time.Now().Add(c.opts.ReadTimeout))
	if _, err := DecodeResponse(reader); err != nil {
		return fmt.Errorf("redis: select db failed: %w", err)
	}
	return nil
}

// getConn acquires a connection from the pool or creates a new one.
func (c *Client) getConn() (net.Conn, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errors.New("redis: client is closed")
	}

	select {
	case conn := <-c.pool:
		c.mu.Unlock()
		return conn, nil
	default:
		c.active++
		c.mu.Unlock()
		conn, err := c.dial()
		if err != nil {
			c.mu.Lock()
			c.active--
			c.mu.Unlock()
			return nil, err
		}
		return conn, nil
	}
}

// releaseConn returns a healthy connection to the pool or closes it on failure.
func (c *Client) releaseConn(conn net.Conn, err error) {
	if conn == nil {
		return
	}

	if err != nil && !errors.Is(err, ErrNil) {
		// Fatal network or protocol error: close connection
		_ = conn.Close()
		c.mu.Lock()
		c.active--
		c.mu.Unlock()
		return
	}

	c.mu.Lock()
	if c.closed {
		c.active--
		c.mu.Unlock()
		_ = conn.Close()
		return
	}

	select {
	case c.pool <- conn:
		c.mu.Unlock()
	default:
		// Pool full, close extra connection
		c.active--
		c.mu.Unlock()
		_ = conn.Close()
	}
}

// Do executes an arbitrary Redis command with the given arguments and returns the raw decoded reply.
func (c *Client) Do(args ...any) (any, error) {
	if len(args) == 0 {
		return nil, errors.New("redis: empty command")
	}

	conn, err := c.getConn()
	if err != nil {
		return nil, err
	}

	encoded := EncodeCommand(args...)

	// Set write deadline
	if c.opts.WriteTimeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(c.opts.WriteTimeout))
	}
	if _, err := conn.Write(encoded); err != nil {
		c.releaseConn(conn, err)
		return nil, err
	}

	// Set read deadline
	if c.opts.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(c.opts.ReadTimeout))
	}
	reader := bufio.NewReader(conn)
	reply, err := DecodeResponse(reader)
	c.releaseConn(conn, err)
	return reply, err
}

// Ping tests server connectivity.
func (c *Client) Ping(message ...string) (string, error) {
	args := []any{"PING"}
	if len(message) > 0 {
		args = append(args, message[0])
	}
	res, err := c.Do(args...)
	if err != nil {
		return "", err
	}
	if str, ok := res.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", res), nil
}

// Get retrieves the string value of a key. Returns ErrNil if the key does not exist.
func (c *Client) Get(key string) (string, error) {
	res, err := c.Do("GET", key)
	if err != nil {
		return "", err
	}
	if str, ok := res.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", res), nil
}

// Set stores a key-value pair, with an optional TTL expiration (0 = no expiration).
func (c *Client) Set(key string, value any, expiration time.Duration) error {
	args := []any{"SET", key, value}
	if expiration > 0 {
		millis := expiration.Milliseconds()
		if millis > 0 {
			args = append(args, "PX", millis)
		} else {
			args = append(args, "EX", int64(expiration.Seconds()))
		}
	}
	_, err := c.Do(args...)
	return err
}

// SetNX sets key to hold value if key does not exist.
func (c *Client) SetNX(key string, value any, expiration time.Duration) (bool, error) {
	args := []any{"SET", key, value}
	if expiration > 0 {
		args = append(args, "PX", expiration.Milliseconds())
	}
	args = append(args, "NX")
	res, err := c.Do(args...)
	if err != nil {
		if errors.Is(err, ErrNil) {
			return false, nil
		}
		return false, err
	}
	return res != nil, nil
}

// Del removes the specified keys. Returns the number of keys removed.
func (c *Client) Del(keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	args := make([]any, 0, len(keys)+1)
	args = append(args, "DEL")
	for _, k := range keys {
		args = append(args, k)
	}
	res, err := c.Do(args...)
	if err != nil {
		return 0, err
	}
	if val, ok := res.(int64); ok {
		return val, nil
	}
	return 0, nil
}

// Exists checks if key(s) exist in Redis.
func (c *Client) Exists(keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	args := make([]any, 0, len(keys)+1)
	args = append(args, "EXISTS")
	for _, k := range keys {
		args = append(args, k)
	}
	res, err := c.Do(args...)
	if err != nil {
		return 0, err
	}
	if val, ok := res.(int64); ok {
		return val, nil
	}
	return 0, nil
}

// Expire sets a timeout on key.
func (c *Client) Expire(key string, expiration time.Duration) (bool, error) {
	res, err := c.Do("EXPIRE", key, int64(expiration.Seconds()))
	if err != nil {
		return false, err
	}
	if val, ok := res.(int64); ok {
		return val == 1, nil
	}
	return false, nil
}

// TTL returns the remaining time to live of a key.
func (c *Client) TTL(key string) (time.Duration, error) {
	res, err := c.Do("TTL", key)
	if err != nil {
		return 0, err
	}
	sec, ok := res.(int64)
	if !ok {
		return 0, nil
	}
	if sec == -2 {
		return 0, ErrNil // key does not exist
	}
	if sec == -1 {
		return -1, nil // key exists with no expiration
	}
	return time.Duration(sec) * time.Second, nil
}

// FlushDB clears all keys in the currently selected database.
func (c *Client) FlushDB() error {
	_, err := c.Do("FLUSHDB")
	return err
}

// LPush inserts values at the head of the list stored at key.
func (c *Client) LPush(key string, values ...any) (int64, error) {
	if len(values) == 0 {
		return 0, nil
	}
	args := make([]any, 0, len(values)+2)
	args = append(args, "LPUSH", key)
	args = append(args, values...)
	res, err := c.Do(args...)
	if err != nil {
		return 0, err
	}
	if val, ok := res.(int64); ok {
		return val, nil
	}
	return 0, nil
}

// RPush inserts values at the tail of the list stored at key.
func (c *Client) RPush(key string, values ...any) (int64, error) {
	if len(values) == 0 {
		return 0, nil
	}
	args := make([]any, 0, len(values)+2)
	args = append(args, "RPUSH", key)
	args = append(args, values...)
	res, err := c.Do(args...)
	if err != nil {
		return 0, err
	}
	if val, ok := res.(int64); ok {
		return val, nil
	}
	return 0, nil
}

// LPop removes and returns the first element of the list stored at key.
func (c *Client) LPop(key string) (string, error) {
	res, err := c.Do("LPOP", key)
	if err != nil {
		return "", err
	}
	if str, ok := res.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", res), nil
}

// RPop removes and returns the last element of the list stored at key.
func (c *Client) RPop(key string) (string, error) {
	res, err := c.Do("RPOP", key)
	if err != nil {
		return "", err
	}
	if str, ok := res.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", res), nil
}

// BRPop is the blocking equivalent of RPop. It waits up to timeout (in seconds) for an element.
func (c *Client) BRPop(timeout time.Duration, keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, errors.New("redis: BRPop requires at least one key")
	}

	conn, err := c.getConn()
	if err != nil {
		return nil, err
	}

	args := make([]any, 0, len(keys)+2)
	args = append(args, "BRPOP")
	for _, k := range keys {
		args = append(args, k)
	}
	sec := int(timeout.Seconds())
	if sec <= 0 {
		sec = 0
	}
	args = append(args, sec)

	encoded := EncodeCommand(args...)

	// Adjust read deadline for blocking command
	if sec > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(timeout + 2*time.Second))
	} else {
		_ = conn.SetReadDeadline(time.Time{}) // no deadline for infinite wait
	}

	if _, err := conn.Write(encoded); err != nil {
		c.releaseConn(conn, err)
		return nil, err
	}

	reader := bufio.NewReader(conn)
	reply, err := DecodeResponse(reader)
	c.releaseConn(conn, err)
	if err != nil {
		return nil, err
	}

	if arr, ok := reply.([]any); ok {
		results := make([]string, len(arr))
		for i, v := range arr {
			results[i] = fmt.Sprintf("%v", v)
		}
		return results, nil
	}

	return nil, ErrNil
}

// LLen returns the length of the list stored at key.
func (c *Client) LLen(key string) (int64, error) {
	res, err := c.Do("LLEN", key)
	if err != nil {
		return 0, err
	}
	if val, ok := res.(int64); ok {
		return val, nil
	}
	return 0, nil
}

// Publish posts a message to the given channel. Returns the number of clients that received the message.
func (c *Client) Publish(channel string, message any) (int64, error) {
	res, err := c.Do("PUBLISH", channel, message)
	if err != nil {
		return 0, err
	}
	if count, ok := res.(int64); ok {
		return count, nil
	}
	return 0, nil
}

// Close closes all pooled connections and marks client closed.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.pool)
	for conn := range c.pool {
		_ = conn.Close()
	}
	return nil
}

