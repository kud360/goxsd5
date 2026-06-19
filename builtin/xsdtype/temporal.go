package xsdtype

import (
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdtemporal"
)

// The date/time and duration value spaces and their lexical parsers live in
// the leaf package xsdtemporal (pure calendar arithmetic, no schema model).
// This file is the thin adapter that re-exports those types under their
// historical xsdtype names and wraps each parser into the xsd.ParseFunc shape
// the builtin registry expects — the only xsd-specific knowledge (the
// Value/ValueContext signature) the date/time types need.

// DateTime and Duration are the date/time and duration value spaces; their
// methods (Compare, AddDuration, HasTimezone) and the DateTimeKind constants
// are defined in xsdtemporal. *DateTime satisfies xsd.TimezoneAware and both
// are dispatched by CompareValues.
type (
	DateTime     = xsdtemporal.DateTime
	Duration     = xsdtemporal.Duration
	DateTimeKind = xsdtemporal.DateTimeKind
)

const (
	KindDateTime   = xsdtemporal.KindDateTime
	KindDate       = xsdtemporal.KindDate
	KindTime       = xsdtemporal.KindTime
	KindGYearMonth = xsdtemporal.KindGYearMonth
	KindGYear      = xsdtemporal.KindGYear
	KindGMonthDay  = xsdtemporal.KindGMonthDay
	KindGDay       = xsdtemporal.KindGDay
	KindGMonth     = xsdtemporal.KindGMonth
)

// ParseDuration parses an xs:duration lexical into its value. It is called
// directly (not through xsd.ParseFunc) where a *Duration is wanted.
func ParseDuration(s string) (*Duration, error) { return xsdtemporal.ParseDuration(s) }

// The eight date/time primitives are registered through xsd.ParseFunc, so each
// adapter discards the (unused) ValueContext and lifts the concrete *DateTime
// into an xsd.Value — taking care to return a nil interface, not a typed-nil
// *DateTime, on error.
func ParseDateTime(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseDateTime(s))
}
func ParseDate(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseDate(s))
}
func ParseTime(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseTime(s))
}
func ParseGYear(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseGYear(s))
}
func ParseGYearMonth(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseGYearMonth(s))
}
func ParseGMonth(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseGMonth(s))
}
func ParseGDay(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseGDay(s))
}
func ParseGMonthDay(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return liftDT(xsdtemporal.ParseGMonthDay(s))
}

func liftDT(dt *DateTime, err error) (xsd.Value, error) {
	if err != nil {
		return nil, err
	}
	return dt, nil
}
