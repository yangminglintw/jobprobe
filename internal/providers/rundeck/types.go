// Package rundeck provides a Rundeck job execution provider.
package rundeck

import "time"

// ExecutionStatus represents Rundeck execution statuses.
type ExecutionStatus string

const (
	ExecutionStatusRunning         ExecutionStatus = "running"
	ExecutionStatusSucceeded       ExecutionStatus = "succeeded"
	ExecutionStatusFailed          ExecutionStatus = "failed"
	ExecutionStatusAborted         ExecutionStatus = "aborted"
	ExecutionStatusTimedOut        ExecutionStatus = "timedout"
	ExecutionStatusFailedWithRetry ExecutionStatus = "failed-with-retry"
	ExecutionStatusScheduled       ExecutionStatus = "scheduled"
	ExecutionStatusPending         ExecutionStatus = "pending"
)

// IsTerminal returns true if the status is a terminal state.
func (s ExecutionStatus) IsTerminal() bool {
	switch s {
	case ExecutionStatusSucceeded, ExecutionStatusFailed, ExecutionStatusAborted,
		ExecutionStatusTimedOut, ExecutionStatusFailedWithRetry:
		return true
	default:
		return false
	}
}

// RunJobRequest represents a request to run a Rundeck job.
type RunJobRequest struct {
	Options map[string]string `json:"options,omitempty"`
}

// RunJobResponse represents the response from running a Rundeck job.
type RunJobResponse struct {
	ID          int       `json:"id"`
	Href        string    `json:"href"`
	Permalink   string    `json:"permalink"`
	Status      string    `json:"status"`
	DateStarted DateInfo  `json:"date-started"`
	Job         JobInfo   `json:"job"`
	Description string    `json:"description"`
	ArgString   string    `json:"argstring"`
	Project     string    `json:"project"`
}

// DateInfo represents a date in Rundeck responses.
type DateInfo struct {
	UnixTime int64     `json:"unixtime"`
	Date     time.Time `json:"date"`
}

// JobInfo represents job information in Rundeck responses.
type JobInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Group           string `json:"group"`
	Project         string `json:"project"`
	Href            string `json:"href"`
	Permalink       string `json:"permalink"`
	AverageDuration int64  `json:"averageDuration"`
}

// ExecutionResponse represents a Rundeck execution status response.
type ExecutionResponse struct {
	ID            int             `json:"id"`
	Href          string          `json:"href"`
	Permalink     string          `json:"permalink"`
	Status        ExecutionStatus `json:"status"`
	DateStarted   DateInfo        `json:"date-started"`
	DateEnded     DateInfo        `json:"date-ended"`
	Job           JobInfo         `json:"job"`
	Description   string          `json:"description"`
	ArgString     string          `json:"argstring"`
	Project       string          `json:"project"`
	FailedNodes   []string        `json:"failedNodes,omitempty"`
	SuccessfulNodes []string      `json:"successfulNodes,omitempty"`
}

// ErrorResponse represents an error response from Rundeck.
type ErrorResponse struct {
	Error      bool   `json:"error"`
	APIVersion int    `json:"apiversion"`
	ErrorCode  string `json:"errorCode"`
	Message    string `json:"message"`
}
