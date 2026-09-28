// Package price parses the yen amounts external catalogs return as free text.
package price

import (
	"strconv"
	"strings"
)

// ParseYen reads the leading digits of s (e.g. "310 (税込)" → 310, "1,650" → 1650).
// Catalogs mix tax-included and tax-excluded values, so the result only prefills a
// form the user confirms; nil means no usable amount.
func ParseYen(s string) *int64 {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return nil
	}
	n, err := strconv.ParseInt(s[:end], 10, 64)
	if err != nil {
		return nil
	}
	return &n
}
