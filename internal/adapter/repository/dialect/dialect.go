package dialect

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Driver string

const (
	DriverSQLite   Driver = "sqlite"
	DriverPostgres Driver = "postgres"
	DriverOracle   Driver = "oracle"
)

type PlaceholderStyle int

const (
	PlaceholderQuestion PlaceholderStyle = iota // ? (SQLite, MySQL)
	PlaceholderDollar                           // $1, $2, ... (PostgreSQL)
	PlaceholderColon                            // :1, :2, ... (Oracle)
)

type Dialect interface {
	DriverName() Driver
	PlaceholderStyle() PlaceholderStyle
	Rebind(query string) string
	IsUniqueConstraintViolation(err error) bool
	UpsertStockLevelQuery() string
	UpsertTenantSettingsQuery() string
	FormatTime(t time.Time) any
	ParseTime(val any) (time.Time, error)
	FormatBool(b bool) any
	ParseBool(val any) (bool, error)
}

// RebindQuery converts standard '?' placeholders outside of string literals
// to the target placeholder format ($1, $2 or :1, :2).
func RebindQuery(query string, style PlaceholderStyle) string {
	if style == PlaceholderQuestion {
		return query
	}

	var sb strings.Builder
	sb.Grow(len(query) + 32)

	paramIndex := 1
	inSingleQuote := false
	runes := []rune(query)
	n := len(runes)

	for i := 0; i < n; i++ {
		ch := runes[i]
		if ch == '\'' {
			// Handle escaped single quote '' inside string literals
			if inSingleQuote && i+1 < n && runes[i+1] == '\'' {
				sb.WriteRune(ch)
				sb.WriteRune('\'')
				i++
				continue
			}
			inSingleQuote = !inSingleQuote
			sb.WriteRune(ch)
			continue
		}

		if ch == '?' && !inSingleQuote {
			switch style {
			case PlaceholderDollar:
				sb.WriteByte('$')
				sb.WriteString(strconv.Itoa(paramIndex))
			case PlaceholderColon:
				sb.WriteByte(':')
				sb.WriteString(strconv.Itoa(paramIndex))
			default:
				sb.WriteRune(ch)
			}
			paramIndex++
		} else {
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}

// CommonParseTime parses a database timestamp into time.Time flexibly.
func CommonParseTime(val any) (time.Time, error) {
	switch v := val.(type) {
	case time.Time:
		return v.UTC(), nil
	case string:
		str := strings.TrimSpace(v)
		if str == "" {
			return time.Time{}, nil
		}
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, str); err == nil {
				return t.UTC(), nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse time string: %s", str)
	case []byte:
		return CommonParseTime(string(v))
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type: %T", val)
	}
}

// CommonParseBool parses an arbitrary database boolean representation.
func CommonParseBool(val any) (bool, error) {
	switch v := val.(type) {
	case bool:
		return v, nil
	case int64:
		return v == 1, nil
	case int:
		return v == 1, nil
	case int32:
		return v == 1, nil
	case string:
		s := strings.TrimSpace(strings.ToLower(v))
		return s == "1" || s == "true" || s == "t" || s == "yes" || s == "y", nil
	case []byte:
		return CommonParseBool(string(v))
	default:
		return false, fmt.Errorf("unsupported boolean value type: %T", val)
	}
}
