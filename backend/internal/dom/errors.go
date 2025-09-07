package dom

import (
	"errors"
	"fmt"
)

// -----------------------------------------------------------------------------
// Sentinel errors (infra tags with these; services convert via ToDomErr)
// -----------------------------------------------------------------------------

var (
	ErrUnavailable        = errors.New("provider unavailable")
	ErrInternal           = errors.New("internal provider error")
	ErrNoResults          = errors.New("no results found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidInput       = errors.New("invalid input")
	ErrTimeout            = errors.New("timeout")
	ErrExpired            = errors.New("resource expired or revoked")
	ErrRateLimit          = errors.New("rate limited")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrOwnershipViolation = errors.New("ownership violation")
	ErrInvalidState       = errors.New("invalid state")
	ErrNotPingable        = errors.New("not pingable")
)

// -----------------------------------------------------------------------------
// Infrastructure Error
// -----------------------------------------------------------------------------

// TaggedError is returned by the infrastructure layer.
// It wraps an underlying error with a domain-specific sentinel tag.
type TaggedError struct {
	Cause error
	Tag   error
}

// Error implements the error interface.
func (e *TaggedError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Tag, e.Cause)
	}
	return e.Tag.Error()
}

// Unwrap returns the root cause, allowing for error chaining via errors.Unwrap.
func (e *TaggedError) Unwrap() error {
	return e.Cause
}

// Is allows checking if the error matches a given sentinel tag via errors.Is.
func (e *TaggedError) Is(target error) bool {
	return errors.Is(e.Tag, target)
}

// NewTaggedError is a constructor for TaggedError.
func NewTaggedError(tag, cause error) error {
	return &TaggedError{Tag: tag, Cause: cause}
}

// -----------------------------------------------------------------------------
// Stable codes (use these for constructors + tests to avoid typos)
// -----------------------------------------------------------------------------

const (
	CodeTimeout            = "TIMEOUT"
	CodeInvalidArgument    = "INVALID_ARGUMENT"
	CodeNotFound           = "NOT_FOUND"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeConflict           = "CONFLICT"
	CodeInternal           = "INTERNAL"
	CodeUnavailable        = "UNAVAILABLE"
	CodeExpired            = "EXPIRED"
	CodeRateLimit          = "RATE_LIMIT"
	CodeOwnershipViolation = "OWNERSHIP_VIOLATION"
	CodeInvalidState       = "INVALID_STATE"
	CodeNotPingable        = "NOT_PINGABLE"
)

// -----------------------------------------------------------------------------
// Domain Error
// -----------------------------------------------------------------------------

// DomainError is the primary error type for the service layer. It enriches
// a raw error with structured, domain-specific information.
type DomainError struct {
	Cause     error
	Code      string
	Message   string
	Retryable bool
}

// Error implements the error interface, providing a developer-friendly output.
func (e *DomainError) Error() string {
	return fmt.Sprintf("code=%s, retryable=%t, message='%s', cause=%v",
		e.Code, e.Retryable, e.Message, e.Cause)
}

// Unwrap returns the original underlying error.
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// -----------------------------------------------------------------------------
// Conversion & Helpers
// -----------------------------------------------------------------------------

var errorMetadata = map[error]struct {
	Code      string
	Retryable bool
}{
	ErrUnavailable:        {Code: CodeUnavailable, Retryable: true},
	ErrInternal:           {Code: CodeInternal, Retryable: true},
	ErrNoResults:          {Code: CodeNotFound, Retryable: false},
	ErrConflict:           {Code: CodeConflict, Retryable: false},
	ErrInvalidInput:       {Code: CodeInvalidArgument, Retryable: false},
	ErrTimeout:            {Code: CodeTimeout, Retryable: true},
	ErrExpired:            {Code: CodeExpired, Retryable: false},
	ErrRateLimit:          {Code: CodeRateLimit, Retryable: true},
	ErrUnauthorized:       {Code: CodeUnauthorized, Retryable: false},
	ErrForbidden:          {Code: CodeForbidden, Retryable: false},
	ErrOwnershipViolation: {Code: CodeOwnershipViolation, Retryable: false},
	ErrInvalidState:       {Code: CodeInvalidState, Retryable: false},
	ErrNotPingable:        {Code: CodeNotPingable, Retryable: true},
}

// ToDomainError converts any error into a structured DomainError.
// It inspects the error chain for known sentinel tags to apply the correct
// code and properties. If no known tag is found, it defaults to a generic
// internal error.
func ToDomainError(err error) *DomainError {
	if err == nil {
		return nil
	}

	// Default to a generic internal error.
	de := &DomainError{
		Cause:     err,
		Code:      CodeInternal,
		Message:   "An unexpected error occurred.",
		Retryable: true, // Default to retryable for safety.
	}

	// Search for a known sentinel tag in the error chain.
	for tag, meta := range errorMetadata {
		if errors.Is(err, tag) {
			de.Code = meta.Code
			de.Retryable = meta.Retryable
			de.Message = tag.Error() // Use the sentinel's message for the dev message.
			break
		}
	}

	return de
}

// AsDomainError attempts to convert an error to a *DomainError.
// It returns the error and true if successful, otherwise nil and false.
func AsDomainError(err error) (*DomainError, bool) {
	var de *DomainError
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}

// IsRetryable checks if an error is marked as retryable.
// It safely handles non-DomainError types.
func IsRetryable(err error) bool {
	if de, ok := AsDomainError(err); ok {
		return de.Retryable
	}
	return ToDomainError(err).Retryable
}

// GetCode returns the machine-readable code of a DomainError.
// Returns CodeInternal for any non-DomainError types.
func GetCode(err error) string {
	if de, ok := AsDomainError(err); ok {
		return de.Code
	}
	return ToDomainError(err).Code
}
