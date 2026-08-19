package cli

import (
	"fmt"
	"slices"
	"strings"
)

// parseCSV splits a comma-separated flag value, trims and dedupes its tokens,
// and validates each against avail. what leads the rejection message ("unknown
// language profile"). Empty-set policy and defaults stay at the call sites.
func parseCSV[T ~string](csv, what string, avail []T) ([]T, error) {
	var out []T
	for tok := range strings.SplitSeq(csv, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		v := T(tok)
		if !slices.Contains(avail, v) {
			return nil, fmt.Errorf("%s %q (available: %s)", what, tok, joinNames(avail))
		}
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out, nil
}

func joinNames[T ~string](vals []T) string {
	s := make([]string, len(vals))
	for i, v := range vals {
		s[i] = string(v)
	}
	return strings.Join(s, ", ")
}
