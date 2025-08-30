package logstorage

import (
	"sync/atomic"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"
)

// QueryStats contains various query execution stats.
type QueryStats struct {
	// BytesReadColumnsHeaders is the total number of columns header bytes read from disk during the search.
	BytesReadColumnsHeaders uint64

	// BytesReadColumnsHeaderIndexes is the total number of columns header index bytes read from disk during the search.
	BytesReadColumnsHeaderIndexes uint64

	// BytesReadBloomFilters is the total number of bloom filter bytes read from disk during the search.
	BytesReadBloomFilters uint64

	// BytesReadValues is the total number of values bytes read from disk during the search.
	BytesReadValues uint64

	// BytesReadTimestamps is the total number of timestamps bytes read from disk during the search.
	BytesReadTimestamps uint64

	// BytesReadBlockHeaders is the total number of headers bytes read from disk during the search.
	BytesReadBlockHeaders uint64

	// BlocksProcessed is the number of data blocks processed during query execution.
	BlocksProcessed uint64

	// RowsProcessed is the number of log rows processed during query execution.
	RowsProcessed uint64

	// ValuesRead is the number of log field values read during query exection.
	ValuesRead uint64

	// TimestampsRead is the number of timestamps read during query execution.
	TimestampsRead uint64

	// BytesProcessedUncompressedValues is the total number of uncompressed values bytes processed during the search.
	BytesProcessedUncompressedValues uint64
}

// GetBytesReadTotal returns the total number of bytes read, which is tracked by qs.
func (qs *QueryStats) GetBytesReadTotal() uint64 {
	return qs.BytesReadColumnsHeaders + qs.BytesReadColumnsHeaderIndexes + qs.BytesReadBloomFilters + qs.BytesReadValues + qs.BytesReadTimestamps + qs.BytesReadBlockHeaders
}

func (qs *QueryStats) updateAtomic(src *QueryStats) {
	atomic.AddUint64(&qs.BytesReadColumnsHeaders, src.BytesReadColumnsHeaders)
	atomic.AddUint64(&qs.BytesReadColumnsHeaderIndexes, src.BytesReadColumnsHeaderIndexes)
	atomic.AddUint64(&qs.BytesReadBloomFilters, src.BytesReadBloomFilters)
	atomic.AddUint64(&qs.BytesReadValues, src.BytesReadValues)
	atomic.AddUint64(&qs.BytesReadTimestamps, src.BytesReadTimestamps)
	atomic.AddUint64(&qs.BytesReadTimestamps, src.BytesReadTimestamps)
	atomic.AddUint64(&qs.BytesReadBlockHeaders, src.BytesReadBlockHeaders)

	atomic.AddUint64(&qs.BlocksProcessed, src.BlocksProcessed)
	atomic.AddUint64(&qs.RowsProcessed, src.RowsProcessed)
	atomic.AddUint64(&qs.ValuesRead, src.ValuesRead)
	atomic.AddUint64(&qs.TimestampsRead, src.TimestampsRead)
	atomic.AddUint64(&qs.BytesProcessedUncompressedValues, src.BytesProcessedUncompressedValues)
}

func pipeQueryStatsWriteResult(ppNext pipeProcessor, qs *QueryStats, queryDurationNsecs int64) {
	var rcs []resultColumn

	var buf []byte
	addUint64Entry := func(name string, value uint64) {
		rcs = append(rcs, resultColumn{})
		rc := &rcs[len(rcs)-1]
		rc.name = name
		bufLen := len(buf)
		buf = marshalUint64String(buf, value)
		v := bytesutil.ToUnsafeString(buf[bufLen:])
		rc.addValue(v)
	}

	addUint64Entry("BytesReadColumnsHeaders", qs.BytesReadColumnsHeaders)
	addUint64Entry("BytesReadColumnsHeaderIndexes", qs.BytesReadColumnsHeaderIndexes)
	addUint64Entry("BytesReadBloomFilters", qs.BytesReadBloomFilters)
	addUint64Entry("BytesReadValues", qs.BytesReadValues)
	addUint64Entry("BytesReadTimestamps", qs.BytesReadTimestamps)
	addUint64Entry("BytesReadBlockHeaders", qs.BytesReadBlockHeaders)

	addUint64Entry("BytesReadTotal", qs.GetBytesReadTotal())

	addUint64Entry("BlocksProcessed", qs.BlocksProcessed)
	addUint64Entry("RowsProcessed", qs.RowsProcessed)
	addUint64Entry("ValuesRead", qs.ValuesRead)
	addUint64Entry("TimestampsRead", qs.TimestampsRead)
	addUint64Entry("BytesProcessedUncompressedValues", qs.BytesProcessedUncompressedValues)

	addUint64Entry("QueryDurationNsecs", uint64(queryDurationNsecs))

	var br blockResult
	br.setResultColumns(rcs, 1)
	ppNext.writeBlock(0, &br)
}

func pipeQueryStatsUpdateAtomic(dst *QueryStats, br *blockResult) {
	getUint64Entry := func(name string) uint64 {
		c := br.getColumnByName(name)
		v := c.getValueAtRow(br, 0)
		n, _ := tryParseUint64(v)
		return n
	}

	var qs QueryStats

	qs.BytesReadColumnsHeaders = getUint64Entry("BytesReadColumnsHeaders")
	qs.BytesReadColumnsHeaderIndexes = getUint64Entry("BytesReadColumnsHeaderIndexes")
	qs.BytesReadBloomFilters = getUint64Entry("BytesReadBloomFilters")
	qs.BytesReadValues = getUint64Entry("BytesReadValues")
	qs.BytesReadTimestamps = getUint64Entry("BytesReadTimestamps")
	qs.BytesReadBlockHeaders = getUint64Entry("BytesReadBlockHeaders")

	qs.BlocksProcessed = getUint64Entry("BlocksProcessed")
	qs.RowsProcessed = getUint64Entry("RowsProcessed")
	qs.ValuesRead = getUint64Entry("ValuesRead")
	qs.TimestampsRead = getUint64Entry("TimestampsRead")
	qs.BytesProcessedUncompressedValues = getUint64Entry("BytesProcessedUncompressedValues")

	dst.updateAtomic(&qs)
}
