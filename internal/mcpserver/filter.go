// Package mcpserver provides shared runtime behavior for product MCP servers.
package mcpserver

import (
	"fmt"

	"github.com/gobwas/glob"
)

type toolFilter struct {
	patterns []*glob.Pattern
}

func newToolFilter(patterns []string) (toolFilter, error) {
	filter := toolFilter{patterns: make([]*glob.Pattern, 0, len(patterns))}
	for _, pattern := range patterns {
		compiled, err := glob.Compile(pattern)
		if err != nil {
			return toolFilter{}, fmt.Errorf("invalid tool pattern %q: %w", pattern, err)
		}
		filter.patterns = append(filter.patterns, compiled)
	}
	return filter, nil
}

func (f toolFilter) matches(name string) bool {
	for _, pattern := range f.patterns {
		if pattern.Match(name) {
			return true
		}
	}
	return false
}
