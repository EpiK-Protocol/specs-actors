package expert_test

import (
	"context"
	"testing"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/expert"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
	"github.com/filecoin-project/specs-actors/v2/support/ipld"
	tutils "github.com/filecoin-project/specs-actors/v2/support/testing"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestData(t *testing.T) {
	t.Run("Put, get and delete", func(t *testing.T) {
		harness := constructStateHarness(t)

		pieceID := tutils.MakeCID("1", &miner.SealedCIDPrefix)
		data := newDataOnChainInfo(pieceID)
		harness.putData(data)
		assert.Equal(t, data, harness.getData(pieceID))

		pieceID = tutils.MakeCID("2", &miner.SealedCIDPrefix)
		data = newDataOnChainInfo(pieceID)
		harness.putData(data)
		assert.Equal(t, data, harness.getData(pieceID))

		harness.deleteData(pieceID)
		assert.False(t, harness.hasData(pieceID))
	})

	t.Run("Delete nonexistent value returns an error", func(t *testing.T) {
		harness := constructStateHarness(t)

		pieceID := tutils.MakeCID("1", &miner.SealedCIDPrefix)
		err := harness.s.DeleteData(harness.store, pieceID.String())
		assert.Error(t, err)
	})

	t.Run("Get nonexistent value returns false", func(t *testing.T) {
		harness := constructStateHarness(t)

		pieceID := tutils.MakeCID("1", &miner.SealedCIDPrefix)
		assert.False(t, harness.hasData(pieceID))
	})
}

type stateHarness struct {
	t testing.TB

	s     *expert.State
	store adt.Store
}

func (h *stateHarness) putData(data *expert.DataOnChainInfo) {
	err := h.s.PutData(h.store, data)
	require.NoError(h.t, err)
}

func (h *stateHarness) getData(pieceID cid.Cid) *expert.DataOnChainInfo {
	data, _, err := h.s.GetData(h.store, pieceID.String())
	require.NoError(h.t, err)
	return data
}

func (h *stateHarness) deleteData(pieceID cid.Cid) {
	err := h.s.DeleteData(h.store, pieceID.String())
	require.NoError(h.t, err)
}

func (h *stateHarness) hasData(pieceID cid.Cid) bool {
	_, found, err := h.s.GetData(h.store, pieceID.String())
	require.NoError(h.t, err)
	return found
}

func constructStateHarness(t *testing.T) *stateHarness {
	// store init
	store := ipld.NewADTStore(context.Background())
	emptyMap, err := adt.MakeEmptyMap(store).Root()
	require.NoError(t, err)

	// state field init
	owner := tutils.NewBLSAddr(t, 1)
	state := expert.ConstructState(emptyMap, owner, abi.PeerID("peer"), testMultiaddrs)

	return &stateHarness{
		t:     t,
		s:     state,
		store: store,
	}
}

func newDataOnChainInfo(pieceID cid.Cid) *expert.DataOnChainInfo {
	return &expert.DataOnChainInfo{
		PieceID: pieceID.String(),
	}
}