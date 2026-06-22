package xsd

import (
	"errors"
	"fmt"
	"strings"
)

// Error is the structured error for every spec violation reported by this
// module. It renders as `uri:line:col: [id] message`.
type Error struct {
	Pos Pos
	// OtherPos is the secondary position for two-sided errors, e.g. the
	// previous declaration in a "conflicting definition" report.
	OtherPos Pos
	Ref      SpecRef
	Msg      string
	// cause is an optional underlying error, set when an *Error wraps a
	// plain error reported by a deeper layer (e.g. a builtin lexical parser).
	// It preserves errors.Is/As over that error without exposing it in Error().
	cause error
}

func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(e.Pos.String())
	b.WriteString(": ")
	if !e.Ref.IsZero() {
		b.WriteString("[")
		b.WriteString(e.Ref.ID)
		b.WriteString("] ")
	}
	b.WriteString(e.Msg)
	if !e.OtherPos.IsZero() {
		b.WriteString(" (see also ")
		b.WriteString(e.OtherPos.String())
		b.WriteString(")")
	}
	return b.String()
}

// NewError builds a spec violation at pos. The message is fmt.Sprintf-style.
func NewError(ref SpecRef, pos Pos, format string, args ...any) *Error {
	return &Error{Pos: pos, Ref: ref, Msg: fmt.Sprintf(format, args...)}
}

// WrapError attaches a SpecRef (the governing clause) to err, an error that a
// deeper layer reported without one — e.g. a builtin atomic parser's plain
// lexical-invalidity error. The original message is preserved verbatim and the
// wrapped error stays reachable via errors.Is/As.
func WrapError(ref SpecRef, err error) *Error {
	return &Error{Ref: ref, Msg: err.Error(), cause: err}
}

// Unwrap exposes the wrapped cause so errors.Is/As traverse it.
func (e *Error) Unwrap() error { return e.cause }

// ErrorList accumulates many *Error values so one parse reports every
// problem found rather than stopping at the first.
type ErrorList struct {
	errs []error
}

func (l *ErrorList) Add(err error) {
	if err == nil {
		return
	}
	// Flatten nested lists/joins so callers can re-aggregate freely.
	switch e := err.(type) {
	case interface{ Unwrap() []error }:
		l.errs = append(l.errs, e.Unwrap()...)
	default:
		l.errs = append(l.errs, err)
	}
}

func (l *ErrorList) Addf(ref SpecRef, pos Pos, format string, args ...any) {
	l.errs = append(l.errs, NewError(ref, pos, format, args...))
}

func (l *ErrorList) Empty() bool { return len(l.errs) == 0 }
func (l *ErrorList) Len() int    { return len(l.errs) }

// Err returns the accumulated errors joined, or nil if none.
func (l *ErrorList) Err() error {
	if len(l.errs) == 0 {
		return nil
	}
	return errors.Join(l.errs...)
}

// AllErrors unwraps an error produced by this module into its flat list of
// constituent errors.
func AllErrors(err error) []error {
	if err == nil {
		return nil
	}
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		return u.Unwrap()
	}
	return []error{err}
}

// RefIDs collects the distinct SpecRef IDs carried by err's constituents,
// in order of first appearance. Negative-conformance tests assert on these.
func RefIDs(err error) []string {
	var ids []string
	seen := map[string]bool{}
	for _, e := range AllErrors(err) {
		var xe *Error
		if errors.As(e, &xe) && !xe.Ref.IsZero() && !seen[xe.Ref.ID] {
			seen[xe.Ref.ID] = true
			ids = append(ids, xe.Ref.ID)
		}
	}
	return ids
}
