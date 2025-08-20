package logstorage

import (
	"testing"
)

func TestParseGenerateSequenceSuccess(t *testing.T) {
	f := func(pipeStr string) {
		t.Helper()
		expectParsePipeSuccess(t, pipeStr)
	}

	f(`generate_sequence 1`)
	f(`generate_sequence 123456789`)
}

func TestParsePipeGenerateSequenceFailure(t *testing.T) {
	f := func(pipeStr string) {
		t.Helper()
		expectParsePipeFailure(t, pipeStr)
	}

	f(`generate_sequence`)
	f(`generate_sequence 0`)
	f(`generate_sequence -123`)
	f(`generate_sequence foo`)
}

func TestPipeGenerateSequence(t *testing.T) {
	f := func(pipeStr string, rows, rowsExpected [][]Field) {
		t.Helper()
		expectPipeResults(t, pipeStr, rows, rowsExpected)
	}

	// non-empty input
	f(`generate_sequence 3`, [][]Field{
		{
			{"foo", `bar`},
			{"bar", `abcde`},
		},
	}, [][]Field{
		{
			{"_msg", `0`},
		},
		{
			{"_msg", `1`},
		},
		{
			{"_msg", `2`},
		},
	})
	f(`generate_sequence 1`, [][]Field{
		{
			{"foo", `bar`},
			{"bar", `abcde`},
		},
		{
			{"x", `y`},
		},
	}, [][]Field{
		{
			{"_msg", `0`},
		},
	})

	// empty input
	f(`generate_sequence 3`, [][]Field{}, [][]Field{
		{
			{"_msg", `0`},
		},
		{
			{"_msg", `1`},
		},
		{
			{"_msg", `2`},
		},
	})
}

func TestPipeGenerateSequenceUpdateNeededFields(t *testing.T) {
	f := func(s string, allowFilters, denyFilters, allowFiltersExpected, denyFiltersExpected string) {
		t.Helper()
		expectPipeNeededFields(t, s, allowFilters, denyFilters, allowFiltersExpected, denyFiltersExpected)
	}

	// all the needed fields
	f("generate_sequence 12", "*", "", "", "")

	// all the needed fields, unneeded fields do not intersect with _msg
	f("generate_sequence 34", "*", "f1,f2", "", "")

	// all the needed fields, unneeded fields intersect with _msg
	f("generate_sequence 45", "*", "_msg,f1,f2", "", "")

	// needed fields do not intersect with _msg
	f("generate_sequence 1", "f1,f2", "", "", "")

	// needed fields intersect with _msg
	f("generate_sequence 2", "_msg,f1,f2", "", "", "")
}
