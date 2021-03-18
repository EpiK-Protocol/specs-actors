// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package builtin

import (
	"fmt"
	"io"

	address "github.com/filecoin-project/go-address"
	abi "github.com/filecoin-project/go-state-types/abi"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf

var lengthBufGetControlAddressesReturn = []byte{132}

func (t *GetControlAddressesReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufGetControlAddressesReturn); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Owner (address.Address) (struct)
	if err := t.Owner.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Worker (address.Address) (struct)
	if err := t.Worker.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Coinbase (address.Address) (struct)
	if err := t.Coinbase.MarshalCBOR(w); err != nil {
		return err
	}

	// t.ControlAddrs ([]address.Address) (slice)
	if len(t.ControlAddrs) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.ControlAddrs was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.ControlAddrs))); err != nil {
		return err
	}
	for _, v := range t.ControlAddrs {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *GetControlAddressesReturn) UnmarshalCBOR(r io.Reader) error {
	*t = GetControlAddressesReturn{}

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

	// t.Owner (address.Address) (struct)

	{

		if err := t.Owner.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Owner: %w", err)
		}

	}
	// t.Worker (address.Address) (struct)

	{

		if err := t.Worker.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Worker: %w", err)
		}

	}
	// t.Coinbase (address.Address) (struct)

	{

		if err := t.Coinbase.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Coinbase: %w", err)
		}

	}
	// t.ControlAddrs ([]address.Address) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.ControlAddrs: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.ControlAddrs = make([]address.Address, extra)
	}

	for i := 0; i < int(extra); i++ {

		var v address.Address
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.ControlAddrs[i] = v
	}

	return nil
}

var lengthBufExpertControlAddressReturn = []byte{129}

func (t *ExpertControlAddressReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufExpertControlAddressReturn); err != nil {
		return err
	}

	// t.Owner (address.Address) (struct)
	if err := t.Owner.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *ExpertControlAddressReturn) UnmarshalCBOR(r io.Reader) error {
	*t = ExpertControlAddressReturn{}

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

	// t.Owner (address.Address) (struct)

	{

		if err := t.Owner.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Owner: %w", err)
		}

	}
	return nil
}

var lengthBufConfirmSectorProofsParams = []byte{129}

func (t *ConfirmSectorProofsParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufConfirmSectorProofsParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Sectors ([]abi.SectorNumber) (slice)
	if len(t.Sectors) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Sectors was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Sectors))); err != nil {
		return err
	}
	for _, v := range t.Sectors {
		if err := cbg.CborWriteHeader(w, cbg.MajUnsignedInt, uint64(v)); err != nil {
			return err
		}
	}
	return nil
}

func (t *ConfirmSectorProofsParams) UnmarshalCBOR(r io.Reader) error {
	*t = ConfirmSectorProofsParams{}

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

	// t.Sectors ([]abi.SectorNumber) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Sectors: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Sectors = make([]abi.SectorNumber, extra)
	}

	for i := 0; i < int(extra); i++ {

		maj, val, err := cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return xerrors.Errorf("failed to read uint64 for t.Sectors slice: %w", err)
		}

		if maj != cbg.MajUnsignedInt {
			return xerrors.Errorf("value read for array t.Sectors was not a uint, instead got %d", maj)
		}

		t.Sectors[i] = abi.SectorNumber(val)
	}

	return nil
}

var lengthBufApplyRewardParams = []byte{130}

func (t *ApplyRewardParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufApplyRewardParams); err != nil {
		return err
	}

	// t.Reward (big.Int) (struct)
	if err := t.Reward.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Penalty (big.Int) (struct)
	if err := t.Penalty.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *ApplyRewardParams) UnmarshalCBOR(r io.Reader) error {
	*t = ApplyRewardParams{}

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

	// t.Reward (big.Int) (struct)

	{

		if err := t.Reward.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Reward: %w", err)
		}

	}
	// t.Penalty (big.Int) (struct)

	{

		if err := t.Penalty.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Penalty: %w", err)
		}

	}
	return nil
}

var lengthBufOnExpertImportParams = []byte{129}

func (t *OnExpertImportParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufOnExpertImportParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.PieceID (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.PieceID); err != nil {
		return xerrors.Errorf("failed to write cid field t.PieceID: %w", err)
	}

	return nil
}

func (t *OnExpertImportParams) UnmarshalCBOR(r io.Reader) error {
	*t = OnExpertImportParams{}

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

	// t.PieceID (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.PieceID: %w", err)
		}

		t.PieceID = c

	}
	return nil
}

var lengthBufBoolValue = []byte{129}

func (t *BoolValue) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufBoolValue); err != nil {
		return err
	}

	// t.Bool (bool) (bool)
	if err := cbg.WriteBool(w, t.Bool); err != nil {
		return err
	}
	return nil
}

func (t *BoolValue) UnmarshalCBOR(r io.Reader) error {
	*t = BoolValue{}

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

	// t.Bool (bool) (bool)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajOther {
		return fmt.Errorf("booleans must be major type 7")
	}
	switch extra {
	case 20:
		t.Bool = false
	case 21:
		t.Bool = true
	default:
		return fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
	}
	return nil
}

var lengthBufValidateGrantedParams = []byte{130}

func (t *ValidateGrantedParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufValidateGrantedParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Caller (address.Address) (struct)
	if err := t.Caller.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Method (abi.MethodNum) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Method)); err != nil {
		return err
	}

	return nil
}

func (t *ValidateGrantedParams) UnmarshalCBOR(r io.Reader) error {
	*t = ValidateGrantedParams{}

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

	// t.Caller (address.Address) (struct)

	{

		if err := t.Caller.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Caller: %w", err)
		}

	}
	// t.Method (abi.MethodNum) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Method = abi.MethodNum(extra)

	}
	return nil
}

var lengthBufBatchPieceCIDParams = []byte{129}

func (t *BatchPieceCIDParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufBatchPieceCIDParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.PieceCIDs ([]builtin.CheckedCID) (slice)
	if len(t.PieceCIDs) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.PieceCIDs was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.PieceCIDs))); err != nil {
		return err
	}
	for _, v := range t.PieceCIDs {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *BatchPieceCIDParams) UnmarshalCBOR(r io.Reader) error {
	*t = BatchPieceCIDParams{}

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

	// t.PieceCIDs ([]builtin.CheckedCID) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.PieceCIDs: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.PieceCIDs = make([]CheckedCID, extra)
	}

	for i := 0; i < int(extra); i++ {

		var v CheckedCID
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.PieceCIDs[i] = v
	}

	return nil
}

var lengthBufCheckedCID = []byte{129}

func (t *CheckedCID) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCheckedCID); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.CID (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.CID); err != nil {
		return xerrors.Errorf("failed to write cid field t.CID: %w", err)
	}

	return nil
}

func (t *CheckedCID) UnmarshalCBOR(r io.Reader) error {
	*t = CheckedCID{}

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

	// t.CID (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.CID: %w", err)
		}

		t.CID = c

	}
	return nil
}
