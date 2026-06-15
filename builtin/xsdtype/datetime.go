package xsdtype

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/kud360/goxsd5/xsd"
)

// DateTimeKind says which of the seven date/time primitives a DateTime
// value belongs to (which properties are present).
type DateTimeKind int

const (
	KindDateTime DateTimeKind = iota
	KindDate
	KindTime
	KindGYearMonth
	KindGYear
	KindGMonthDay
	KindGDay
	KindGMonth
)

// DateTime is the value space of the date/time primitives, the
// seven-property model of Part 2 §3.3.7 (year, month, day, hour, minute,
// second, timezoneOffset), with absent properties zeroed and identified by
// Kind. Seconds carry arbitrary-precision fractions.
type DateTime struct {
	Kind   DateTimeKind
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second *big.Rat // nil means 0

	HasTZ bool
	TZ    int // offset in minutes, range ±840
}

// HasTimezone implements xsd.TimezoneAware: it reports whether this value
// carries a timezone offset (drives the explicitTimezone facet).
func (dt *DateTime) HasTimezone() bool { return dt.HasTZ }

func (dt *DateTime) sec() *big.Rat {
	if dt.Second == nil {
		return new(big.Rat)
	}
	return dt.Second
}

// daysInMonth for the proleptic Gregorian calendar (year 0 exists in the
// XSD 1.1 value space and is a leap year).
func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeap(year) {
			return 29
		}
		return 28
	}
	return 0
}

func isLeap(y int) bool {
	return y%4 == 0 && (y%100 != 0 || y%400 == 0)
}

// dayNumber is the count of days from 0000-01-01 (proleptic Gregorian).
func dayNumber(y, m, d int) int64 {
	// Shift so the year starts in March; standard civil-calendar math.
	a := int64(0)
	yy := int64(y)
	mm := int64(m)
	if mm <= 2 {
		yy--
		a = 1
	}
	era := yy / 400
	if yy < 0 && yy%400 != 0 {
		era--
	}
	yoe := yy - era*400
	doy := (153*(mm+12*a-3)+2)/5 + int64(d) - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe + 60 // +60 realigns to Jan 1 of year 0
}

// timeline returns the value's position on the timeline as seconds (an
// exact big.Rat), per the "time on timeline" mapping, with absent date
// fields defaulted consistently per Kind (both operands of a comparison
// share a Kind, so the defaults cancel out). tzShift is applied (in
// minutes) for the ±14:00 rule when the value has no timezone.
func (dt *DateTime) timeline(tzShift int) *big.Rat {
	y, mo, d := dt.Year, dt.Month, dt.Day
	switch dt.Kind {
	case KindTime:
		y, mo, d = 0, 1, 1
	case KindGYear, KindGYearMonth:
		if mo == 0 {
			mo = 1
		}
		d = 1
	case KindGMonth:
		y, d = 0, 1 // year 0 is leap: all month lengths admissible
	case KindGMonthDay:
		y = 0
	case KindGDay:
		y, mo = 0, 1
	}
	days := dayNumber(y, mo, d)
	secs := new(big.Rat).SetInt64(days*86400 + int64(dt.Hour)*3600 + int64(dt.Minute)*60)
	secs.Add(secs, dt.sec())
	tz := 0
	if dt.HasTZ {
		tz = dt.TZ
	} else {
		tz = tzShift
	}
	secs.Sub(secs, new(big.Rat).SetInt64(int64(tz)*60))
	return secs
}

// Compare implements the partial order of the date/time value spaces
// (Part 2 §3.3.7.2): values with timezones compare on the timeline; a
// timezoned and an untimezoned value compare only if the order is the same
// with the untimezoned one shifted to both +14:00 and -14:00.
func (dt *DateTime) Compare(o *DateTime) (xsd.Order, bool) {
	if dt.Kind != o.Kind {
		return 0, false
	}
	if dt.HasTZ == o.HasTZ {
		return xsd.Order(dt.timeline(0).Cmp(o.timeline(0))), true
	}
	lo := xsd.Order(dt.timeline(-840).Cmp(o.timeline(-840)))
	hi := xsd.Order(dt.timeline(840).Cmp(o.timeline(840)))
	if lo == hi && lo != xsd.OrderEqual {
		return lo, true
	}
	return 0, false
}

// ---- Parsing ----

// parseTZ splits a trailing timezone (Z or ±hh:mm) off s.
func parseTZ(s string) (rest string, hasTZ bool, tz int, err error) {
	if strings.HasSuffix(s, "Z") {
		return s[:len(s)-1], true, 0, nil
	}
	n := len(s)
	if n >= 6 && (s[n-6] == '+' || s[n-6] == '-') && s[n-3] == ':' {
		h, err1 := atoi2(s[n-5 : n-3])
		m, err2 := atoi2(s[n-2:])
		if err1 != nil || err2 != nil || h > 14 || m > 59 || (h == 14 && m != 0) {
			return "", false, 0, fmt.Errorf("invalid timezone in %q", s)
		}
		off := h*60 + m
		if s[n-6] == '-' {
			off = -off
		}
		return s[:n-6], true, off, nil
	}
	return s, false, 0, nil
}

func atoi2(s string) (int, error) {
	if len(s) != 2 || s[0] < '0' || s[0] > '9' || s[1] < '0' || s[1] > '9' {
		return 0, fmt.Errorf("bad digits %q", s)
	}
	return int(s[0]-'0')*10 + int(s[1]-'0'), nil
}

// parseYear parses the year field: optional '-', 4+ digits, no leading
// zeros beyond 4 digits. XSD 1.1 admits year 0000.
func parseYear(s string) (int, error) {
	t := s
	neg := false
	if strings.HasPrefix(t, "-") {
		neg = true
		t = t[1:]
	}
	if len(t) < 4 || (len(t) > 4 && t[0] == '0') {
		return 0, fmt.Errorf("invalid year %q", s)
	}
	y := 0
	for i := 0; i < len(t); i++ {
		if t[i] < '0' || t[i] > '9' {
			return 0, fmt.Errorf("invalid year %q", s)
		}
		y = y*10 + int(t[i]-'0')
		if y > 1<<31 {
			return 0, fmt.Errorf("year out of range %q", s)
		}
	}
	if neg {
		y = -y
	}
	return y, nil
}

// parseSecond parses ss(.fff…)? with exactly two integer digits.
func parseSecond(s string) (*big.Rat, error) {
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
		if frac == "" {
			return nil, fmt.Errorf("invalid seconds %q", s)
		}
	}
	si, err := atoi2(intPart)
	if err != nil || si > 60 {
		return nil, fmt.Errorf("invalid seconds %q", s)
	}
	r := new(big.Rat).SetInt64(int64(si))
	if frac != "" {
		den := big.NewInt(1)
		num := big.NewInt(0)
		for i := 0; i < len(frac); i++ {
			if frac[i] < '0' || frac[i] > '9' {
				return nil, fmt.Errorf("invalid seconds %q", s)
			}
			num.Mul(num, bigTen)
			num.Add(num, big.NewInt(int64(frac[i]-'0')))
			den.Mul(den, bigTen)
		}
		r.Add(r, new(big.Rat).SetFrac(num, den))
	}
	return r, nil
}

// parseDatePart parses yyyy-mm-dd (year already variable width).
func parseDatePart(s string) (y, m, d int, err error) {
	i := strings.LastIndexByte(s, '-')
	if i < 0 {
		return 0, 0, 0, fmt.Errorf("invalid date %q", s)
	}
	j := strings.LastIndexByte(s[:i], '-')
	if j < 0 {
		return 0, 0, 0, fmt.Errorf("invalid date %q", s)
	}
	if y, err = parseYear(s[:j]); err != nil {
		return
	}
	if m, err = atoi2(s[j+1 : i]); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid month in date %q: %w", s, err)
	}
	if d, err = atoi2(s[i+1:]); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid day in date %q: %w", s, err)
	}
	if m < 1 || m > 12 || d < 1 || d > daysInMonth(y, m) {
		return 0, 0, 0, fmt.Errorf("no such date %q", s)
	}
	return y, m, d, nil
}

// parseTimePart parses hh:mm:ss(.fff…)?, accepting 24:00:00 with the
// canonical rollover signalled via addDay.
func parseTimePart(s string) (h, mi int, sec *big.Rat, addDay bool, err error) {
	if len(s) < 8 || s[2] != ':' || s[5] != ':' {
		return 0, 0, nil, false, fmt.Errorf("invalid time %q", s)
	}
	if h, err = atoi2(s[:2]); err != nil {
		return 0, 0, nil, false, fmt.Errorf("invalid hours in time %q: %w", s, err)
	}
	if mi, err = atoi2(s[3:5]); err != nil {
		return 0, 0, nil, false, fmt.Errorf("invalid minutes in time %q: %w", s, err)
	}
	if sec, err = parseSecond(s[6:]); err != nil {
		return
	}
	secLimit := new(big.Rat).SetInt64(60)
	if sec.Cmp(secLimit) >= 0 {
		return 0, 0, nil, false, fmt.Errorf("seconds out of range in %q", s)
	}
	if h == 24 {
		if mi != 0 || sec.Sign() != 0 {
			return 0, 0, nil, false, fmt.Errorf("invalid 24:xx time %q", s)
		}
		return 0, 0, sec, true, nil
	}
	if h > 23 || mi > 59 {
		return 0, 0, nil, false, fmt.Errorf("time out of range %q", s)
	}
	return h, mi, sec, false, nil
}

func addOneDay(dt *DateTime) {
	dt.Day++
	if dt.Day > daysInMonth(dt.Year, dt.Month) {
		dt.Day = 1
		dt.Month++
		if dt.Month > 12 {
			dt.Month = 1
			dt.Year++
		}
	}
}

// newTemporal strips the timezoneFrag shared by every date/time
// ·…LexicalRep· and seeds a DateTime of the given kind (month and day default
// to 1, the value used for absent properties). It returns the timezone-free
// body for the caller's kind-specific lexical mapping.
func newTemporal(kind DateTimeKind, s string) (dt *DateTime, body string, err error) {
	body, hasTZ, tz, err := parseTZ(s)
	if err != nil {
		return nil, "", err
	}
	return &DateTime{Kind: kind, HasTZ: hasTZ, TZ: tz, Month: 1, Day: 1}, body, nil
}

// Each of the eight primitives is parsed by its own function — the lexical
// grammars genuinely differ, so there is no runtime dispatch on the production
// path: a caller registering xsd:date calls ParseDate directly (see
// builtin.go). The doc comment of each gives its Part 2 ·…LexicalRep·.

// ParseDateTime parses xsd:dateTime (§3.3.7):
//
//	dateTimeLexicalRep ::= yearFrag '-' monthFrag '-' dayFrag 'T'
//	                       ((hourFrag ':' minuteFrag ':' secondFrag) | endOfDayFrag)
//	                       timezoneFrag?
func ParseDateTime(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindDateTime, s)
	if err != nil {
		return nil, err
	}
	ti := strings.IndexByte(body, 'T')
	if ti < 0 {
		return nil, fmt.Errorf("invalid dateTime %q", s)
	}
	if dt.Year, dt.Month, dt.Day, err = parseDatePart(body[:ti]); err != nil {
		return nil, err
	}
	var addDay bool // endOfDayFrag (24:00:00) rolls into the next day
	if dt.Hour, dt.Minute, dt.Second, addDay, err = parseTimePart(body[ti+1:]); err != nil {
		return nil, err
	}
	if addDay {
		addOneDay(dt)
	}
	return dt, nil
}

// ParseDate parses xsd:date (§3.3.9):
//
//	dateLexicalRep ::= yearFrag '-' monthFrag '-' dayFrag timezoneFrag?
func ParseDate(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindDate, s)
	if err != nil {
		return nil, err
	}
	if dt.Year, dt.Month, dt.Day, err = parseDatePart(body); err != nil {
		return nil, err
	}
	return dt, nil
}

// ParseTime parses xsd:time (§3.3.8):
//
//	timeLexicalRep ::= ((hourFrag ':' minuteFrag ':' secondFrag) | endOfDayFrag)
//	                   timezoneFrag?
func ParseTime(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindTime, s)
	if err != nil {
		return nil, err
	}
	// endOfDayFrag (24:00:00) is 00:00:00 in the time value space: there is no
	// day to carry into, so the rollover flag is discarded.
	if dt.Hour, dt.Minute, dt.Second, _, err = parseTimePart(body); err != nil {
		return nil, err
	}
	return dt, nil
}

// ParseGYear parses xsd:gYear (§3.3.11):
//
//	gYearLexicalRep ::= yearFrag timezoneFrag?
func ParseGYear(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindGYear, s)
	if err != nil {
		return nil, err
	}
	if dt.Year, err = parseYear(body); err != nil {
		return nil, err
	}
	return dt, nil
}

// ParseGYearMonth parses xsd:gYearMonth (§3.3.10):
//
//	gYearMonthLexicalRep ::= yearFrag '-' monthFrag timezoneFrag?
func ParseGYearMonth(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindGYearMonth, s)
	if err != nil {
		return nil, err
	}
	i := strings.LastIndexByte(body, '-')
	if i < 1 {
		return nil, fmt.Errorf("invalid gYearMonth %q", s)
	}
	if dt.Year, err = parseYear(body[:i]); err != nil {
		return nil, err
	}
	if dt.Month, err = atoi2(body[i+1:]); err != nil || dt.Month < 1 || dt.Month > 12 {
		return nil, fmt.Errorf("invalid gYearMonth %q", s)
	}
	return dt, nil
}

// ParseGMonth parses xsd:gMonth (§3.3.14):
//
//	gMonthLexicalRep ::= '--' monthFrag timezoneFrag?
func ParseGMonth(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindGMonth, s)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(body, "--") {
		return nil, fmt.Errorf("invalid gMonth %q", s)
	}
	if dt.Month, err = atoi2(body[2:]); err != nil || dt.Month < 1 || dt.Month > 12 {
		return nil, fmt.Errorf("invalid gMonth %q", s)
	}
	return dt, nil
}

// ParseGDay parses xsd:gDay (§3.3.13):
//
//	gDayLexicalRep ::= '---' dayFrag timezoneFrag?
func ParseGDay(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindGDay, s)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(body, "---") {
		return nil, fmt.Errorf("invalid gDay %q", s)
	}
	if dt.Day, err = atoi2(body[3:]); err != nil || dt.Day < 1 || dt.Day > 31 {
		return nil, fmt.Errorf("invalid gDay %q", s)
	}
	return dt, nil
}

// ParseGMonthDay parses xsd:gMonthDay (§3.3.12):
//
//	gMonthDayLexicalRep ::= '--' monthFrag '-' dayFrag timezoneFrag?
func ParseGMonthDay(s string, _ xsd.ValueContext) (xsd.Value, error) {
	dt, body, err := newTemporal(KindGMonthDay, s)
	if err != nil {
		return nil, err
	}
	if len(body) != 7 || !strings.HasPrefix(body, "--") || body[4] != '-' {
		return nil, fmt.Errorf("invalid gMonthDay %q", s)
	}
	if dt.Month, err = atoi2(body[2:4]); err != nil || dt.Month < 1 || dt.Month > 12 {
		return nil, fmt.Errorf("invalid gMonthDay %q", s)
	}
	// Day validity uses a leap year so --02-29 is admissible.
	if dt.Day, err = atoi2(body[5:7]); err != nil || dt.Day < 1 || dt.Day > daysInMonth(0, dt.Month) {
		return nil, fmt.Errorf("invalid gMonthDay %q", s)
	}
	return dt, nil
}

// AddDuration implements dateTime + duration per Part 2 Appendix E: months
// are added with the day pinned to the resulting month's length, then the
// day/time part is added exactly.
func (dt *DateTime) AddDuration(d *Duration) *DateTime {
	res := *dt
	if res.Second == nil {
		res.Second = new(big.Rat)
	} else {
		res.Second = new(big.Rat).Set(res.Second)
	}
	// Months, carrying into years.
	total := int64(res.Year)*12 + int64(res.Month-1) + d.Months
	y := total / 12
	m := total % 12
	if m < 0 {
		m += 12
		y--
	}
	res.Year, res.Month = int(y), int(m)+1
	// Day pinning (Appendix E), not time.AddDate-style rollover.
	if max := daysInMonth(res.Year, res.Month); res.Day > max {
		res.Day = max
	}
	// Seconds: work on the timeline of the (pinned) date.
	days := dayNumber(res.Year, res.Month, res.Day)
	secs := new(big.Rat).SetInt64(days*86400 + int64(res.Hour)*3600 + int64(res.Minute)*60)
	secs.Add(secs, res.Second)
	secs.Add(secs, d.Seconds)
	// Decompose back.
	intSecs := new(big.Int).Quo(secs.Num(), secs.Denom())
	frac := new(big.Rat).Sub(secs, new(big.Rat).SetInt(intSecs))
	if frac.Sign() < 0 {
		intSecs.Sub(intSecs, big.NewInt(1))
		frac.Add(frac, big.NewRat(1, 1))
	}
	is := intSecs.Int64()
	day := is / 86400
	rem := is % 86400
	if rem < 0 {
		rem += 86400
		day--
	}
	res.Hour = int(rem / 3600)
	res.Minute = int(rem % 3600 / 60)
	res.Second = new(big.Rat).Add(new(big.Rat).SetInt64(rem%60), frac)
	res.Year, res.Month, res.Day = civilFromDays(day)
	return &res
}

// civilFromDays inverts dayNumber.
func civilFromDays(z int64) (y, m, d int) {
	z -= 60
	era := z / 146097
	if z < 0 && z%146097 != 0 {
		era--
	}
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	yy := yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	d = int(doy - (153*mp+2)/5 + 1)
	m = int(mp + 3)
	if mp >= 10 {
		m = int(mp - 9)
	}
	if m <= 2 {
		yy++
	}
	return int(yy), m, d
}
