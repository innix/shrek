package main

import (
	"fmt"
	"strings"
)

type appOptions struct {
	NumAddresses  int
	SaveDirectory string
	NumThreads    int
	Formatting    formatting
	Patterns      []string
}

type formatting string

const (
	BasicFormatting    = formatting("basic")
	ColorFormatting    = formatting("colored")
	EnhancedFormatting = formatting("enhanced")
	AllFormatting      = formatting("")
)

func (f *formatting) String() string {
	return string(*f)
}

func (f *formatting) Set(v string) error {
	fv := formatting(strings.ToLower(v))

	switch fv {
	case BasicFormatting, ColorFormatting, EnhancedFormatting, AllFormatting:
		*f = fv
		return nil
	case "all":
		*f = AllFormatting
		return nil
	default:
		return fmt.Errorf("parsing %q: invalid format kind", v)
	}
}

func (f *formatting) Type() string {
	return "string"
}

func (f *formatting) UseColors() bool {
	return *f == ColorFormatting || *f == AllFormatting
}

func (f *formatting) UseEnhanced() bool {
	return *f == EnhancedFormatting || *f == AllFormatting
}
