package shrek_test

import (
	"fmt"
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
