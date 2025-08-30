package logstorage

import (
	"context"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

// RunNetQueryFunc must run q and pass the query results to writeBlock.
type RunNetQueryFunc func(ctx context.Context, tenantIDs []TenantID, q *Query, writeBlock WriteDataBlockFunc) error

// NetQueryRunner is a runner for distributed query.
type NetQueryRunner struct {
	// qRemote is the query to execute at remote storage nodes.
	qRemote *Query

	// pipesLocal are pipes to execute locally after receiving the data from remote storage nodes.
	pipesLocal []pipe

	// writeBlock is the function for writing the resulting data block.
	writeBlock writeBlockResultFunc
}

// NewNetQueryRunner creates a new NetQueryRunner for the given q.
//
// runNetQuery is used for running distributed query.
// q results are sent to writeNetBlock.
func NewNetQueryRunner(ctx context.Context, tenantIDs []TenantID, q *Query, runNetQuery RunNetQueryFunc, writeNetBlock WriteDataBlockFunc) (*NetQueryRunner, error) {
	runQuery := func(ctx context.Context, tenantIDs []TenantID, q *Query, writeBlock writeBlockResultFunc) error {
		writeNetBlock := writeBlock.newDataBlockWriter()
		return runNetQuery(ctx, tenantIDs, q, writeNetBlock)
	}

	qNew, err := initSubqueries(ctx, tenantIDs, q, runQuery, false)
	if err != nil {
		return nil, err
	}
	q = qNew

	qRemote, pipesLocal := splitQueryToRemoteAndLocal(q)

	writeBlock := writeNetBlock.newBlockResultWriter()

	nqr := &NetQueryRunner{
		qRemote:    qRemote,
		pipesLocal: pipesLocal,
		writeBlock: writeBlock,
	}
	return nqr, nil
}

// Run runs the nqr query.
//
// The concurrency limits the number of concurrent goroutines, which process the query results at the local host.
//
// netSearch must execute the given query q at remote storage nodes and pass results to writeBlock.
func (nqr *NetQueryRunner) Run(ctx context.Context, concurrency int, netSearch func(stopCh <-chan struct{}, q *Query, writeBlock WriteDataBlockFunc) error) error {
	search := func(stopCh <-chan struct{}, writeBlockToPipes writeBlockResultFunc) error {
		writeNetBlock := writeBlockToPipes.newDataBlockWriter()
		return netSearch(stopCh, nqr.qRemote, writeNetBlock)
	}

	ss := &searchStats{}
	return runPipes(ctx, ss, nqr.pipesLocal, search, nqr.writeBlock, concurrency)
}

// splitQueryToRemoteAndLocal splits q into remotely executed query and into locally executed pipes.
func splitQueryToRemoteAndLocal(q *Query) (*Query, []pipe) {
	timestamp := q.GetTimestamp()
	qRemote := q.Clone(timestamp)
	qRemote.DropAllPipes()

	pipesRemote, pipesLocal := getRemoteAndLocalPipes(q)
	qRemote.pipes = pipesRemote

	// Limit fields to select at the remote storage.
	pf := getNeededColumns(pipesLocal)
	qRemote.addFieldsFilters(pf)

	return qRemote, pipesLocal
}

func getRemoteAndLocalPipes(q *Query) ([]pipe, []pipe) {
	timestamp := q.GetTimestamp()

	var pipesRemote []pipe
	var pipesLocal []pipe

	for i, p := range q.pipes {
		if _, ok := p.(*pipeQueryStats); ok {
			// Special case for query_stats pipe: push all the pipes until query_stats to remote side.
			pRemote, psLocal := p.splitToRemoteAndLocal(timestamp)
			pipesRemote = append(pipesRemote[:0], q.pipes[:i]...)
			pipesRemote = append(pipesRemote, pRemote)
			pipesLocal = append(pipesLocal, psLocal...)
			pipesLocal = append(pipesLocal, q.pipes[i+1:]...)
			return pipesRemote, pipesLocal
		}
	}

	for i, p := range q.pipes {
		pRemote, psLocal := p.splitToRemoteAndLocal(timestamp)
		if pRemote != nil {
			pipesRemote = append(pipesRemote, pRemote)
			if len(psLocal) == 0 {
				continue
			}
		}

		if len(psLocal) == 0 {
			logger.Panicf("BUG: psLocal must be non non-empty here")
		}

		pipesLocal = append(pipesLocal, psLocal...)
		pipesLocal = append(pipesLocal, q.pipes[i+1:]...)
		return pipesRemote, pipesLocal
	}

	return nil, nil
}
