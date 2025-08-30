package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeQueryStats implements '| query_stats' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#query_stats-pipe
type pipeQueryStats struct {
}

func (ps *pipeQueryStats) String() string {
	return "query_stats"
}

func (ps *pipeQueryStats) splitToRemoteAndLocal(_ int64) (pipe, []pipe) {
	psLocal := &pipeQueryStatsLocal{}
	return ps, []pipe{psLocal}
}

func (ps *pipeQueryStats) canLiveTail() bool {
	return false
}

func (ps *pipeQueryStats) canReturnLastNResults() bool {
	return false
}

func (ps *pipeQueryStats) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("*")
}

func (ps *pipeQueryStats) hasFilterInWithQuery() bool {
	return false
}

func (ps *pipeQueryStats) initFilterInValues(_ *inValuesCache, _ getFieldValuesFunc, _ bool) (pipe, error) {
	return ps, nil
}

func (ps *pipeQueryStats) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func (ps *pipeQueryStats) newPipeProcessor(_ int, _ <-chan struct{}, _ func(), ppNext pipeProcessor) pipeProcessor {
	psp := &pipeQueryStatsProcessor{
		ps:     ps,
		ppNext: ppNext,
	}
	return psp
}

type pipeQueryStatsProcessor struct {
	ps     *pipeQueryStats
	ppNext pipeProcessor

	// qs must be initialized via setQueryStats() before flush() call.
	qs queryStats
}

func (psp *pipeQueryStatsProcessor) setQueryStats(qs *queryStats) {
	psp.qs = *qs
}

func (psp *pipeQueryStatsProcessor) writeBlock(_ uint, _ *blockResult) {
	// Nothing to do
}

func (psp *pipeQueryStatsProcessor) flush() error {
	pipeQueryStatsWriteResult(psp.ppNext, &psp.qs)
	return nil
}

func parsePipeQueryStats(lex *lexer) (pipe, error) {
	if !lex.isKeyword("query_stats") {
		return nil, fmt.Errorf("expecting 'query_stats'; got %q", lex.token)
	}
	lex.nextToken()

	ps := &pipeQueryStats{}

	return ps, nil
}
