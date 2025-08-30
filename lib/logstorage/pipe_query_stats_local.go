package logstorage

import (
	"sync"

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

	ss     searchStats
	ssLock sync.Mutex
}

func (psp *pipeQueryStatsLocalProcessor) writeBlock(_ uint, br *blockResult) {
	if br.rowsLen <= 0 {
		return
	}

	psp.ssLock.Lock()
	defer psp.ssLock.Unlock()

	ss := &psp.ss

	updateUint64Entry := func(dst *uint64, name string) {
		c := br.getColumnByName(name)
		v := c.getValueAtRow(br, 0)
		n, ok := tryParseUint64(v)
		if ok {
			*dst += n
		}
	}

	updateUint64Entry(&ss.bytesReadColumnsHeaders, "bytesReadColumnsHeaders")
	updateUint64Entry(&ss.bytesReadColumnsHeaderIndexes, "bytesReadColumnsHeaderIndexes")
	updateUint64Entry(&ss.bytesReadBloomFilters, "bytesReadBloomFilters")
	updateUint64Entry(&ss.bytesReadValues, "bytesReadValues")
	updateUint64Entry(&ss.bytesReadTimestamps, "bytesReadTimestamps")
	updateUint64Entry(&ss.bytesReadBlockHeaders, "bytesReadBlockHeaders")
	updateUint64Entry(&ss.blocksProcessed, "blocksProcessed")
	updateUint64Entry(&ss.valuesRead, "valuesRead")
	updateUint64Entry(&ss.timestampsRead, "timestampsRead")
}

func (psp *pipeQueryStatsLocalProcessor) flush() error {
	pipeQueryStatsWriteResult(psp.ppNext, &psp.ss)
	return nil
}
