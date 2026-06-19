package xsdtemporal

import (
	"fmt"
	"testing"
)

// parseByKind dispatches on a runtime DateTimeKind. Production code never
// needs this — the kind of every primitive is statically known, so it calls
// the type-specific ParseX directly. It exists only for the tests below, which
// sweep every kind in a loop.
func parseByKind(kind DateTimeKind, s string) (*DateTime, error) {
	switch kind {
	case KindDateTime:
		return ParseDateTime(s)
	case KindDate:
		return ParseDate(s)
	case KindTime:
		return ParseTime(s)
	case KindGYear:
		return ParseGYear(s)
	case KindGYearMonth:
		return ParseGYearMonth(s)
	case KindGMonth:
		return ParseGMonth(s)
	case KindGDay:
		return ParseGDay(s)
	case KindGMonthDay:
		return ParseGMonthDay(s)
	default:
		return nil, fmt.Errorf("unknown date/time kind %d parsing %q", kind, s)
	}
}

func TestDateTimeParse(t *testing.T) {
	valid := []struct {
		kind DateTimeKind
		in   string
	}{
		{KindDateTime, "2002-10-10T12:00:00-05:00"},
		{KindDateTime, "2002-10-10T17:00:00Z"},
		{KindDateTime, "2002-10-10T24:00:00"}, // rolls to next day
		{KindDateTime, "-0001-01-01T00:00:00"},
		{KindDateTime, "0000-01-01T00:00:00"}, // year 0 OK in 1.1
		{KindDateTime, "2024-02-29T00:00:00"}, // leap
		{KindDateTime, "12024-02-29T00:00:00.5"},
		{KindDate, "2002-10-10+13:00"},
		{KindTime, "13:20:00-05:00"},
		{KindTime, "24:00:00"},
		{KindGYear, "1999"},
		{KindGYear, "-0044"},
		{KindGYearMonth, "1999-05"},
		{KindGMonth, "--05"},
		{KindGMonth, "--05Z"},
		{KindGDay, "---31"},
		{KindGMonthDay, "--02-29"},
	}
	for _, tc := range valid {
		if _, err := parseByKind(tc.kind, tc.in); err != nil {
			t.Errorf("ParseDateTime(%v, %q): %v", tc.kind, tc.in, err)
		}
	}
	invalid := []struct {
		kind DateTimeKind
		in   string
	}{
		{KindDateTime, "2002-10-10"},          // missing time
		{KindDateTime, "2023-02-29T00:00:00"}, // not a leap year
		{KindDateTime, "2002-13-01T00:00:00"},
		{KindDateTime, "2002-00-01T00:00:00"},
		{KindDateTime, "2002-01-00T00:00:00"},
		{KindDateTime, "2002-01-01T25:00:00"},
		{KindDateTime, "2002-01-01T10:61:00"},
		{KindDateTime, "2002-01-01T10:00:61"},
		{KindDateTime, "2002-01-01T24:00:01"},
		{KindDateTime, "2002-01-01T00:00:00+15:00"}, // tz > 14h
		{KindDateTime, "02-01-01T00:00:00"},         // 2-digit year
		{KindDateTime, "002002-1-01T00:00:00"},      // 1-digit month
		{KindGYear, "99"},
		{KindGYear, "01999"}, // leading zero on 5 digits
		{KindGMonth, "--13"},
		{KindGMonthDay, "--02-30"},
		{KindGDay, "---32"},
		{KindDate, "2002/10/10"},
	}
	for _, tc := range invalid {
		if _, err := parseByKind(tc.kind, tc.in); err == nil {
			t.Errorf("ParseDateTime(%v, %q) should fail", tc.kind, tc.in)
		}
	}
}

func mustDT(t *testing.T, kind DateTimeKind, s string) *DateTime {
	t.Helper()
	dt, err := parseByKind(kind, s)
	if err != nil {
		t.Fatalf("ParseDateTime(%q): %v", s, err)
	}
	return dt
}

func TestDateTimeCompare(t *testing.T) {
	cmpDT := func(a, b string) (Order, bool) {
		return mustDT(t, KindDateTime, a).Compare(mustDT(t, KindDateTime, b))
	}
	// Same timezone presence: timeline order.
	if o, ok := cmpDT("2002-10-10T12:00:00Z", "2002-10-10T13:00:00Z"); !ok || o != OrderLess {
		t.Error("Z vs Z ordering")
	}
	// Cross-zone normalization: 12:00-05:00 == 17:00Z.
	if o, ok := cmpDT("2002-10-10T12:00:00-05:00", "2002-10-10T17:00:00Z"); !ok || o != OrderEqual {
		t.Error("12:00-05:00 should equal 17:00Z")
	}
	// 24:00 rollover.
	if o, ok := cmpDT("2002-10-09T24:00:00Z", "2002-10-10T00:00:00Z"); !ok || o != OrderEqual {
		t.Error("24:00 should roll to next midnight")
	}
	// TZ vs no-TZ: incomparable when within ±14h.
	if _, ok := cmpDT("2002-10-10T12:00:00Z", "2002-10-10T12:00:00"); ok {
		t.Error("should be incomparable")
	}
	// TZ vs no-TZ: comparable when beyond ±14h.
	if o, ok := cmpDT("2002-10-10T00:00:00Z", "2002-10-11T15:00:00"); !ok || o != OrderLess {
		t.Error("should be comparable: more than 14h apart")
	}
}

func TestDurationParseAndCompare(t *testing.T) {
	valid := []string{"P1Y", "P1M", "PT0S", "P1DT2H", "-P60D", "PT1M30.5S", "P1Y2M3DT10H30M", "PT0.5S"}
	for _, s := range valid {
		if _, err := ParseDuration(s); err != nil {
			t.Errorf("ParseDuration(%q): %v", s, err)
		}
	}
	invalid := []string{"", "P", "PT", "P1S", "PS", "1Y", "P-1Y", "P1.5Y", "PT1.5M2S", "P1Y2M3DT", "P1YT", "pt1s", "P1Y1M1DT1H1M1.S"}
	for _, s := range invalid {
		if _, err := ParseDuration(s); err == nil {
			t.Errorf("ParseDuration(%q) should fail", s)
		}
	}

	cmpDur := func(a, b string) (Order, bool) {
		da, _ := ParseDuration(a)
		db, _ := ParseDuration(b)
		return da.Compare(db)
	}
	if o, ok := cmpDur("P1Y", "P12M"); !ok || o != OrderEqual {
		t.Error("P1Y == P12M")
	}
	if o, ok := cmpDur("PT24H", "P1D"); !ok || o != OrderEqual {
		t.Error("PT24H == P1D")
	}
	if o, ok := cmpDur("P1Y", "P380D"); !ok || o != OrderLess {
		t.Error("P1Y < P380D")
	}
	// The canonical indeterminate pair.
	if _, ok := cmpDur("P1M", "P30D"); ok {
		t.Error("P1M vs P30D should be indeterminate")
	}
	if o, ok := cmpDur("P1M", "P27D"); !ok || o != OrderGreater {
		t.Error("P1M > P27D")
	}
	if o, ok := cmpDur("P1M", "P32D"); !ok || o != OrderLess {
		t.Error("P1M < P32D")
	}
}

func TestDayPinning(t *testing.T) {
	// 2000-01-31 + P1M pins to 2000-02-29 (leap), not March 2.
	dt := mustDT(t, KindDateTime, "2000-01-31T00:00:00")
	d, _ := ParseDuration("P1M")
	got := dt.AddDuration(d)
	if got.Year != 2000 || got.Month != 2 || got.Day != 29 {
		t.Errorf("2000-01-31 + P1M = %04d-%02d-%02d, want 2000-02-29", got.Year, got.Month, got.Day)
	}
	// Non-leap year pins to 28.
	dt2 := mustDT(t, KindDateTime, "2001-01-31T00:00:00")
	got2 := dt2.AddDuration(d)
	if got2.Month != 2 || got2.Day != 28 {
		t.Errorf("2001-01-31 + P1M = %02d-%02d, want 02-28", got2.Month, got2.Day)
	}
	// Negative seconds borrow across midnight.
	dn, _ := ParseDuration("-PT1S")
	got3 := mustDT(t, KindDateTime, "2000-03-01T00:00:00").AddDuration(dn)
	if got3.Month != 2 || got3.Day != 29 || got3.Hour != 23 || got3.Minute != 59 {
		t.Errorf("2000-03-01T00:00:00 - PT1S = %+v", got3)
	}
}

var durationSeeds = []string{
	"P1Y", "P1M", "P1D", "PT1H", "PT1M", "PT1S", "P1Y2M3DT4H5M6S",
	"-P1Y", "P0Y", "PT0.5S", "P1YT1S", "P10675199DT2H48M5.4775807S",
	"", "P", "PT", "1Y", "P1S", "PT1Y", "P-1Y", "PYM",
}

var dateTimeSeeds = []string{
	"2001-10-26T21:32:52", "2001-10-26T21:32:52+02:00", "2001-10-26T21:32:52Z",
	"2001-10-26", "21:32:52", "2001-10", "2001", "--10-26", "---26", "--10",
	"-0001-01-01T00:00:00", "9999-12-31T23:59:59.999999999",
	"24:00:00", "0000-01-01", "2001-13-01", "2001-02-30",
	"", "T", "Z", "2001-10-26T", "2001-10-26T25:00:00", "not-a-date",
}

func allKinds() []DateTimeKind {
	return []DateTimeKind{
		KindDateTime, KindDate, KindTime, KindGYearMonth,
		KindGYear, KindGMonthDay, KindGDay, KindGMonth,
	}
}

// FuzzParseDuration asserts no panic and reflexive comparability of any value.
func FuzzParseDuration(f *testing.F) {
	for _, s := range durationSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		d, err := ParseDuration(s)
		if err != nil {
			return
		}
		if d == nil {
			t.Fatalf("ParseDuration(%q) returned nil value, nil error", s)
		}
		if o, ok := d.Compare(d); !ok || o != OrderEqual {
			t.Fatalf("ParseDuration(%q): d.Compare(d) = (%v,%v), want (OrderEqual,true)", s, o, ok)
		}
	})
}

// FuzzParseDateTime drives all seven date/time primitives. Each parsed value
// must compare equal to itself.
func FuzzParseDateTime(f *testing.F) {
	for _, s := range dateTimeSeeds {
		for _, k := range allKinds() {
			f.Add(int(k), s)
		}
	}
	f.Fuzz(func(t *testing.T, kindIdx int, s string) {
		kinds := allKinds()
		kind := kinds[((kindIdx%len(kinds))+len(kinds))%len(kinds)]
		dt, err := parseByKind(kind, s)
		if err != nil {
			return
		}
		if dt == nil {
			t.Fatalf("ParseDateTime(%v,%q) returned nil value, nil error", kind, s)
		}
		if o, ok := dt.Compare(dt); !ok || o != OrderEqual {
			t.Fatalf("ParseDateTime(%v,%q): self-compare = (%v,%v), want (OrderEqual,true)", kind, s, o, ok)
		}
	})
}
