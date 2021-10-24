package shrek

import (
	"bytes"
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
