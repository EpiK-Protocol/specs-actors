// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package retrieval

import (
	"fmt"
	"io"

	address "github.com/filecoin-project/go-address"
	abi "github.com/filecoin-project/go-state-types/abi"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf

var lengthBufState = []byte{135}

func (t *State) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.RetrievalStates (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.RetrievalStates); err != nil {
		return xerrors.Errorf("failed to write cid field t.RetrievalStates: %w", err)
	}

	// t.Pledges (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Pledges); err != nil {
		return xerrors.Errorf("failed to write cid field t.Pledges: %w", err)
	}

	// t.LockedTable (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.LockedTable); err != nil {
		return xerrors.Errorf("failed to write cid field t.LockedTable: %w", err)
	}

	// t.TotalLockedCollateral (big.Int) (struct)
	if err := t.TotalLockedCollateral.MarshalCBOR(w); err != nil {
		return err
	}

	// t.TotalCollateral (big.Int) (struct)
	if err := t.TotalCollateral.MarshalCBOR(w); err != nil {
		return err
	}

	// t.TotalRetrievalReward (big.Int) (struct)
	if err := t.TotalRetrievalReward.MarshalCBOR(w); err != nil {
		return err
	}

	// t.PendingReward (big.Int) (struct)
	if err := t.PendingReward.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *State) UnmarshalCBOR(r io.Reader) error {
	*t = State{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 7 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.RetrievalStates (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.RetrievalStates: %w", err)
		}

		t.RetrievalStates = c

	}
	// t.Pledges (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Pledges: %w", err)
		}

		t.Pledges = c

	}
	// t.LockedTable (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.LockedTable: %w", err)
		}

		t.LockedTable = c

	}
	// t.TotalLockedCollateral (big.Int) (struct)

	{

		if err := t.TotalLockedCollateral.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalLockedCollateral: %w", err)
		}

	}
	// t.TotalCollateral (big.Int) (struct)

	{

		if err := t.TotalCollateral.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalCollateral: %w", err)
		}

	}
	// t.TotalRetrievalReward (big.Int) (struct)

	{

		if err := t.TotalRetrievalReward.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalRetrievalReward: %w", err)
		}

	}
	// t.PendingReward (big.Int) (struct)

	{

		if err := t.PendingReward.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.PendingReward: %w", err)
		}

	}
	return nil
}

var lengthBufPledgeParams = []byte{130}

func (t *PledgeParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufPledgeParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Address (address.Address) (struct)
	if err := t.Address.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Miners ([]address.Address) (slice)
	if len(t.Miners) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Miners was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Miners))); err != nil {
		return err
	}
	for _, v := range t.Miners {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *PledgeParams) UnmarshalCBOR(r io.Reader) error {
	*t = PledgeParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Address (address.Address) (struct)

	{

		if err := t.Address.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Address: %w", err)
		}

	}
	// t.Miners ([]address.Address) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Miners: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Miners = make([]address.Address, extra)
	}

	for i := 0; i < int(extra); i++ {

		var v address.Address
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.Miners[i] = v
	}

	return nil
}

var lengthBufWithdrawBalanceParams = []byte{130}

func (t *WithdrawBalanceParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufWithdrawBalanceParams); err != nil {
		return err
	}

	// t.ProviderOrClientAddress (address.Address) (struct)
	if err := t.ProviderOrClientAddress.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Amount (big.Int) (struct)
	if err := t.Amount.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *WithdrawBalanceParams) UnmarshalCBOR(r io.Reader) error {
	*t = WithdrawBalanceParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.ProviderOrClientAddress (address.Address) (struct)

	{

		if err := t.ProviderOrClientAddress.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.ProviderOrClientAddress: %w", err)
		}

	}
	// t.Amount (big.Int) (struct)

	{

		if err := t.Amount.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Amount: %w", err)
		}

	}
	return nil
}

var lengthBufRetrievalDataParams = []byte{132}

func (t *RetrievalDataParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufRetrievalDataParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.PayloadId (string) (string)
	if len(t.PayloadId) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.PayloadId was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.PayloadId))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.PayloadId)); err != nil {
		return err
	}

	// t.Size (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Size)); err != nil {
		return err
	}

	// t.Client (address.Address) (struct)
	if err := t.Client.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Provider (address.Address) (struct)
	if err := t.Provider.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *RetrievalDataParams) UnmarshalCBOR(r io.Reader) error {
	*t = RetrievalDataParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 4 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.PayloadId (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.PayloadId = string(sval)
	}
	// t.Size (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Size = uint64(extra)

	}
	// t.Client (address.Address) (struct)

	{

		if err := t.Client.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Client: %w", err)
		}

	}
	// t.Provider (address.Address) (struct)

	{

		if err := t.Provider.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Provider: %w", err)
		}

	}
	return nil
}

var lengthBufRetrievalState = []byte{134}

func (t *RetrievalState) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufRetrievalState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Pledges (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Pledges); err != nil {
		return xerrors.Errorf("failed to write cid field t.Pledges: %w", err)
	}

	// t.Miners ([]address.Address) (slice)
	if len(t.Miners) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Miners was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Miners))); err != nil {
		return err
	}
	for _, v := range t.Miners {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}

	// t.Datas (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Datas); err != nil {
		return xerrors.Errorf("failed to write cid field t.Datas: %w", err)
	}

	// t.Amount (big.Int) (struct)
	if err := t.Amount.MarshalCBOR(w); err != nil {
		return err
	}

	// t.EpochDate (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.EpochDate)); err != nil {
		return err
	}

	// t.DateSize (abi.PaddedPieceSize) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.DateSize)); err != nil {
		return err
	}

	return nil
}

func (t *RetrievalState) UnmarshalCBOR(r io.Reader) error {
	*t = RetrievalState{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 6 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Pledges (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Pledges: %w", err)
		}

		t.Pledges = c

	}
	// t.Miners ([]address.Address) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Miners: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Miners = make([]address.Address, extra)
	}

	for i := 0; i < int(extra); i++ {

		var v address.Address
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.Miners[i] = v
	}

	// t.Datas (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Datas: %w", err)
		}

		t.Datas = c

	}
	// t.Amount (big.Int) (struct)

	{

		if err := t.Amount.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Amount: %w", err)
		}

	}
	// t.EpochDate (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.EpochDate = uint64(extra)

	}
	// t.DateSize (abi.PaddedPieceSize) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.DateSize = abi.PaddedPieceSize(extra)

	}
	return nil
}

var lengthBufRetrievalData = []byte{133}

func (t *RetrievalData) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufRetrievalData); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.PayloadId (string) (string)
	if len(t.PayloadId) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.PayloadId was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.PayloadId))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.PayloadId)); err != nil {
		return err
	}

	// t.PieceSize (abi.PaddedPieceSize) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.PieceSize)); err != nil {
		return err
	}

	// t.Client (address.Address) (struct)
	if err := t.Client.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Provider (address.Address) (struct)
	if err := t.Provider.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Epoch (abi.ChainEpoch) (int64)
	if t.Epoch >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Epoch)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.Epoch-1)); err != nil {
			return err
		}
	}
	return nil
}

func (t *RetrievalData) UnmarshalCBOR(r io.Reader) error {
	*t = RetrievalData{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 5 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.PayloadId (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.PayloadId = string(sval)
	}
	// t.PieceSize (abi.PaddedPieceSize) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.PieceSize = abi.PaddedPieceSize(extra)

	}
	// t.Client (address.Address) (struct)

	{

		if err := t.Client.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Client: %w", err)
		}

	}
	// t.Provider (address.Address) (struct)

	{

		if err := t.Provider.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Provider: %w", err)
		}

	}
	// t.Epoch (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			return err
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			return fmt.Errorf("wrong type for int64 field: %d", maj)
		}

		t.Epoch = abi.ChainEpoch(extraI)
	}
	return nil
}

var lengthBufPledgeState = []byte{129}

func (t *PledgeState) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufPledgeState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Targets ([]address.Address) (slice)
	if len(t.Targets) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Targets was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Targets))); err != nil {
		return err
	}
	for _, v := range t.Targets {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *PledgeState) UnmarshalCBOR(r io.Reader) error {
	*t = PledgeState{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 1 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Targets ([]address.Address) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Targets: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Targets = make([]address.Address, extra)
	}

	for i := 0; i < int(extra); i++ {

		var v address.Address
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.Targets[i] = v
	}

	return nil
}

var lengthBufLockedState = []byte{130}

func (t *LockedState) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufLockedState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Amount (big.Int) (struct)
	if err := t.Amount.MarshalCBOR(w); err != nil {
		return err
	}

	// t.ApplyEpoch (abi.ChainEpoch) (int64)
	if t.ApplyEpoch >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.ApplyEpoch)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.ApplyEpoch-1)); err != nil {
			return err
		}
	}
	return nil
}

func (t *LockedState) UnmarshalCBOR(r io.Reader) error {
	*t = LockedState{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Amount (big.Int) (struct)

	{

		if err := t.Amount.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Amount: %w", err)
		}

	}
	// t.ApplyEpoch (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			return err
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			return fmt.Errorf("wrong type for int64 field: %d", maj)
		}

		t.ApplyEpoch = abi.ChainEpoch(extraI)
	}
	return nil
}

var lengthBufTotalCollateralReturn = []byte{129}

func (t *TotalCollateralReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufTotalCollateralReturn); err != nil {
		return err
	}

	// t.TotalCollateral (big.Int) (struct)
	if err := t.TotalCollateral.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *TotalCollateralReturn) UnmarshalCBOR(r io.Reader) error {
	*t = TotalCollateralReturn{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 1 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.TotalCollateral (big.Int) (struct)

	{

		if err := t.TotalCollateral.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalCollateral: %w", err)
		}

	}
	return nil
}

var lengthBufBindMinersParams = []byte{130}

func (t *BindMinersParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufBindMinersParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Pledger (address.Address) (struct)
	if err := t.Pledger.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Miners ([]address.Address) (slice)
	if len(t.Miners) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Miners was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Miners))); err != nil {
		return err
	}
	for _, v := range t.Miners {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *BindMinersParams) UnmarshalCBOR(r io.Reader) error {
	*t = BindMinersParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Pledger (address.Address) (struct)

	{

		if err := t.Pledger.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Pledger: %w", err)
		}

	}
	// t.Miners ([]address.Address) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Miners: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Miners = make([]address.Address, extra)
	}

	for i := 0; i < int(extra); i++ {

		var v address.Address
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.Miners[i] = v
	}

	return nil
}
