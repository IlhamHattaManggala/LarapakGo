package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// ErrNil is returned when a key does not exist or a nil bulk reply is received.
var ErrNil = errors.New("redis: nil")

// Protocol markers for Redis RESP
const (
	respSimpleString = '+'
	respError        = '-'
	respInteger      = ':'
	respBulkString   = '$'
	respArray        = '*'
)

// EncodeCommand encodes a command name and its arguments into RESP array format.
func EncodeCommand(args ...any) []byte {
	var buf []byte
	buf = append(buf, respArray)
	buf = append(buf, strconv.Itoa(len(args))...)
	buf = append(buf, '\r', '\n')

	for _, arg := range args {
		str := fmt.Sprintf("%v", arg)
		buf = append(buf, respBulkString)
		buf = append(buf, strconv.Itoa(len(str))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, str...)
		buf = append(buf, '\r', '\n')
	}

	return buf
}

// DecodeResponse reads a single RESP response from a buffered reader.
func DecodeResponse(reader *bufio.Reader) (any, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 3 || line[len(line)-2] != '\r' {
		return nil, errors.New("redis: protocol error, invalid line terminator")
	}

	prefix := line[0]
	payload := string(line[1 : len(line)-2])

	switch prefix {
	case respSimpleString:
		return payload, nil

	case respError:
		return nil, errors.New("redis error: " + payload)

	case respInteger:
		val, err := strconv.ParseInt(payload, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("redis: invalid integer %q: %w", payload, err)
		}
		return val, nil

	case respBulkString:
		length, err := strconv.Atoi(payload)
		if err != nil {
			return nil, fmt.Errorf("redis: invalid bulk string length %q: %w", payload, err)
		}
		if length == -1 {
			return nil, ErrNil
		}

		data := make([]byte, length)
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, err
		}

		// Consume the trailing \r\n
		cr, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		lf, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if cr != '\r' || lf != '\n' {
			return nil, errors.New("redis: protocol error, expected CRLF after bulk data")
		}

		return string(data), nil

	case respArray:
		count, err := strconv.Atoi(payload)
		if err != nil {
			return nil, fmt.Errorf("redis: invalid array length %q: %w", payload, err)
		}
		if count == -1 {
			return nil, ErrNil
		}

		items := make([]any, count)
		for i := 0; i < count; i++ {
			item, err := DecodeResponse(reader)
			if err != nil {
				if errors.Is(err, ErrNil) {
					items[i] = nil
					continue
				}
				return nil, err
			}
			items[i] = item
		}
		return items, nil

	default:
		return nil, fmt.Errorf("redis: unknown protocol prefix %q", prefix)
	}
}

