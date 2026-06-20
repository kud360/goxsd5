package gotype

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/kud360/goxsd5/xsd"
)

// The lax lexical→value parsers. They accept the same lexical forms as
// xsdtype's strict parsers but land the value in an ordinary Go type, taking
// the Go-native shortcut where the strict space would do exact arithmetic:
// floats via strconv.ParseFloat, integers via strconv.ParseInt, the date/time
// family via time.Parse with a small layout table, durations via a lenient
// ISO-8601 walk into time.Duration.

func parseString(s string, _ xsd.ValueContext) (xsd.Value, error) { return strVal(s), nil }

func parseBoolean(s string, _ xsd.ValueContext) (xsd.Value, error) {
	switch s {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	}
	return nil, fmt.Errorf("invalid boolean %q", s)
}

func parseNumber(s string, _ xsd.ValueContext) (xsd.Value, error) {
	switch s {
	case "INF", "+INF":
		return math.Inf(1), nil
	case "-INF":
		return math.Inf(-1), nil
	case "NaN":
		return math.NaN(), nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number %q: %w", s, err)
	}
	return v, nil
}

func parseInteger(s string, _ xsd.ValueContext) (xsd.Value, error) {
	v, err := strconv.ParseInt(strings.TrimPrefix(s, "+"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	return v, nil
}

func parseHexBinary(s string, _ xsd.ValueContext) (xsd.Value, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hexBinary %q: %w", s, err)
	}
	return binVal(b), nil
}

func parseBase64Binary(s string, _ xsd.ValueContext) (xsd.Value, error) {
	b, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid base64Binary %q: %w", s, err)
	}
	return binVal(b), nil
}

func parseQName(s string, ctx xsd.ValueContext) (xsd.Value, error) {
	prefix, local := "", s
	if i := strings.IndexByte(s, ':'); i >= 0 {
		prefix, local = s[:i], s[i+1:]
	}
	if ctx == nil {
		return nil, fmt.Errorf("cannot resolve QName %q without a namespace context", s)
	}
	name, ok := ctx.ResolveQName(prefix, local)
	if !ok {
		return nil, fmt.Errorf("undefined namespace prefix %q in QName %q", prefix, s)
	}
	return name, nil
}

// temporalLayouts maps each date/time primitive's local name to the time.Parse
// layouts accepted for it, most specific first. The lax space lands every
// member of the date/time family on time.Time; forms that time.Parse cannot
// express (e.g. the gMonth --MM-- shape) are handled by the dedicated parsers
// below.
var temporalLayouts = map[string][]string{
	"dateTime":   {"2006-01-02T15:04:05.999999999Z07:00", "2006-01-02T15:04:05Z07:00", "2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05"},
	"date":       {"2006-01-02Z07:00", "2006-01-02"},
	"time":       {"15:04:05.999999999Z07:00", "15:04:05Z07:00", "15:04:05.999999999", "15:04:05"},
	"gYearMonth": {"2006-01Z07:00", "2006-01"},
	"gYear":      {"2006Z07:00", "2006"},
}

// parseTemporal returns the lax parser for the named date/time primitive,
// closing over its layout table.
func parseTemporal(local string) xsd.ParseFunc {
	layouts := temporalLayouts[local]
	return func(s string, _ xsd.ValueContext) (xsd.Value, error) {
		for _, layout := range layouts {
			if t, err := time.Parse(layout, s); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("invalid %s %q", local, s)
	}
}

// parseGMonthDay/parseGDay/parseGMonth handle the three recurring forms whose
// XSD lexical shape (leading --) time.Parse cannot consume directly; each strips
// the marker and lands the value on a time.Time in a fixed reference year/month,
// which is all the lax space promises. A trailing timezone is tolerated but
// folded into the parse, not preserved.
func parseGMonthDay(s string, _ xsd.ValueContext) (xsd.Value, error) {
	// --MM-DD
	body, _ := splitTimezone(strings.TrimPrefix(s, "--"))
	t, err := time.Parse("01-02", body)
	if err != nil {
		return nil, fmt.Errorf("invalid gMonthDay %q", s)
	}
	return t, nil
}

func parseGDay(s string, _ xsd.ValueContext) (xsd.Value, error) {
	// ---DD
	body, _ := splitTimezone(strings.TrimPrefix(s, "---"))
	t, err := time.Parse("02", body)
	if err != nil {
		return nil, fmt.Errorf("invalid gDay %q", s)
	}
	return t, nil
}

func parseGMonth(s string, _ xsd.ValueContext) (xsd.Value, error) {
	// --MM
	body, _ := splitTimezone(strings.TrimPrefix(s, "--"))
	t, err := time.Parse("01", body)
	if err != nil {
		return nil, fmt.Errorf("invalid gMonth %q", s)
	}
	return t, nil
}

// splitTimezone peels an optional trailing timezone (Z or ±HH:MM) off a lexical
// form, returning the body and the timezone text (empty when none).
func splitTimezone(s string) (body, tz string) {
	if strings.HasSuffix(s, "Z") {
		return s[:len(s)-1], "Z"
	}
	if i := strings.LastIndexAny(s, "+-"); i > 0 && len(s)-i == 6 {
		return s[:i], s[i:]
	}
	return s, ""
}

// parseDuration lands an xs:duration on time.Duration via a lenient ISO-8601
// walk. The lax space approximates the year and month designators (365 and 30
// days) since time.Duration has no calendar — a documented lenience.
func parseDuration(s string, _ xsd.ValueContext) (xsd.Value, error) {
	d, err := scanDuration(s)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func scanDuration(s string) (time.Duration, error) {
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	if !strings.HasPrefix(s, "P") {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	s = s[1:]
	datePart, timePart := s, ""
	if i := strings.IndexByte(s, 'T'); i >= 0 {
		datePart, timePart = s[:i], s[i+1:]
	}
	var total time.Duration
	dateUnits := []struct {
		designator byte
		span       time.Duration
	}{{'Y', 365 * 24 * time.Hour}, {'M', 30 * 24 * time.Hour}, {'D', 24 * time.Hour}}
	timeUnits := []struct {
		designator byte
		span       time.Duration
	}{{'H', time.Hour}, {'M', time.Minute}, {'S', time.Second}}
	d, err := accumulate(datePart, dateUnits)
	if err != nil {
		return 0, err
	}
	total += d
	d, err = accumulate(timePart, timeUnits)
	if err != nil {
		return 0, err
	}
	total += d
	if neg {
		total = -total
	}
	return total, nil
}

func accumulate(part string, units []struct {
	designator byte
	span       time.Duration
},
) (time.Duration, error) {
	var total time.Duration
	num := ""
	for i := 0; i < len(part); i++ {
		c := part[i]
		if (c >= '0' && c <= '9') || c == '.' {
			num += string(c)
			continue
		}
		for _, u := range units {
			if u.designator != c {
				continue
			}
			v, err := strconv.ParseFloat(num, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid duration component %q%c", num, c)
			}
			total += time.Duration(v * float64(u.span))
			num = ""
		}
	}
	return total, nil
}
