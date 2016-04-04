package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"database/sql/driver"
	"regexp"
)

// SqlStringSlice is a slice with support for postgres arrays
// Inspired by https://gist.github.com/adharris/4163702
type SqlStringSlice []string

var arrayRegex = regexp.MustCompile(`((((([^",\\{}\s(NULL)])+|"([^"\\]|\\"|\\\\)*")))(,)?)`)
var regexValIndex = 3

// Scan implements the Scanner interface.
func (s *SqlStringSlice) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return errors.New("Souce wan not []byte")
	}

	asStr := string(asBytes)
	results := make([]string, 0)
	matches := arrayRegex.FindAllStringSubmatch(asStr, -1)
	for _, match := range matches {
		s := match[regexValIndex]
		s = strings.Trim(s, "\"")
		results = append(results, s)
	}

	(*s) = SqlStringSlice(results)
	return nil
}

// Value implements the driver Valuer interface.
func (s SqlStringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return `{}`, nil
	}

	quoted := make([]string, len(s), len(s))
	for i, s := range s {
		quoted[i] = fmt.Sprintf(strconv.Quote(s))
	}

	return fmt.Sprintf("{%s}", strings.Join(quoted, ",")), nil
}
