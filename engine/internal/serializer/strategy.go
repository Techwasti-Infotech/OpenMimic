package serializer

import (
	"fmt"
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// SubjectNamingStrategy defines how a Schema Registry subject name is computed.
type SubjectNamingStrategy string

const (
	StrategyTopicName       SubjectNamingStrategy = "TopicNameStrategy"
	StrategyRecordName      SubjectNamingStrategy = "RecordNameStrategy"
	StrategyTopicRecordName SubjectNamingStrategy = "TopicRecordNameStrategy"
)

// ComputeSubject resolves the subject name based on strategy, topic, record name, and key/value orientation.
//
// 1. If an explicit subject is configured, it is respected directly.
// 2. TopicNameStrategy (default):
//    - Value: "<topic>-value"
//    - Key:   "<topic>-key"
// 3. RecordNameStrategy:
//    - "<record_name>" (e.g. "io.confluent.examples.OrderEvent")
// 4. TopicRecordNameStrategy:
//    - Value: "<topic>-<record_name>"
//    - Key:   "<topic>-key-<record_name>" (or "<topic>-<record_name>-key")
func ComputeSubject(cfg *pb.SchemaRegistryConfig, topic string) string {
	if cfg == nil {
		if topic == "" {
			return "openmimic-default-value"
		}
		return topic + "-value"
	}

	// 1. Explicit override
	if strings.TrimSpace(cfg.GetSubject()) != "" {
		return strings.TrimSpace(cfg.GetSubject())
	}

	strategyStr := strings.TrimSpace(cfg.GetSubjectNamingStrategy())
	if strategyStr == "" {
		strategyStr = string(StrategyTopicName)
	}

	recordName := strings.TrimSpace(cfg.GetRecordName())
	if recordName == "" {
		recordName = "openmimic.Record"
	}

	suffix := "value"
	if cfg.GetIsKey() {
		suffix = "key"
	}

	switch strings.ToLower(strategyStr) {
	case strings.ToLower(string(StrategyRecordName)):
		return recordName

	case strings.ToLower(string(StrategyTopicRecordName)):
		if topic == "" {
			topic = "default"
		}
		if cfg.GetIsKey() {
			return fmt.Sprintf("%s-%s-key", topic, recordName)
		}
		return fmt.Sprintf("%s-%s", topic, recordName)

	case strings.ToLower(string(StrategyTopicName)):
		fallthrough
	default:
		if topic == "" {
			topic = "default"
		}
		return fmt.Sprintf("%s-%s", topic, suffix)
	}
}

