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

func TestPipeQueryStats(t *testing.T) {
	f := func(pipeStr string, rows, rowsExpected [][]Field) {
		t.Helper()
		expectPipeResults(t, pipeStr, rows, rowsExpected)
	}

	// empty input
	f(`query_stats`, [][]Field{}, [][]Field{
		{
			{"blocksProcessed", "0"},
			{"bytesReadBlockHeaders", "0"},
			{"bytesReadBloomFilters", "0"},
			{"bytesReadColumnsHeaderIndexes", "0"},
			{"bytesReadColumnsHeaders", "0"},
			{"bytesReadTimestamps", "0"},
			{"bytesReadTotal", "0"},
			{"bytesReadValues", "0"},
			{"timestampsRead", "0"},
			{"valuesRead", "0"},
			{"bytesProcessedUncompressedValues", "0"},
		},
	})

	// non-empty input
	//
	// The returned query stats is empty because the expectPipeResults() doesn't store
	// the rows into database and doesn't read them from the database.
	f(`query_stats`, [][]Field{
		{
			{"foo", "bar"},
			{"abc", "defaaa"},
		},
		{
			{"_msg", "qfdskj lj lkfdsjfds"},
			{"_time", "2025-08-30T10:20:30Z"},
		},
	}, [][]Field{
		{
			{"blocksProcessed", "0"},
			{"bytesReadBlockHeaders", "0"},
			{"bytesReadBloomFilters", "0"},
			{"bytesReadColumnsHeaderIndexes", "0"},
			{"bytesReadColumnsHeaders", "0"},
			{"bytesReadTimestamps", "0"},
			{"bytesReadTotal", "0"},
			{"bytesReadValues", "0"},
			{"timestampsRead", "0"},
			{"valuesRead", "0"},
			{"bytesProcessedUncompressedValues", "0"},
		},
	})

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
