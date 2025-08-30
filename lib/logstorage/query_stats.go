package logstorage

import (
	"sync"
	"sync/atomic"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"
	"github.com/VictoriaMetrics/metrics"
)

var (
	queryStatsOnce sync.Once

	bytesReadPerQueryColumnsHeaders       *metrics.Histogram
	bytesReadPerQueryColumnsHeaderIndexes *metrics.Histogram
	bytesReadPerQueryBloomFilters         *metrics.Histogram
	bytesReadPerQueryValues               *metrics.Histogram
	bytesReadPerQueryTimestamps           *metrics.Histogram
	bytesReadPerQueryBlockHeaders         *metrics.Histogram

	bytesReadPerQueryTotal *metrics.Histogram

	blocksProcessedPerQuery                  *metrics.Histogram
	valuesReadPerQuery                       *metrics.Histogram
	timestampsReadPerQuery                   *metrics.Histogram
	bytesProcessedPerQueryUncompressedValues *metrics.Histogram
)

func initQueryStats() {
	bytesReadPerQueryColumnsHeaders = metrics.NewHistogram(`vl_storage_per_query_columns_headers_read_bytes`)
	bytesReadPerQueryColumnsHeaderIndexes = metrics.NewHistogram(`vl_storage_per_query_columns_header_indexes_read_bytes`)
	bytesReadPerQueryBloomFilters = metrics.NewHistogram(`vl_storage_per_query_bloom_filters_read_bytes`)
	bytesReadPerQueryValues = metrics.NewHistogram(`vl_storage_per_query_values_read_bytes`)
	bytesReadPerQueryTimestamps = metrics.NewHistogram(`vl_storage_per_query_timestamps_read_bytes`)
	bytesReadPerQueryBlockHeaders = metrics.NewHistogram(`vl_storage_per_query_block_headers_read_bytes`)

	bytesReadPerQueryTotal = metrics.NewHistogram(`vl_storage_per_query_total_read_bytes`)

	blocksProcessedPerQuery = metrics.NewHistogram(`vl_storage_per_query_processed_blocks`)
	valuesReadPerQuery = metrics.NewHistogram(`vl_storage_per_query_read_values`)
	timestampsReadPerQuery = metrics.NewHistogram(`vl_storage_per_query_read_timestamps`)
	bytesProcessedPerQueryUncompressedValues = metrics.NewHistogram(`vl_storage_per_query_uncompressed_values_processed_bytes`)
}

func updateQueryStatsMetrics(qs *queryStats) {
	queryStatsOnce.Do(initQueryStats)

	bytesReadPerQueryColumnsHeaders.Update(float64(qs.bytesReadColumnsHeaders))
	bytesReadPerQueryColumnsHeaderIndexes.Update(float64(qs.bytesReadColumnsHeaderIndexes))
	bytesReadPerQueryBloomFilters.Update(float64(qs.bytesReadBloomFilters))
	bytesReadPerQueryValues.Update(float64(qs.bytesReadValues))
	bytesReadPerQueryTimestamps.Update(float64(qs.bytesReadTimestamps))
	bytesReadPerQueryBlockHeaders.Update(float64(qs.bytesReadBlockHeaders))

	bytesReadPerQueryTotal.Update(float64(qs.getBytesReadTotal()))

	blocksProcessedPerQuery.Update(float64(qs.blocksProcessed))
	valuesReadPerQuery.Update(float64(qs.valuesRead))
	timestampsReadPerQuery.Update(float64(qs.timestampsRead))
	bytesProcessedPerQueryUncompressedValues.Update(float64(qs.bytesProcessedUncompressedValues))
}

// queryStats contains various query execution stats.
type queryStats struct {
	// bytesReadColumnsHeaders is the total number of columns header bytes read from disk during the search.
	bytesReadColumnsHeaders uint64

	// bytesReadColumnsHeaderIndexes is the total number of columns header index bytes read from disk during the search.
	bytesReadColumnsHeaderIndexes uint64

	// bytesReadBloomFilters is the total number of bloom filter bytes read from disk during the search.
	bytesReadBloomFilters uint64

	// bytesReadValues is the total number of values bytes read from disk during the search.
	bytesReadValues uint64

	// bytesReadTimestamps is the total number of timestamps bytes read from disk during the search.
	bytesReadTimestamps uint64

	// bytesReadBlockHeaders is the total number of headers bytes read from disk during the search.
	bytesReadBlockHeaders uint64

	// blocksProcessed is the number of data blocks processed during query execution.
	blocksProcessed uint64

	// valuesRead is the number of log field values read during query exection.
	valuesRead uint64

	// timestampsRead is the number of timestamps read during query execution.
	timestampsRead uint64

	// bytesProcessedUncompressedValues is the total number of uncompressed values bytes processed during the search.
	bytesProcessedUncompressedValues uint64
}

func (qs *queryStats) getBytesReadTotal() uint64 {
	return qs.bytesReadColumnsHeaders + qs.bytesReadColumnsHeaderIndexes + qs.bytesReadBloomFilters + qs.bytesReadValues + qs.bytesReadTimestamps + qs.bytesReadBlockHeaders
}

func (qs *queryStats) updateAtomic(src *queryStats) {
	atomic.AddUint64(&qs.bytesReadColumnsHeaders, src.bytesReadColumnsHeaders)
	atomic.AddUint64(&qs.bytesReadColumnsHeaderIndexes, src.bytesReadColumnsHeaderIndexes)
	atomic.AddUint64(&qs.bytesReadBloomFilters, src.bytesReadBloomFilters)
	atomic.AddUint64(&qs.bytesReadValues, src.bytesReadValues)
	atomic.AddUint64(&qs.bytesReadTimestamps, src.bytesReadTimestamps)
	atomic.AddUint64(&qs.bytesReadTimestamps, src.bytesReadTimestamps)
	atomic.AddUint64(&qs.bytesReadBlockHeaders, src.bytesReadBlockHeaders)

	atomic.AddUint64(&qs.blocksProcessed, src.blocksProcessed)
	atomic.AddUint64(&qs.valuesRead, src.valuesRead)
	atomic.AddUint64(&qs.timestampsRead, src.timestampsRead)
	atomic.AddUint64(&qs.bytesProcessedUncompressedValues, src.bytesProcessedUncompressedValues)
}

func pipeQueryStatsWriteResult(ppNext pipeProcessor, qs *queryStats) {
	rcs := make([]resultColumn, 11)

	var buf []byte
	addUint64Entry := func(rc *resultColumn, name string, value uint64) {
		rc.name = name
		bufLen := len(buf)
		buf = marshalUint64String(buf, value)
		v := bytesutil.ToUnsafeString(buf[bufLen:])
		rc.addValue(v)
	}

	addUint64Entry(&rcs[0], "bytesReadColumnsHeaders", qs.bytesReadColumnsHeaders)
	addUint64Entry(&rcs[1], "bytesReadColumnsHeaderIndexes", qs.bytesReadColumnsHeaderIndexes)
	addUint64Entry(&rcs[2], "bytesReadBloomFilters", qs.bytesReadBloomFilters)
	addUint64Entry(&rcs[3], "bytesReadValues", qs.bytesReadValues)
	addUint64Entry(&rcs[4], "bytesReadTimestamps", qs.bytesReadTimestamps)
	addUint64Entry(&rcs[5], "bytesReadBlockHeaders", qs.bytesReadBlockHeaders)

	addUint64Entry(&rcs[6], "bytesReadTotal", qs.getBytesReadTotal())

	addUint64Entry(&rcs[7], "blocksProcessed", qs.blocksProcessed)
	addUint64Entry(&rcs[8], "valuesRead", qs.valuesRead)
	addUint64Entry(&rcs[9], "timestampsRead", qs.timestampsRead)
	addUint64Entry(&rcs[10], "bytesProcessedUncompressedValues", qs.bytesProcessedUncompressedValues)

	var br blockResult
	br.setResultColumns(rcs, 1)
	ppNext.writeBlock(0, &br)
}

func pipeQueryStatsUpdateAtomic(dst *queryStats, br *blockResult) {
	getUint64Entry := func(name string) uint64 {
		c := br.getColumnByName(name)
		v := c.getValueAtRow(br, 0)
		n, _ := tryParseUint64(v)
		return n
	}

	var qs queryStats

	qs.bytesReadColumnsHeaders = getUint64Entry("bytesReadColumnsHeaders")
	qs.bytesReadColumnsHeaderIndexes = getUint64Entry("bytesReadColumnsHeaderIndexes")
	qs.bytesReadBloomFilters = getUint64Entry("bytesReadBloomFilters")
	qs.bytesReadValues = getUint64Entry("bytesReadValues")
	qs.bytesReadTimestamps = getUint64Entry("bytesReadTimestamps")
	qs.bytesReadBlockHeaders = getUint64Entry("bytesReadBlockHeaders")

	qs.blocksProcessed = getUint64Entry("blocksProcessed")
	qs.valuesRead = getUint64Entry("valuesRead")
	qs.timestampsRead = getUint64Entry("timestampsRead")
	qs.bytesProcessedUncompressedValues = getUint64Entry("bytesProcessedUncompressedValues")

	dst.updateAtomic(&qs)
}
