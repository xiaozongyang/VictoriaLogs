package logstorage

import (
	"testing"
)

func TestParseQueryStatsSuccess(t *testing.T) {
	f := func(pipeStr string) {
		t.Helper()
		expectParsePipeSuccess(t, pipeStr)
	}

	f(`query_stats`)
}

func TestParseQueryStatsFailure(t *testing.T) {
	f := func(pipeStr string) {
		t.Helper()
		expectParsePipeFailure(t, pipeStr)
	}

	f(`query_stats x`)
	f(`query_stats 0`)
}

func TestPipeQueryStatsUpdateNeededFields(t *testing.T) {
	f := func(s string, allowFilters, denyFilters, allowFiltersExpected, denyFiltersExpected string) {
		t.Helper()
		expectPipeNeededFields(t, s, allowFilters, denyFilters, allowFiltersExpected, denyFiltersExpected)
	}

	// all the needed fields
	f("query_stats", "*", "", "*", "")

	// all the needed fields, plus unneeded fields
	f("query_stats", "*", "f1,f2", "*", "")

	// needed fields
	f("query_stats", "f1,f2", "", "*", "")
}
