package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"

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

	// ss must be updated before flush() call.
	ss searchStats
}

func (psp *pipeQueryStatsProcessor) writeBlock(_ uint, _ *blockResult) {
	// Nothing to do
}

func (psp *pipeQueryStatsProcessor) flush() error {
	pipeQueryStatsWriteResult(psp.ppNext, &psp.ss)
	return nil
}

func pipeQueryStatsWriteResult(ppNext pipeProcessor, ss *searchStats) {
	rcs := make([]resultColumn, 10)

	var buf []byte
	addUint64Entry := func(rc *resultColumn, name string, value uint64) {
		rc.name = name
		bufLen := len(buf)
		buf = marshalUint64String(buf, value)
		v := bytesutil.ToUnsafeString(buf[bufLen:])
		rc.addValue(v)
	}

	addUint64Entry(&rcs[0], "bytesReadColumnsHeaders", ss.bytesReadColumnsHeaders)
	addUint64Entry(&rcs[1], "bytesReadColumnsHeaderIndexes", ss.bytesReadColumnsHeaderIndexes)
	addUint64Entry(&rcs[2], "bytesReadBloomFilters", ss.bytesReadBloomFilters)
	addUint64Entry(&rcs[3], "bytesReadValues", ss.bytesReadValues)
	addUint64Entry(&rcs[4], "bytesReadTimestamps", ss.bytesReadTimestamps)
	addUint64Entry(&rcs[5], "bytesReadBlockHeaders", ss.bytesReadBlockHeaders)

	bytesReadTotal := ss.bytesReadColumnsHeaders + ss.bytesReadColumnsHeaderIndexes + ss.bytesReadBloomFilters + ss.bytesReadValues + ss.bytesReadTimestamps + ss.bytesReadBlockHeaders
	addUint64Entry(&rcs[6], "bytesReadTotal", bytesReadTotal)

	addUint64Entry(&rcs[7], "blocksProcessed", ss.blocksProcessed)
	addUint64Entry(&rcs[8], "valuesRead", ss.valuesRead)
	addUint64Entry(&rcs[9], "timestampsRead", ss.timestampsRead)

	var br blockResult
	br.setResultColumns(rcs, 1)
	ppNext.writeBlock(0, &br)
}

func parsePipeQueryStats(lex *lexer) (pipe, error) {
	if !lex.isKeyword("query_stats") {
		return nil, fmt.Errorf("expecting 'query_stats'; got %q", lex.token)
	}
	lex.nextToken()

	ps := &pipeQueryStats{}

	return ps, nil
}
