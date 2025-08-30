package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeQueryStatsLocal processes local part of the pipeQueryStats in cluster.
type pipeQueryStatsLocal struct {
}

func (ps *pipeQueryStatsLocal) String() string {
	return "query_stats_local"
}

func (ps *pipeQueryStatsLocal) splitToRemoteAndLocal(_ int64) (pipe, []pipe) {
	logger.Panicf("BUG: unexpected call for %T", ps)
	return nil, nil
}

func (ps *pipeQueryStatsLocal) canLiveTail() bool {
	return false
}

func (ps *pipeQueryStatsLocal) canReturnLastNResults() bool {
	return false
}

func (ps *pipeQueryStatsLocal) updateNeededFields(_ *prefixfilter.Filter) {
	// Nothing to do
}

func (ps *pipeQueryStatsLocal) hasFilterInWithQuery() bool {
	return false
}

func (ps *pipeQueryStatsLocal) initFilterInValues(_ *inValuesCache, _ getFieldValuesFunc, _ bool) (pipe, error) {
	return ps, nil
}

func (ps *pipeQueryStatsLocal) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func (ps *pipeQueryStatsLocal) newPipeProcessor(_ int, stopCh <-chan struct{}, _ func(), ppNext pipeProcessor) pipeProcessor {
	psp := &pipeQueryStatsLocalProcessor{
		ppNext: ppNext,
	}
	return psp
}

type pipeQueryStatsLocalProcessor struct {
	ppNext pipeProcessor

	qs QueryStats

	// queryDurationNsecs must be initialized before flush() call via setStartTime().
	queryDurationNsecs int64
}

func (psp *pipeQueryStatsLocalProcessor) setStartTime(queryDurationNsecs int64) {
	psp.queryDurationNsecs = queryDurationNsecs
}

func (psp *pipeQueryStatsLocalProcessor) writeBlock(_ uint, br *blockResult) {
	if br.rowsLen <= 0 {
		return
	}
	pipeQueryStatsUpdateAtomic(&psp.qs, br)
}

func (psp *pipeQueryStatsLocalProcessor) flush() error {
	pipeQueryStatsWriteResult(psp.ppNext, &psp.qs, psp.queryDurationNsecs)
	return nil
}
