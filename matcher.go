package shrek

import (
	"bytes"
	"fmt"
	"strings"
)

type Matcher interface {
	MatchApprox(approx []byte) bool
	Match(exact []byte) bool
}

type StartEndMatcher struct {
	Start []byte
	End   []byte
}

func (m StartEndMatcher) MatchApprox(approx []byte) bool {
	return bytes.HasPrefix(approx[:EncodedPublicKeyApproxSize], m.Start)
}

func (m StartEndMatcher) Match(exact []byte) bool {
	return bytes.HasPrefix(exact, m.Start) && bytes.HasSuffix(exact, m.End)
}

func (m StartEndMatcher) Validate() error {
	const validRunes = "abcdefghijklmnopqrstuvwxyz234567"
	const maxLength = 56

	// Check filter length isn't too long.
	if l := len(m.Start) + len(m.End); l > maxLength {
		return fmt.Errorf("shrek: filter is too long (%d > %d)", l, maxLength)
	}

	if len(m.Start) > 0 {
		// Check for invalid chars in Start.
		if invalid := strings.Trim(string(m.Start), validRunes); invalid != "" {
			return fmt.Errorf("shrek: start part contains invalid chars: %q", invalid)
		}
	}

	// If no end search filter, then there's nothing else to validate.
	// Return early to reduce indenting.
	if len(m.End) == 0 {
		return nil
	}

	// Check for invalid chars in End.
	if invalid := strings.Trim(string(m.End), validRunes); invalid != "" {
		return fmt.Errorf("shrek: end part contains invalid chars: %q", invalid)
	}

	// If last char isn't "d".
	if chr := string(m.End[len(m.End)-1]); chr != "d" {
		return fmt.Errorf("shrek: last char in end part must be %q, not %q", "d", chr)
	}

	if len(m.End) > 1 {
		// If 2nd last char isn't any of "aiqy".
		if chr := string(m.End[len(m.End)-2]); strings.Trim(chr, "aiqy") != "" {
			return fmt.Errorf("shrek: 2nd last char in end part must be one of %q, not %q", "aiqy", chr)
		}
	}

	return nil
}

type MultiMatcher struct {
	Inner []Matcher

	// If All is true, then all the Inner matchers must match. If false, then only 1 of them
	// must match.
	All bool
}

func (m MultiMatcher) MatchApprox(approx []byte) bool {
	for _, im := range m.Inner {
		if match := im.MatchApprox(approx); match && !m.All {
			return true
		} else if !match && m.All {
			return false
		}
	}

	return m.All
}

func (m MultiMatcher) Match(exact []byte) bool {
	for _, im := range m.Inner {
		if match := im.MatchApprox(exact) && im.Match(exact); match && !m.All {
			return true
		} else if !match && m.All {
			return false
		}
	}

	return m.All
}
