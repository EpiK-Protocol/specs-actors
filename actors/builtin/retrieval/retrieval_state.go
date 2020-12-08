package retrieval

import (
	addr "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
)

// State of retrieval
type State struct {
	RetrievalBatch cid.Cid // Multimap, (HAMT[Address]AMT[RetrievalState])

	// Total amount held in escrow, indexed by actor address (including both locked and unlocked amounts).
	EscrowTable cid.Cid // BalanceTable

	// Amount locked, indexed by actor address.
	LockedTable cid.Cid // // BalanceTable

	// Total Client Collateral that is locked -> unlocked when deal is terminated
	TotalLockedCollateral abi.TokenAmount

	// Total escrow table.
	TotalCollateral abi.TokenAmount

	// TotalRetrievalReward retrieval reward
	TotalRetrievalReward abi.TokenAmount

	// PendingReward temp pending reward
	PendingReward abi.TokenAmount
}

// RetrievalState record retrieval data statistics
type RetrievalState struct {
	PieceID   string
	PieceSize abi.PaddedPieceSize
	Client    addr.Address
	Provider  addr.Address
	Epoch     abi.ChainEpoch
}

// LockedState record lock state
type LockedState struct {
	Amount     abi.TokenAmount
	ApplyEpoch abi.ChainEpoch
}

// ConstructState retrieval construct
func ConstructState(store adt.Store) (*State, error) {

	emptyMapCid, err := adt.MakeEmptyMap(store, builtin.DefaultHamtBitwidth).Root()
	if err != nil {
		return nil, xerrors.Errorf("failed to create empty map: %w", err)
	}
	emptyMMapCid, err := adt.MakeEmptyMultimap(store, builtin.DefaultHamtBitwidth).Root()
	if err != nil {
		return nil, xerrors.Errorf("failed to create empty multi map: %w", err)
	}

	return &State{
		RetrievalBatch: emptyMMapCid,
		EscrowTable:    emptyMapCid,
		LockedTable:    emptyMapCid,

		TotalLockedCollateral: abi.NewTokenAmount(0),
		TotalCollateral:       abi.NewTokenAmount(0),
		TotalRetrievalReward:  abi.NewTokenAmount(0),
		PendingReward:         abi.NewTokenAmount(0),
	}, nil
}

// AddBalance add balance for
func (st *State) AddBalance(rt Runtime, fromAddr addr.Address, amount abi.TokenAmount) error {
	if amount.LessThan(big.Zero()) {
		return xerrors.Errorf("negative amount %v of funds to add", amount)
	}

	escrow, err := adt.AsBalanceTable(adt.AsStore(rt), st.EscrowTable)
	if err != nil {
		return err
	}
	if err := escrow.Add(fromAddr, amount); err != nil {
		return err
	}
	if st.EscrowTable, err = escrow.Root(); err != nil {
		return err
	}
	st.TotalCollateral = big.Add(st.TotalCollateral, amount)
	return nil
}

// ApplyForWithdraw apply for withdraw amount
func (st *State) ApplyForWithdraw(rt Runtime, fromAddr addr.Address, amount abi.TokenAmount) (exitcode.ExitCode, error) {
	if amount.LessThan(big.Zero()) {
		return exitcode.ErrIllegalState, xerrors.Errorf("negative amount %v of funds to apply", amount)
	}

	escrowTable, err := adt.AsBalanceTable(adt.AsStore(rt), st.EscrowTable)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}

	lockedMap, err := adt.AsMap(adt.AsStore(rt), st.LockedTable, builtin.DefaultHamtBitwidth)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}
	var out LockedState
	found, err := lockedMap.Get(abi.AddrKey(fromAddr), &out)

	if err != nil {
		return exitcode.ErrIllegalState, xerrors.Errorf("failed to get locked: %w", err)
	}

	if !found {
		out = LockedState{
			Amount: abi.NewTokenAmount(0),
		}
	}

	escrowBalance, err := escrowTable.Get(fromAddr)
	if err != nil {
		return exitcode.ErrIllegalState, xerrors.Errorf("failed to get escrow balance: %w", err)
	}

	if big.Add(out.Amount, amount).GreaterThan(escrowBalance) {
		return exitcode.ErrInsufficientFunds, xerrors.Errorf("not enough balance to lock for addr %s: escrow balance %s < locked %s + required %s", fromAddr, escrowBalance, out.Amount, amount)
	}

	out.Amount = big.Add(out.Amount, amount)
	out.ApplyEpoch = rt.CurrEpoch()
	if err := lockedMap.Put(abi.AddrKey(fromAddr), &out); err != nil {
		return exitcode.ErrIllegalState, err
	}
	if st.LockedTable, err = lockedMap.Root(); err != nil {
		return exitcode.ErrIllegalState, err
	}
	st.TotalLockedCollateral = big.Add(st.TotalLockedCollateral, out.Amount)
	return exitcode.Ok, nil
}

// Withdraw withdraw amount
func (st *State) Withdraw(rt Runtime, fromAddr addr.Address, amount abi.TokenAmount) (exitcode.ExitCode, error) {
	if amount.LessThan(big.Zero()) {
		return exitcode.ErrIllegalState, xerrors.Errorf("negative amount %v of funds to withdraw", amount)
	}

	lockedMap, err := adt.AsMap(adt.AsStore(rt), st.LockedTable, builtin.DefaultHamtBitwidth)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}
	var out LockedState
	found, err := lockedMap.Get(abi.AddrKey(fromAddr), &out)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}
	if !found {
		return exitcode.ErrIllegalState, xerrors.Errorf("withdraw not applied")
	}
	if rt.CurrEpoch()-out.ApplyEpoch < RetrievalLockPeriod || big.Sub(out.Amount, amount).LessThan(big.Zero()) {
		return exitcode.ErrForbidden, xerrors.Errorf("failed to withdraw at %d: %s", out.ApplyEpoch, amount)
	}
	out.Amount = big.Sub(out.Amount, amount)
	lockedMap.Put(abi.AddrKey(fromAddr), &out)
	if st.LockedTable, err = lockedMap.Root(); err != nil {
		return exitcode.ErrIllegalState, err
	}
	st.TotalLockedCollateral = big.Sub(st.TotalLockedCollateral, out.Amount)

	escrowTable, err := adt.AsBalanceTable(adt.AsStore(rt), st.EscrowTable)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}
	if err = escrowTable.MustSubtract(fromAddr, amount); err != nil {
		return exitcode.ErrForbidden, err
	}
	if st.EscrowTable, err = escrowTable.Root(); err != nil {
		return exitcode.ErrIllegalState, err
	}
	st.TotalCollateral = big.Sub(st.TotalCollateral, out.Amount)

	return exitcode.Ok, nil
}

// RetrievalData record the retrieval data
func (st *State) RetrievalData(rt Runtime, fromAddr addr.Address, state *RetrievalState) (exitcode.ExitCode, error) {
	mmap, err := adt.AsMultimap(adt.AsStore(rt), st.RetrievalBatch, builtin.DefaultHamtBitwidth)
	if err != nil {
		return exitcode.ErrIllegalState, xerrors.Errorf("failed to load retrieval batch set: %w", err)
	}

	var out RetrievalState
	curEpochDay := rt.CurrEpoch() / RetrievalStateDuration

	array, found, err := mmap.Get(abi.AddrKey(fromAddr))
	if err != nil {
		return exitcode.ErrIllegalState, err
	}
	if found && array.Length() > 0 {
		_, err = array.Get(array.Length()-1, &out)
		if err != nil {
			return exitcode.ErrIllegalState, err
		}
		lastEpochDay := out.Epoch / RetrievalStateDuration
		if lastEpochDay < curEpochDay {
			mmap.RemoveAll(abi.AddrKey(fromAddr))
		}
	}

	var totalSize abi.PaddedPieceSize
	err = mmap.ForEach(abi.AddrKey(fromAddr), &out, func(i int64) error {
		if out.Epoch/RetrievalStateDuration >= curEpochDay {
			totalSize += out.PieceSize
		}
		return nil
	})
	if err != nil {
		return exitcode.ErrIllegalState, err
	}

	escrow, err := adt.AsBalanceTable(adt.AsStore(rt), st.EscrowTable)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}
	balance, err := escrow.Get(fromAddr)
	if err != nil {
		return exitcode.ErrIllegalState, err
	}

	required := big.Mul(big.NewInt(int64(totalSize+state.PieceSize)), builtin.TokenPrecision)
	required = big.Div(required, big.NewInt(RetrievalSizePerEPK))
	if big.Sub(balance, required).LessThan(big.Zero()) {
		return exitcode.ErrInsufficientFunds, xerrors.Errorf("not enough balance to statistics for addr %s: escrow balance %s < required %s", fromAddr, balance, required)
	}
	mmap.Add(abi.AddrKey(fromAddr), state)
	if st.RetrievalBatch, err = mmap.Root(); err != nil {
		return exitcode.ErrIllegalState, err
	}

	return exitcode.Ok, nil
}

// ConfirmData record the retrieval data
func (st *State) ConfirmData(store adt.Store, currEpoch abi.ChainEpoch, fromAddr addr.Address, pieceID string) (abi.TokenAmount, error) {
	mmap, err := adt.AsMultimap(store, st.RetrievalBatch, builtin.DefaultHamtBitwidth)
	if err != nil {
		return abi.NewTokenAmount(0), xerrors.Errorf("failed to load retrieval batch set: %w", err)
	}

	index := int64(-1)
	var outs []RetrievalState
	var out RetrievalState
	err = mmap.ForEach(abi.AddrKey(fromAddr), &out, func(i int64) error {
		if out.PieceID == pieceID {
			index = i
		}
		outs = append(outs, out)
		return nil
	})
	if err != nil {
		return abi.NewTokenAmount(0), err
	}
	if index < 0 {
		return abi.NewTokenAmount(0), xerrors.Errorf("confirm data not found for addr %s", fromAddr)
	}

	amount := big.Mul(big.NewInt(int64(outs[index].PieceSize)), RetrievalRewardPerByte)
	if st.PendingReward.GreaterThanEqual(amount) {
		st.PendingReward = big.Sub(st.PendingReward, amount)
	} else {
		amount = st.PendingReward
		st.PendingReward = abi.NewTokenAmount(0)
	}
	return amount, nil
}

// EscrowBalance balance for address
func (st *State) EscrowBalance(store adt.Store, fromAddr addr.Address) (abi.TokenAmount, error) {
	escrowTable, err := adt.AsBalanceTable(store, st.EscrowTable)
	if err != nil {
		return abi.NewTokenAmount(0), err
	}

	escrowBalance, err := escrowTable.Get(fromAddr)
	if err != nil {
		return abi.NewTokenAmount(0), xerrors.Errorf("failed to get escrow balance: %w", err)
	}
	return escrowBalance, nil
}

// DayExpend balance for address
func (st *State) DayExpend(store adt.Store, epoch abi.ChainEpoch, fromAddr addr.Address) (abi.TokenAmount, error) {
	mmap, err := adt.AsMultimap(store, st.RetrievalBatch, builtin.DefaultHamtBitwidth)
	if err != nil {
		return abi.NewTokenAmount(0), xerrors.Errorf("failed to load retrieval batch set: %w", err)
	}

	curEpochDay := epoch / RetrievalStateDuration

	var totalSize abi.PaddedPieceSize
	var out RetrievalState
	err = mmap.ForEach(abi.AddrKey(fromAddr), &out, func(i int64) error {
		if out.Epoch/RetrievalStateDuration >= curEpochDay {
			totalSize += out.PieceSize
		}
		return nil
	})
	if err != nil {
		return abi.NewTokenAmount(0), err
	}
	expend := big.Mul(big.NewInt(int64(totalSize)), builtin.TokenPrecision)
	expend = big.Div(expend, big.NewInt(RetrievalSizePerEPK))
	return expend, nil
}

// LockedState locked state for address
func (st *State) LockedState(store adt.Store, fromAddr addr.Address) (*LockedState, error) {
	lockedMap, err := adt.AsMap(store, st.LockedTable, builtin.DefaultHamtBitwidth)
	if err != nil {
		return nil, err
	}
	var out LockedState
	_, err = lockedMap.Get(abi.AddrKey(fromAddr), &out)
	return &out, nil
}