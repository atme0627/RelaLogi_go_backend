// Package apperror provides a transport-agnostic application error that
// carries enough classification (Kind, Code, Message) for an outer layer to
// translate it into a protocol-specific response (e.g. an HTTP status code).
package apperror

import "fmt"

// Kind classifies an error so that a transport layer can decide how to render
// it. The zero value is KindInternal, so an unclassified error defaults to a
// server-side (500-equivalent) failure.
type Kind int

const (
	// KindInternal is a server-side failure that should not be exposed in detail.
	KindInternal Kind = iota
	// KindBadRequest is a client-side failure caused by invalid input.
	KindBadRequest
)

// Error is the internal error type. It is independent of any transport: it
// holds a machine-readable Code and a client-safe Message, and wraps the
// underlying cause so the diagnostic chain is preserved for logging.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause so errors.Is/errors.As can walk the chain.
func (e *Error) Unwrap() error {
	return e.Err
}

// BadRequest builds a KindBadRequest error. cause may be nil.
func BadRequest(code, message string, cause error) *Error {
	return &Error{Kind: KindBadRequest, Code: code, Message: message, Err: cause}
}

// Internal builds a KindInternal error. cause may be nil.
func Internal(code, message string, cause error) *Error {
	return &Error{Kind: KindInternal, Code: code, Message: message, Err: cause}
}
