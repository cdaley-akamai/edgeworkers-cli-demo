package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItem struct {
	id   string
	name string
}

func testFields(item testItem) []string {
	return []string{item.id, item.name}
}

func TestFuzzyFilterMatchesAllFieldsAndOrdersAllMatches(t *testing.T) {
	items := []testItem{
		{id: "99", name: "billing-123"},
		{id: "1234", name: "unrelated"},
		{id: "123", name: "unrelated"},
		{id: "999", name: "unrelated"},
	}

	got := FuzzyFilter(items, "123", testFields)

	assert.Equal(t, []testItem{
		{id: "123", name: "unrelated"},
		{id: "1234", name: "unrelated"},
		{id: "99", name: "billing-123"},
	}, got)
}

func TestFuzzyFilterNormalizesCaseAndUnicode(t *testing.T) {
	items := []testItem{{id: "1", name: "Café Worker"}}

	got := FuzzyFilter(items, "CAFE", testFields)

	assert.Equal(t, items, got)
}

func TestFuzzyFilterPreservesOrderForEqualScores(t *testing.T) {
	items := []testItem{
		{id: "1", name: "ab"},
		{id: "2", name: "ac"},
	}

	got := FuzzyFilter(items, "a", testFields)

	assert.Equal(t, items, got)
}

func TestFuzzyFilterEmptyQueryPreservesItems(t *testing.T) {
	items := []testItem{
		{id: "1", name: "first"},
		{id: "2", name: "second"},
	}

	got := FuzzyFilter(items, "", testFields)

	assert.Equal(t, items, got)
}
