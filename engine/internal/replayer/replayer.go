package replayer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// LogReplayer parses raw historical log lines into structured event maps.
type LogReplayer struct {
	format          string
	speedMultiplier float64
	loop            bool
}

// NewLogReplayer creates a new LogReplayer.
func NewLogReplayer(cfg *pb.LogReplayerConfig) *LogReplayer {
	format := "jsonl_audit"
	if cfg.GetLogFormat() != "" {
		format = cfg.GetLogFormat()
	}
	speed := cfg.GetSpeedMultiplier()
	if speed <= 0 {
		speed = 1.0
	}
	return &LogReplayer{
		format:          format,
		speedMultiplier: speed,
		loop:            cfg.GetLoopReplay(),
	}
}

// Regex for Nginx/Apache Combined Log format:
// 127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
var combinedLogRegex = regexp.MustCompile(`^(\S+)\s+\S+\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+([^"]+)\s+([^"]+)"\s+(\d{3})\s+(\d+|-)(?:\s+"([^"]*)"\s+"([^"]*)")?`)

// ParseLine parses a single log line into a key-value record map.
func (r *LogReplayer) ParseLine(line string) (map[string]any, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, nil
	}

	switch r.format {
	case "jsonl_audit":
		var record map[string]any
		if err := json.Unmarshal([]byte(trimmed), &record); err != nil {
			return nil, fmt.Errorf("failed to parse jsonl log line: %w", err)
		}
		// Inflate timestamp to current time for streaming simulation
		record["_replayed_at"] = time.Now().UTC().Format(time.RFC3339)
		return record, nil

	case "nginx_combined", "apache_common":
		matches := combinedLogRegex.FindStringSubmatch(trimmed)
		if len(matches) < 9 {
			return nil, fmt.Errorf("line does not match combined log regex: %s", trimmed)
		}

		status, _ := strconv.Atoi(matches[7])
		var bytesSent int64
		if matches[8] != "-" {
			bytesSent, _ = strconv.ParseInt(matches[8], 10, 64)
		}

		record := map[string]any{
			"remote_ip":     matches[1],
			"remote_user":   matches[2],
			"raw_time":      matches[3],
			"method":        matches[4],
			"path":          matches[5],
			"protocol":      matches[6],
			"status":        status,
			"bytes_sent":    bytesSent,
			"_replayed_at":  time.Now().UTC().Format(time.RFC3339),
		}

		if len(matches) >= 11 {
			record["referer"] = matches[9]
			record["user_agent"] = matches[10]
		}
		return record, nil

	default:
		return map[string]any{
			"raw_log":      trimmed,
			"_replayed_at": time.Now().UTC().Format(time.RFC3339),
		}, nil
	}
}

// ReplayFile reads a log file and invokes the callback for each parsed record.
func (r *LogReplayer) ReplayFile(path string, handler func(record map[string]any) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		rec, err := r.ParseLine(line)
		if err != nil || rec == nil {
			continue
		}
		if err := handler(rec); err != nil {
			return err
		}
	}
	return scanner.Err()
}
