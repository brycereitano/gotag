package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/brycereitano/gotag/tagger"
)

// ParseOffsetFlag interprets the "-offset" flag value as a renaming specification.
func ParseOffsetFlag(offsetFlag string) (*tagger.FilePosition, error) {
	// Validate -offset, e.g. file.go:#123
	parts := strings.Split(offsetFlag, ":#")
	if len(parts) != 2 {
		return nil, fmt.Errorf("-offset %q: invalid offset specification", offsetFlag)
	}

	for _, r := range parts[1] {
		if !unicode.IsDigit(r) {
			return nil, fmt.Errorf("-offset %q: non-numeric offset", offsetFlag)
		}
	}

	offset, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("-offset %q: non-numeric offset", offsetFlag)
	}

	return tagger.NewFilePosition(parts[0], offset)
}
