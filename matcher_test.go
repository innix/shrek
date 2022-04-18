package shrek_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/innix/shrek"
)

func TestStartEndMatcher_MatchApprox(t *testing.T) {
	t.Parallel()

	input := "abcdyjsviqu5fqvqzv5mnfonrapka477vonf6fuko7duolp5g3i"
	table := []struct {
		Input string
		Start string
		End   string
		Match bool
	}{
		{Input: input, Start: "abcd", End: "i", Match: true},
		{Input: input, Start: "a", End: "5g3i", Match: true},
		{Input: input, Start: "abcd", End: "5g3i", Match: true},
		{Input: input, Start: "", End: "5g3i", Match: true},
		{Input: input, Start: "abcd", End: "", Match: true},
		{Input: input, Start: "", End: "", Match: true},
		{Input: input, Start: input, End: input, Match: true},

		{Input: input, Start: "b", End: "z", Match: false},
		{Input: input, Start: "bbb", End: "zzz", Match: false},
		{Input: input, Start: "b", End: "", Match: false},
		{Input: input, Start: "bbb", End: "", Match: false},
		{Input: input, Start: "bbb", End: "i", Match: false},
		{Input: input, Start: "bbb", End: "5g3i", Match: false},
	}

	for _, tc := range table {
		tc := tc
		name := fmt.Sprintf("%s:%s~=%s", tc.Start, tc.End, tc.Input)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := shrek.StartEndMatcher{
				Start: []byte(tc.Start),
				End:   []byte(tc.End),
			}

			if match := m.MatchApprox([]byte(tc.Input)); match != tc.Match {
				t.Errorf("invalid match result: got %v, wanted %v", match, tc.Match)
			}
		})
	}
}

func TestStartEndMatcher_Match(t *testing.T) {
	t.Parallel()

	const input = "abcdyjsviqu5fqvqzv5mnfonrapka477vonf6fuko7duolp5g3i"
	table := []struct {
		Input string
		Start string
		End   string
		Match bool
	}{
		{Input: input, Start: "abcd", End: "i", Match: true},
		{Input: input, Start: "a", End: "5g3i", Match: true},
		{Input: input, Start: "abcd", End: "5g3i", Match: true},
		{Input: input, Start: "", End: "5g3i", Match: true},
		{Input: input, Start: "abcd", End: "", Match: true},
		{Input: input, Start: "", End: "", Match: true},
		{Input: input, Start: input, End: input, Match: true},

		{Input: input, Start: "b", End: "z", Match: false},
		{Input: input, Start: "bbb", End: "zzz", Match: false},
		{Input: input, Start: "b", End: "", Match: false},
		{Input: input, Start: "bbb", End: "", Match: false},
		{Input: input, Start: "bbb", End: "i", Match: false},
		{Input: input, Start: "bbb", End: "5g3i", Match: false},
	}

	for _, tc := range table {
		tc := tc
		name := fmt.Sprintf("%s:%s~=%s", tc.Start, tc.End, tc.Input)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := shrek.StartEndMatcher{
				Start: []byte(tc.Start),
				End:   []byte(tc.End),
			}

			if match := m.Match([]byte(tc.Input)); match != tc.Match {
				t.Errorf("invalid match result: got %v, wanted %v", match, tc.Match)
			}
		})
	}
}

func TestStartEndMatcher_Valid(t *testing.T) {
	t.Parallel()

	type row struct {
		Start string
		End   string
		Valid bool
	}

	// Calculating permutations of all addresses takes way too long.
	// const validRunes = "abcdefghijklmnopqrstuvwxyz234567"
	const subsetValidRunes = "adiqyz7" // "adipqxyz257"
	var table []row

	for _, s := range permutations(t, []rune(subsetValidRunes)) {
		// Test End field.
		se := s[len(s)-2:]
		valid := se == "ad" || se == "id" || se == "qd" || se == "yd"
		table = append(table, row{Start: "", End: s, Valid: valid})

		// Test Start field - these should all be valid.
		table = append(table, row{Start: s, End: "", Valid: true})
	}

	// Repeat for uppercase.
	for _, s := range permutations(t, []rune(strings.ToUpper(subsetValidRunes))) {
		// Test End field - these should all be invalid.
		table = append(table, row{Start: "", End: s, Valid: false})

		// Test Start field - these should all be invalid..
		table = append(table, row{Start: s, End: "", Valid: false})
	}

	// Check search length of filter text.
	// An onion address is 56 chars, so anything above 56 is invalid.

	// Check Start length.
	table = append(table,
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaa", Valid: true},      // len = 56
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaa", Valid: false},    // len = 57
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaa", Valid: false},   // len = 58
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa", Valid: false},  // len = 59
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7", Valid: false}, // len = 60
	)

	// Check End length.
	table = append(table,
		row{End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaad", Valid: true},      // len = 56
		row{End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaad", Valid: false},    // len = 57
		row{End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaad", Valid: false},   // len = 58
		row{End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaad", Valid: false},  // len = 59
		row{End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaad", Valid: false}, // len = 60
	)

	// Check Start + End length.
	table = append(table,
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaa", End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaad", Valid: true},   // len = 56
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaa", End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaad", Valid: true},   // len = 56
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaa", End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaad", Valid: false}, // len = 57
		row{Start: "aaaaaaaaa7", End: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaad", Valid: false}, // len = 57
		row{Start: "aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaaaaa7aaaaaad", End: "aaaaaaaaad", Valid: false}, // len = 57
	)

	// Some realistic hand-written test cases, for good measure.
	table = append(table,
		row{Start: "food", End: "xid", Valid: true},
		row{Start: "food", End: "", Valid: true},
		row{Start: "", End: "xid", Valid: true},
		row{Start: "dark", End: "", Valid: true},
		row{Start: "dark", End: "yd", Valid: true},
		row{Start: "dark", End: "ydd", Valid: false},
		row{Start: "alpine9", End: "", Valid: false},
		row{Start: "alpine2", End: "", Valid: true},
	)

	for _, tc := range table {
		tc := tc
		name := fmt.Sprintf("%s:%s=%v", tc.Start, tc.End, tc.Valid)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := shrek.StartEndMatcher{
				Start: []byte(tc.Start),
				End:   []byte(tc.End),
			}

			if err := m.Validate(); err == nil && !tc.Valid {
				t.Errorf("invalid validation result: wanted: non-nil error, got: nil error")
			} else if err != nil && tc.Valid {
				t.Errorf("invalid validation result: wanted: nil error, got: %v", err)
			}
		})
	}
}

func permutations(t *testing.T, charset []rune) []string {
	t.Helper()

	var perms []string
	var permFn func(*testing.T, []rune, int)

	permFn = func(t *testing.T, rs []rune, n int) {
		t.Helper()

		if n == 1 {
			tmp := make([]rune, len(rs))
			copy(tmp, rs)
			perms = append(perms, string(tmp))
			return
		}

		for i := 0; i < n; i++ {
			permFn(t, rs, n-1)

			if n%2 == 1 {
				tmp := rs[i]
				rs[i] = rs[n-1]
				rs[n-1] = tmp
			} else {
				tmp := rs[0]
				rs[0] = rs[n-1]
				rs[n-1] = tmp
			}
		}
	}

	permFn(t, charset, len(charset))
	return perms
}
