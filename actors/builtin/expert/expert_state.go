package expert

import (
	addr "github.com/filecoin-project/go-address"
	cid "github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	xerrors "golang.org/x/xerrors"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
)

// State of expert
type State struct {
	// Information not related to sectors.
	Info cid.Cid

	// Information for all submit rdf data.
	Datas cid.Cid // Map, AMT[key]DataOnChainInfo (sparse)

	// VoteAmount expert vote amount
	VoteAmount abi.TokenAmount

	// LostEpoch record expert votes <  epoch
	LostEpoch abi.ChainEpoch

	// Status of expert
	Status ExpertState

	// OwnerChange owner change info
	OwnerChange cid.Cid
}

// PendingOwnerChange pending owner change
type PendingOwnerChange struct {
	ApplyEpoch abi.ChainEpoch
	ApplyOwner addr.Address
}

// ExpertInfo expert info
type ExpertInfo struct {

	// Account that owns this expert.
	// - Income and returned collateral are paid to this address.
	Owner addr.Address // Must be an ID-address.

	// Byte array representing a Libp2p identity that should be used when connecting to this miner.
	PeerId abi.PeerID

	// Slice of byte arrays representing Libp2p multi-addresses used for establishing a connection with this miner.
	Multiaddrs []abi.Multiaddrs

	// Type expert type
	Type ExpertType

	// ApplicationHash expert application hash
	ApplicationHash string

	// Proposer of expert
	Proposer addr.Address
}

type DataOnChainInfo struct {
	PieceID    string
	PieceSize  abi.PaddedPieceSize
	Redundancy uint64
	Bounty     string
}

func ConstructExpertInfo(owner addr.Address, pid []byte, multiAddrs [][]byte, eType ExpertType, aHash string) (*ExpertInfo, error) {
	return &ExpertInfo{
		Owner:           owner,
		PeerId:          pid,
		Multiaddrs:      multiAddrs,
		Type:            eType,
		ApplicationHash: aHash,
	}, nil
}

func ConstructState(info cid.Cid, emptyMapCid cid.Cid, emptyChange cid.Cid) *State {
	return &State{
		Info:        info,
		Datas:       emptyMapCid,
		VoteAmount:  abi.NewTokenAmount(0),
		LostEpoch:   abi.ChainEpoch(-1),
		Status:      ExpertStateRegistered,
		OwnerChange: emptyChange,
	}
}

func (st *State) GetInfo(store adt.Store) (*ExpertInfo, error) {
	var info ExpertInfo
	if err := store.Get(store.Context(), st.Info, &info); err != nil {
		return nil, xerrors.Errorf("failed to get expert info %w", err)
	}
	return &info, nil
}

func (st *State) SaveInfo(store adt.Store, info *ExpertInfo) error {
	c, err := store.Put(store.Context(), info)
	if err != nil {
		return err
	}
	st.Info = c
	return nil
}

func (st *State) HasDataID(store adt.Store, pieceID string) (bool, error) {
	pieces, err := adt.AsMap(store, st.Datas)
	if err != nil {
		return false, err
	}

	var info DataOnChainInfo
	found, err := pieces.Get(adt.StringKey(pieceID), &info)
	if err != nil {
		return false, xerrors.Errorf("failed to get data %v: %w", pieceID, err)
	}
	return found, nil
}

func (st *State) PutData(store adt.Store, data *DataOnChainInfo) error {
	datas, err := adt.AsMap(store, st.Datas)
	if err != nil {
		return err
	}

	if err := datas.Put(adt.StringKey(data.PieceID), data); err != nil {
		return errors.Wrapf(err, "failed to put data %v", data)
	}
	st.Datas, err = datas.Root()
	return err
}

func (st *State) GetData(store adt.Store, pieceID string) (*DataOnChainInfo, bool, error) {
	datas, err := adt.AsMap(store, st.Datas)
	if err != nil {
		return nil, false, err
	}

	var info DataOnChainInfo
	found, err := datas.Get(adt.StringKey(pieceID), &info)
	if err != nil {
		return nil, false, errors.Wrapf(err, "failed to get data %v", pieceID)
	}
	return &info, found, nil
}

func (st *State) DeleteData(store adt.Store, pieceID string) error {
	datas, err := adt.AsMap(store, st.Datas)
	if err != nil {
		return err
	}
	err = datas.Delete(adt.StringKey(pieceID))
	if err != nil {
		return errors.Wrapf(err, "failed to delete data for %v", pieceID)
	}

	st.Datas, err = datas.Root()
	return err
}

func (st *State) ForEachData(store adt.Store, f func(*DataOnChainInfo)) error {
	datas, err := adt.AsMap(store, st.Datas)
	if err != nil {
		return err
	}
	var info DataOnChainInfo
	return datas.ForEach(&info, func(key string) error {
		f(&info)
		return nil
	})
	return nil
}

func (st *State) GetOwnerChange(rt Runtime) (*PendingOwnerChange, error) {

	var change PendingOwnerChange
	store := adt.AsStore(rt)
	if err := store.Get(store.Context(), st.OwnerChange, &change); err != nil {
		return nil, xerrors.Errorf("failed to get owner change %w", err)
	}
	return &change, nil
}

func (st *State) ApplyOwnerChange(rt Runtime, applyOwner addr.Address) error {
	change := &PendingOwnerChange{
		ApplyEpoch: rt.CurrEpoch(),
		ApplyOwner: applyOwner,
	}
	store := adt.AsStore(rt)
	c, err := store.Put(store.Context(), change)
	if err != nil {
		return err
	}
	st.OwnerChange = c
	return nil
}

func (st *State) AutoUpdateOwnerChange(rt Runtime) error {
	info, err := st.GetInfo(adt.AsStore(rt))
	if err != nil {
		return err
	}

	change, err := st.GetOwnerChange(rt)
	if err != nil {
		return err
	}
	if info.Owner != change.ApplyOwner &&
		change.ApplyEpoch > 0 &&
		(rt.CurrEpoch()-change.ApplyEpoch) >= ExpertVoteCheckPeriod {
		info.Owner = change.ApplyOwner
		if err := st.SaveInfo(adt.AsStore(rt), info); err != nil {
			return err
		}
	}
	return nil
}

func (st *State) Validate(rt Runtime) error {
	switch st.Status {
	case ExpertStateNormal:
		if st.VoteAmount.LessThan(ExpertVoteThreshold) {
			if st.LostEpoch < 0 {
				return xerrors.Errorf("failed to vaildate expert with below vote:%w", st.VoteAmount)
			} else if (st.LostEpoch + ExpertVoteCheckPeriod) < rt.CurrEpoch() {
				return xerrors.Errorf("failed to vaildate expert with lost vote:%w", st.VoteAmount)
			}
		}
	case ExpertStateImplicated:
		if st.VoteAmount.LessThan(ExpertVoteThresholdAddition) {
			if st.LostEpoch < 0 {
				return xerrors.Errorf("failed to vaildate expert with below vote:%w", st.VoteAmount)
			} else if (st.LostEpoch + ExpertVoteCheckPeriod) < rt.CurrEpoch() {
				return xerrors.Errorf("failed to vaildate expert with lost vote:%w", st.VoteAmount)
			}
		}
	default:
		return xerrors.Errorf("failed to validate expert status: %d", st.Status)
	}

	return nil
}