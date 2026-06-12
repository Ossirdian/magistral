package planning

import (
	"strings"
	"sync"
	"time"
)

type StatusStore struct {
	mu      sync.RWMutex
	records map[string]RequestStatus
}

func NewStatusStore() *StatusStore {
	return &StatusStore{records: make(map[string]RequestStatus)}
}

func (s *StatusStore) SaveQueued(schoolID string, req RequestCreate, accepted RequestAccepted) {
	if s == nil {
		return
	}
	requestID := strings.TrimSpace(accepted.RequestID)
	if strings.TrimSpace(schoolID) == "" || requestID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key(schoolID, requestID)] = RequestStatus{
		RequestID:       requestID,
		Status:          accepted.Status,
		PeriodStart:     req.PeriodStart,
		PeriodEnd:       req.PeriodEnd,
		StudentIDs:      cleanStrings(req.StudentIDs),
		GroupIDs:        cleanStrings(req.GroupIDs),
		BranchIDs:       cleanStrings(req.BranchIDs),
		ReplaceStrategy: defaultReplaceStrategy(req.ReplaceStrategy),
		QueuedAt:        time.Now().UTC(),
	}
}

func (s *StatusStore) SaveCompleted(schoolID string, result ResultMessage, appliedLessons int, applySuccessful bool) {
	if s == nil {
		return
	}
	requestID := strings.TrimSpace(result.RequestID)
	if strings.TrimSpace(schoolID) == "" || requestID == "" {
		return
	}

	completedAt := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[key(schoolID, requestID)]
	record.RequestID = requestID
	record.Status = "completed"
	record.CompletedAt = &completedAt
	record.ReplaceStrategy = defaultReplaceStrategy(result.ReplaceStrategy)
	record.ScheduledLessons = len(result.Lessons)
	record.AppliedLessons = appliedLessons
	record.Conflicts = len(result.Conflicts)
	record.UnassignedStudents = len(result.UnassignedStudents)
	record.Score = result.Score
	record.ResourceIdleSeconds = result.ResourceIdleSeconds
	record.ScoreBreakdown = result.ScoreBreakdown
	record.ApplySuccessful = &applySuccessful
	s.records[key(schoolID, requestID)] = record
}

func (s *StatusStore) Get(schoolID, requestID string) (RequestStatus, bool) {
	if s == nil {
		return RequestStatus{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[key(schoolID, requestID)]
	return record, ok
}

func key(schoolID, requestID string) string {
	return strings.TrimSpace(schoolID) + ":" + strings.TrimSpace(requestID)
}
