package planning

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var (
	ErrInvalidRequest = errors.New("invalid planning request")
	ErrNotConfigured  = errors.New("planning producer is not configured")
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	topic = strings.TrimSpace(topic)
	if len(brokers) == 0 || topic == "" {
		return nil, ErrNotConfigured
	}
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: kafka.RequireOne,
			Balancer:     &kafka.Hash{},
		},
	}, nil
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func (p *Producer) Publish(ctx context.Context, schoolID string, userID string, req RequestCreate) (RequestAccepted, error) {
	if p == nil || p.writer == nil {
		return RequestAccepted{}, ErrNotConfigured
	}
	if strings.TrimSpace(schoolID) == "" || !req.PeriodStart.Before(req.PeriodEnd) {
		return RequestAccepted{}, ErrInvalidRequest
	}

	requestID := strings.TrimSpace(valueOrEmpty(req.RequestID))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	message := RequestMessage{
		SchoolID:           strings.TrimSpace(schoolID),
		PeriodStart:        req.PeriodStart,
		PeriodEnd:          req.PeriodEnd,
		StudentIDs:         cleanStrings(req.StudentIDs),
		StudentPreferences: map[string]StudentPreference{},
		GroupIDs:           cleanStrings(req.GroupIDs),
		BranchIDs:          cleanStrings(req.BranchIDs),
		ReplaceStrategy:    defaultReplaceStrategy(req.ReplaceStrategy),
		RequestedByUserID:  strings.TrimSpace(userID),
		RequestID:          requestID,
	}
	if len(message.StudentIDs) == 0 && len(message.GroupIDs) == 0 {
		return RequestAccepted{}, ErrInvalidRequest
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return RequestAccepted{}, err
	}

	writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := p.writer.WriteMessages(writeCtx, kafka.Message{
		Key:   []byte(requestID),
		Value: payload,
		Time:  time.Now(),
	}); err != nil {
		return RequestAccepted{}, err
	}

	return RequestAccepted{RequestID: requestID, Status: "queued"}, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cleanStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func defaultReplaceStrategy(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "append"
	}
	return value
}
