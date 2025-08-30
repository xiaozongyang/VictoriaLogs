package logstorage

import (
	"sync"
	"sync/atomic"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"
	"github.com/VictoriaMetrics/metrics"
)

var (
	searchStatsOnce sync.Once

	bytesReadPerQueryColumnsHeaders       *metrics.Histogram
	bytesReadPerQueryColumnsHeaderIndexes *metrics.Histogram
	bytesReadPerQueryBloomFilters         *metrics.Histogram
	bytesReadPerQueryValues               *metrics.Histogram
	bytesReadPerQueryTimestamps           *metrics.Histogram
	bytesReadPerQueryBlockHeaders         *metrics.Histogram

	bytesReadPerQueryTotal *metrics.Histogram

	blocksProcessedPerQuery *metrics.Histogram

	valuesReadPerQuery     *metrics.Histogram
	timestampsReadPerQuery *metrics.Histogram
)

func initSearchStats() {
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
}

func updateSearchMetrics(ss *searchStats) {
	searchStatsOnce.Do(initSearchStats)

	bytesReadPerQueryColumnsHeaders.Update(float64(ss.bytesReadColumnsHeaders))
	bytesReadPerQueryColumnsHeaderIndexes.Update(float64(ss.bytesReadColumnsHeaderIndexes))
	bytesReadPerQueryBloomFilters.Update(float64(ss.bytesReadBloomFilters))
	bytesReadPerQueryValues.Update(float64(ss.bytesReadValues))
	bytesReadPerQueryTimestamps.Update(float64(ss.bytesReadTimestamps))
	bytesReadPerQueryBlockHeaders.Update(float64(ss.bytesReadBlockHeaders))

	bytesReadTotal := ss.bytesReadColumnsHeaders + ss.bytesReadColumnsHeaderIndexes + ss.bytesReadBloomFilters +
		ss.bytesReadValues + ss.bytesReadTimestamps + ss.bytesReadBlockHeaders
	bytesReadPerQueryTotal.Update(float64(bytesReadTotal))

	blocksProcessedPerQuery.Update(float64(ss.blocksProcessed))

	valuesReadPerQuery.Update(float64(ss.valuesRead))
	timestampsReadPerQuery.Update(float64(ss.timestampsRead))
}

// searchStats contains various stats related to the search.
type searchStats struct {
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
}

func (ss *searchStats) updateAtomic(src *searchStats) {
	atomic.AddUint64(&ss.bytesReadColumnsHeaders, src.bytesReadColumnsHeaders)
	atomic.AddUint64(&ss.bytesReadColumnsHeaderIndexes, src.bytesReadColumnsHeaderIndexes)
	atomic.AddUint64(&ss.bytesReadBloomFilters, src.bytesReadBloomFilters)
	atomic.AddUint64(&ss.bytesReadValues, src.bytesReadValues)
	atomic.AddUint64(&ss.bytesReadTimestamps, src.bytesReadTimestamps)
	atomic.AddUint64(&ss.bytesReadBlockHeaders, src.bytesReadBlockHeaders)

	atomic.AddUint64(&ss.blocksProcessed, src.blocksProcessed)

	atomic.AddUint64(&ss.valuesRead, src.valuesRead)
	atomic.AddUint64(&ss.timestampsRead, src.timestampsRead)
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

func pipeQueryStatsUpdate(ss *searchStats, br *blockResult) {
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
