package txindex_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/evdatsion/aphelion-dpos-bft/abci/types"
	"github.com/evdatsion/aphelion-dpos-bft/libs/db"
	"github.com/evdatsion/aphelion-dpos-bft/libs/log"
	"github.com/evdatsion/aphelion-dpos-bft/state/txindex"
	"github.com/evdatsion/aphelion-dpos-bft/state/txindex/kv"
	"github.com/evdatsion/aphelion-dpos-bft/types"
)

func TestIndexerServiceIndexesBlocks(t *testing.T) {
	// event bus
	eventBus := types.NewEventBus()
	eventBus.SetLogger(log.TestingLogger())
	err := eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// tx indexer
	store := db.NewMemDB()
	txIndexer := kv.NewTxIndex(store, kv.IndexAllTags())

	service := txindex.NewIndexerService(txIndexer, eventBus)
	service.SetLogger(log.TestingLogger())
	err = service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// publish block with txs
	eventBus.PublishEventNewBlockHeader(types.EventDataNewBlockHeader{
		Header: types.Header{Height: 1, NumTxs: 2},
	})
	txResult1 := &types.TxResult{
		Height: 1,
		Index:  uint32(0),
		Tx:     types.Tx("foo"),
		Result: abci.ResponseDeliverTx{Code: 0},
	}
	eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult1})
	txResult2 := &types.TxResult{
		Height: 1,
		Index:  uint32(1),
		Tx:     types.Tx("bar"),
		Result: abci.ResponseDeliverTx{Code: 0},
	}
	eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult2})

	time.Sleep(100 * time.Millisecond)

	// check the result
	res, err := txIndexer.Get(types.Tx("foo").Hash())
	assert.NoError(t, err)
	assert.Equal(t, txResult1, res)
	res, err = txIndexer.Get(types.Tx("bar").Hash())
	assert.NoError(t, err)
	assert.Equal(t, txResult2, res)
}
