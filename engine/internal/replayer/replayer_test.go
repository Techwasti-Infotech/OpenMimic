package replayer

import (
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestParseNginxCombinedLog(t *testing.T) {
	replayer := NewLogReplayer(&pb.LogReplayerConfig{
		LogFormat: "nginx_combined",
	})

	line := `192.168.1.50 - admin [10/Oct/2026:13:55:36 +0000] "GET /api/v1/orders HTTP/1.1" 200 4521 "https://portal.openmimic.org" "Mozilla/5.0"`
	record, err := replayer.ParseLine(line)
	if err != nil {
		t.Fatalf("failed to parse combined log line: %v", err)
	}

	if record["remote_ip"] != "192.168.1.50" {
		t.Errorf("expected remote_ip 192.168.1.50, got %v", record["remote_ip"])
	}
	if record["status"] != 200 {
		t.Errorf("expected status 200, got %v", record["status"])
	}
	if record["bytes_sent"] != int64(4521) {
		t.Errorf("expected bytes_sent 4521, got %v", record["bytes_sent"])
	}
}

func TestParseJSONLAuditLog(t *testing.T) {
	replayer := NewLogReplayer(&pb.LogReplayerConfig{
		LogFormat: "jsonl_audit",
	})

	line := `{"user_id":"usr_99","action":"LOGIN_SUCCESS","ip":"10.0.0.1"}`
	record, err := replayer.ParseLine(line)
	if err != nil {
		t.Fatalf("failed to parse jsonl audit line: %v", err)
	}

	if record["action"] != "LOGIN_SUCCESS" {
		t.Errorf("expected LOGIN_SUCCESS, got %v", record["action"])
	}
	if record["_replayed_at"] == nil {
		t.Errorf("expected _replayed_at timestamp inflation")
	}
}
