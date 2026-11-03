package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// RedisClient defines the minimal interface for writing events to Redis (allows mocking and real client).
type RedisClient interface {
	Do(ctx context.Context, cmd string, args ...any) error
	Close() error
}

// MockRedisClient stores Redis commands in memory for testing.
type MockRedisClient struct {
	Commands []string
	Records  map[string]string
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		Commands: make([]string, 0),
		Records:  make(map[string]string),
	}
}

func (m *MockRedisClient) Do(ctx context.Context, cmd string, args ...any) error {
	line := fmt.Sprintf("%s %v", cmd, args)
	m.Commands = append(m.Commands, line)
	if strings.ToUpper(cmd) == "SET" && len(args) >= 2 {
		key, _ := args[0].(string)
		val, _ := args[1].(string)
		m.Records[key] = val
	}
	return nil
}

func (m *MockRedisClient) Close() error {
	return nil
}

// RedisSink writes generated synthetic events to Redis keys, lists, pubsub channels, or streams.
type RedisSink struct {
	cfg    *pb.RedisSink
	client RedisClient
}

func NewRedisSink(cfg *pb.RedisSink, client RedisClient) *RedisSink {
	if client == nil {
		client = NewMockRedisClient()
	}
	return &RedisSink{
		cfg:    cfg,
		client: client,
	}
}

func (s *RedisSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload for redis: %w", err)
	}

	mode := strings.ToLower(s.cfg.GetMode())
	if mode == "" {
		mode = "set"
	}

	keyStr := string(key)
	if s.cfg.GetKeyPattern() != "" {
		keyStr = strings.ReplaceAll(s.cfg.GetKeyPattern(), "{id}", string(key))
	} else if keyStr == "" {
		keyStr = fmt.Sprintf("event:%v", payload["id"])
	}

	switch mode {
	case "set":
		return s.client.Do(ctx, "SET", keyStr, string(data))
	case "lpush":
		target := s.cfg.GetStreamOrChannel()
		if target == "" {
			target = keyStr
		}
		return s.client.Do(ctx, "LPUSH", target, string(data))
	case "rpush":
		target := s.cfg.GetStreamOrChannel()
		if target == "" {
			target = keyStr
		}
		return s.client.Do(ctx, "RPUSH", target, string(data))
	case "publish":
		channel := s.cfg.GetStreamOrChannel()
		if channel == "" {
			channel = "mimic.events"
		}
		return s.client.Do(ctx, "PUBLISH", channel, string(data))
	case "xadd":
		stream := s.cfg.GetStreamOrChannel()
		if stream == "" {
			stream = "mimic:stream"
		}
		return s.client.Do(ctx, "XADD", stream, "*", "payload", string(data))
	default:
		return s.client.Do(ctx, "SET", keyStr, string(data))
	}
}

func (s *RedisSink) Close() error {
	return s.client.Close()
}

