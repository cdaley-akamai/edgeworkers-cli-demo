// Package search provides reusable local fuzzy filtering for CLI list results.
// FuzzyFilter scores fields supplied by its caller and preserves input order
// when scores are equal.
package search

import (
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/cases"
)

// maxFuzzyScore is the maximum Levenshtein distance accepted for fuzzy-only matches.
const maxFuzzyScore = 8

type match[T any] struct {
	item  T
	score int
}

// fieldScore returns the match score for query against field, or -1 if no match.
// Exact matches score 0, prefix matches 1, other substring matches 2, fuzzy matches 3+.
// This allows an ordering of sorting that makes sense with exact matches and substrings always filtering before fuzzy matches.
func fieldScore(query, foldedQuery, field string) int {
	foldedField := cases.Fold().String(field)
	if foldedField == foldedQuery {
		return 0
	}
	if strings.HasPrefix(foldedField, foldedQuery) {
		return 1
	}
	if strings.Contains(foldedField, foldedQuery) {
		return 2
	}
	if score := fuzzy.RankMatchNormalizedFold(query, field); score >= 0 && score <= maxFuzzyScore {
		return score + 3
	}
	return -1
}

// FuzzyFilter retains items matching query and orders them by best field score.
func FuzzyFilter[T any](items []T, query string, fields func(T) []string) []T {
	if query == "" {
		return items
	}

	foldedQuery := cases.Fold().String(query)
	matches := make([]match[T], 0, len(items))
	for _, item := range items {
		topFieldScore, found := -1, false
		for _, field := range fields(item) {
			score := fieldScore(query, foldedQuery, field)
			if score < 0 {
				continue
			}
			if !found || score < topFieldScore {
				topFieldScore, found = score, true
			}
		}
		if found {
			matches = append(matches, match[T]{item: item, score: topFieldScore})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].score < matches[j].score
	})

	result := make([]T, len(matches))
	for i, m := range matches {
		result[i] = m.item
	}
	return result
}
