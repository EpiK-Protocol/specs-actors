// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package power

import (
	"fmt"
	"io"

	address "github.com/filecoin-project/go-address"
	abi "github.com/filecoin-project/go-state-types/abi"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf

var lengthBufState = []byte{145}

func (t *State) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.TotalRawBytePower (big.Int) (struct)
	if err := t.TotalRawBytePower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.TotalBytesCommitted (big.Int) (struct)
	if err := t.TotalBytesCommitted.MarshalCBOR(w); err != nil {
		return err
	}

	// t.TotalQualityAdjPower (big.Int) (struct)
	if err := t.TotalQualityAdjPower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.TotalQABytesCommitted (big.Int) (struct)
	if err := t.TotalQABytesCommitted.MarshalCBOR(w); err != nil {
		return err
	}

	// t.TotalPledgeCollateral (big.Int) (struct)
	if err := t.TotalPledgeCollateral.MarshalCBOR(w); err != nil {
		return err
	}

	// t.ThisEpochRawBytePower (big.Int) (struct)
	if err := t.ThisEpochRawBytePower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.ThisEpochQualityAdjPower (big.Int) (struct)
	if err := t.ThisEpochQualityAdjPower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.ThisEpochPledgeCollateral (big.Int) (struct)
	if err := t.ThisEpochPledgeCollateral.MarshalCBOR(w); err != nil {
		return err
	}

	// t.ThisEpochQAPowerSmoothed (smoothing.FilterEstimate) (struct)
	if err := t.ThisEpochQAPowerSmoothed.MarshalCBOR(w); err != nil {
		return err
	}

	// t.MinerCount (int64) (int64)
	if t.MinerCount >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.MinerCount)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.MinerCount-1)); err != nil {
			return err
		}
	}

	// t.MinerAboveMinPowerCount (int64) (int64)
	if t.MinerAboveMinPowerCount >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.MinerAboveMinPowerCount)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.MinerAboveMinPowerCount-1)); err != nil {
			return err
		}
	}

	// t.CronEventQueue (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.CronEventQueue); err != nil {
		return xerrors.Errorf("failed to write cid field t.CronEventQueue: %w", err)
	}

	// t.FirstCronEpoch (abi.ChainEpoch) (int64)
	if t.FirstCronEpoch >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.FirstCronEpoch)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.FirstCronEpoch-1)); err != nil {
			return err
		}
	}

	// t.Claims (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Claims); err != nil {
		return xerrors.Errorf("failed to write cid field t.Claims: %w", err)
	}

	// t.ProofValidationBatch (cid.Cid) (struct)

	if t.ProofValidationBatch == nil {
		if _, err := w.Write(cbg.CborNull); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteCidBuf(scratch, w, *t.ProofValidationBatch); err != nil {
			return xerrors.Errorf("failed to write cid field t.ProofValidationBatch: %w", err)
		}
	}

	// t.ExpertCount (int64) (int64)
	if t.ExpertCount >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.ExpertCount)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.ExpertCount-1)); err != nil {
			return err
		}
	}

	// t.Experts (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Experts); err != nil {
		return xerrors.Errorf("failed to write cid field t.Experts: %w", err)
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

	if extra != 17 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.TotalRawBytePower (big.Int) (struct)

	{

		if err := t.TotalRawBytePower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalRawBytePower: %w", err)
		}

	}
	// t.TotalBytesCommitted (big.Int) (struct)

	{

		if err := t.TotalBytesCommitted.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalBytesCommitted: %w", err)
		}

	}
	// t.TotalQualityAdjPower (big.Int) (struct)

	{

		if err := t.TotalQualityAdjPower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalQualityAdjPower: %w", err)
		}

	}
	// t.TotalQABytesCommitted (big.Int) (struct)

	{

		if err := t.TotalQABytesCommitted.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalQABytesCommitted: %w", err)
		}

	}
	// t.TotalPledgeCollateral (big.Int) (struct)

	{

		if err := t.TotalPledgeCollateral.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.TotalPledgeCollateral: %w", err)
		}

	}
	// t.ThisEpochRawBytePower (big.Int) (struct)

	{

		if err := t.ThisEpochRawBytePower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.ThisEpochRawBytePower: %w", err)
		}

	}
	// t.ThisEpochQualityAdjPower (big.Int) (struct)

	{

		if err := t.ThisEpochQualityAdjPower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.ThisEpochQualityAdjPower: %w", err)
		}

	}
	// t.ThisEpochPledgeCollateral (big.Int) (struct)

	{

		if err := t.ThisEpochPledgeCollateral.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.ThisEpochPledgeCollateral: %w", err)
		}

	}
	// t.ThisEpochQAPowerSmoothed (smoothing.FilterEstimate) (struct)

	{

		if err := t.ThisEpochQAPowerSmoothed.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.ThisEpochQAPowerSmoothed: %w", err)
		}

	}
	// t.MinerCount (int64) (int64)
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

		t.MinerCount = int64(extraI)
	}
	// t.MinerAboveMinPowerCount (int64) (int64)
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

		t.MinerAboveMinPowerCount = int64(extraI)
	}
	// t.CronEventQueue (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.CronEventQueue: %w", err)
		}

		t.CronEventQueue = c

	}
	// t.FirstCronEpoch (abi.ChainEpoch) (int64)
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

		t.FirstCronEpoch = abi.ChainEpoch(extraI)
	}
	// t.Claims (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Claims: %w", err)
		}

		t.Claims = c

	}
	// t.ProofValidationBatch (cid.Cid) (struct)

	{

		b, err := br.ReadByte()
		if err != nil {
			return err
		}
		if b != cbg.CborNull[0] {
			if err := br.UnreadByte(); err != nil {
				return err
			}

			c, err := cbg.ReadCid(br)
			if err != nil {
				return xerrors.Errorf("failed to read cid field t.ProofValidationBatch: %w", err)
			}

			t.ProofValidationBatch = &c
		}

	}
	// t.ExpertCount (int64) (int64)
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

		t.ExpertCount = int64(extraI)
	}
	// t.Experts (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Experts: %w", err)
		}

		t.Experts = c

	}
	return nil
}

var lengthBufClaim = []byte{131}

func (t *Claim) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufClaim); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.SealProofType (abi.RegisteredSealProof) (int64)
	if t.SealProofType >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.SealProofType)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.SealProofType-1)); err != nil {
			return err
		}
	}

	// t.RawBytePower (big.Int) (struct)
	if err := t.RawBytePower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.QualityAdjPower (big.Int) (struct)
	if err := t.QualityAdjPower.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *Claim) UnmarshalCBOR(r io.Reader) error {
	*t = Claim{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 3 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.SealProofType (abi.RegisteredSealProof) (int64)
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

		t.SealProofType = abi.RegisteredSealProof(extraI)
	}
	// t.RawBytePower (big.Int) (struct)

	{

		if err := t.RawBytePower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.RawBytePower: %w", err)
		}

	}
	// t.QualityAdjPower (big.Int) (struct)

	{

		if err := t.QualityAdjPower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.QualityAdjPower: %w", err)
		}

	}
	return nil
}

var lengthBufCronEvent = []byte{130}

func (t *CronEvent) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCronEvent); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.MinerAddr (address.Address) (struct)
	if err := t.MinerAddr.MarshalCBOR(w); err != nil {
		return err
	}

	// t.CallbackPayload ([]uint8) (slice)
	if len(t.CallbackPayload) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.CallbackPayload was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.CallbackPayload))); err != nil {
		return err
	}

	if _, err := w.Write(t.CallbackPayload[:]); err != nil {
		return err
	}
	return nil
}

func (t *CronEvent) UnmarshalCBOR(r io.Reader) error {
	*t = CronEvent{}

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

	// t.MinerAddr (address.Address) (struct)

	{

		if err := t.MinerAddr.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.MinerAddr: %w", err)
		}

	}
	// t.CallbackPayload ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.CallbackPayload: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.CallbackPayload = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.CallbackPayload[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufCreateMinerParams = []byte{133}

func (t *CreateMinerParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCreateMinerParams); err != nil {
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

	// t.SealProofType (abi.RegisteredSealProof) (int64)
	if t.SealProofType >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.SealProofType)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.SealProofType-1)); err != nil {
			return err
		}
	}

	// t.Peer ([]uint8) (slice)
	if len(t.Peer) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Peer was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Peer))); err != nil {
		return err
	}

	if _, err := w.Write(t.Peer[:]); err != nil {
		return err
	}

	// t.Multiaddrs ([][]uint8) (slice)
	if len(t.Multiaddrs) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Multiaddrs was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Multiaddrs))); err != nil {
		return err
	}
	for _, v := range t.Multiaddrs {
		if len(v) > cbg.ByteArrayMaxLen {
			return xerrors.Errorf("Byte array in field v was too long")
		}

		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(v))); err != nil {
			return err
		}

		if _, err := w.Write(v[:]); err != nil {
			return err
		}
	}
	return nil
}

func (t *CreateMinerParams) UnmarshalCBOR(r io.Reader) error {
	*t = CreateMinerParams{}

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
	// t.SealProofType (abi.RegisteredSealProof) (int64)
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

		t.SealProofType = abi.RegisteredSealProof(extraI)
	}
	// t.Peer ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.Peer: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Peer = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.Peer[:]); err != nil {
		return err
	}
	// t.Multiaddrs ([][]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Multiaddrs: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Multiaddrs = make([][]uint8, extra)
	}

	for i := 0; i < int(extra); i++ {
		{
			var maj byte
			var extra uint64
			var err error

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Multiaddrs[i]: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Multiaddrs[i] = make([]uint8, extra)
			}

			if _, err := io.ReadFull(br, t.Multiaddrs[i][:]); err != nil {
				return err
			}
		}
	}

	return nil
}

var lengthBufCreateMinerReturn = []byte{130}

func (t *CreateMinerReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCreateMinerReturn); err != nil {
		return err
	}

	// t.IDAddress (address.Address) (struct)
	if err := t.IDAddress.MarshalCBOR(w); err != nil {
		return err
	}

	// t.RobustAddress (address.Address) (struct)
	if err := t.RobustAddress.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *CreateMinerReturn) UnmarshalCBOR(r io.Reader) error {
	*t = CreateMinerReturn{}

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

	// t.IDAddress (address.Address) (struct)

	{

		if err := t.IDAddress.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.IDAddress: %w", err)
		}

	}
	// t.RobustAddress (address.Address) (struct)

	{

		if err := t.RobustAddress.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.RobustAddress: %w", err)
		}

	}
	return nil
}

var lengthBufEnrollCronEventParams = []byte{130}

func (t *EnrollCronEventParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufEnrollCronEventParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.EventEpoch (abi.ChainEpoch) (int64)
	if t.EventEpoch >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.EventEpoch)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.EventEpoch-1)); err != nil {
			return err
		}
	}

	// t.Payload ([]uint8) (slice)
	if len(t.Payload) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Payload was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Payload))); err != nil {
		return err
	}

	if _, err := w.Write(t.Payload[:]); err != nil {
		return err
	}
	return nil
}

func (t *EnrollCronEventParams) UnmarshalCBOR(r io.Reader) error {
	*t = EnrollCronEventParams{}

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

	// t.EventEpoch (abi.ChainEpoch) (int64)
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

		t.EventEpoch = abi.ChainEpoch(extraI)
	}
	// t.Payload ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.Payload: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Payload = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.Payload[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufUpdateClaimedPowerParams = []byte{130}

func (t *UpdateClaimedPowerParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufUpdateClaimedPowerParams); err != nil {
		return err
	}

	// t.RawByteDelta (big.Int) (struct)
	if err := t.RawByteDelta.MarshalCBOR(w); err != nil {
		return err
	}

	// t.QualityAdjustedDelta (big.Int) (struct)
	if err := t.QualityAdjustedDelta.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *UpdateClaimedPowerParams) UnmarshalCBOR(r io.Reader) error {
	*t = UpdateClaimedPowerParams{}

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

	// t.RawByteDelta (big.Int) (struct)

	{

		if err := t.RawByteDelta.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.RawByteDelta: %w", err)
		}

	}
	// t.QualityAdjustedDelta (big.Int) (struct)

	{

		if err := t.QualityAdjustedDelta.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.QualityAdjustedDelta: %w", err)
		}

	}
	return nil
}

var lengthBufCurrentTotalPowerReturn = []byte{132}

func (t *CurrentTotalPowerReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCurrentTotalPowerReturn); err != nil {
		return err
	}

	// t.RawBytePower (big.Int) (struct)
	if err := t.RawBytePower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.QualityAdjPower (big.Int) (struct)
	if err := t.QualityAdjPower.MarshalCBOR(w); err != nil {
		return err
	}

	// t.PledgeCollateral (big.Int) (struct)
	if err := t.PledgeCollateral.MarshalCBOR(w); err != nil {
		return err
	}

	// t.QualityAdjPowerSmoothed (smoothing.FilterEstimate) (struct)
	if err := t.QualityAdjPowerSmoothed.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *CurrentTotalPowerReturn) UnmarshalCBOR(r io.Reader) error {
	*t = CurrentTotalPowerReturn{}

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

	// t.RawBytePower (big.Int) (struct)

	{

		if err := t.RawBytePower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.RawBytePower: %w", err)
		}

	}
	// t.QualityAdjPower (big.Int) (struct)

	{

		if err := t.QualityAdjPower.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.QualityAdjPower: %w", err)
		}

	}
	// t.PledgeCollateral (big.Int) (struct)

	{

		if err := t.PledgeCollateral.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.PledgeCollateral: %w", err)
		}

	}
	// t.QualityAdjPowerSmoothed (smoothing.FilterEstimate) (struct)

	{

		if err := t.QualityAdjPowerSmoothed.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.QualityAdjPowerSmoothed: %w", err)
		}

	}
	return nil
}

var lengthBufMinerConstructorParams = []byte{134}

func (t *MinerConstructorParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufMinerConstructorParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.OwnerAddr (address.Address) (struct)
	if err := t.OwnerAddr.MarshalCBOR(w); err != nil {
		return err
	}

	// t.WorkerAddr (address.Address) (struct)
	if err := t.WorkerAddr.MarshalCBOR(w); err != nil {
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

	// t.SealProofType (abi.RegisteredSealProof) (int64)
	if t.SealProofType >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.SealProofType)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.SealProofType-1)); err != nil {
			return err
		}
	}

	// t.PeerId ([]uint8) (slice)
	if len(t.PeerId) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.PeerId was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.PeerId))); err != nil {
		return err
	}

	if _, err := w.Write(t.PeerId[:]); err != nil {
		return err
	}

	// t.Multiaddrs ([][]uint8) (slice)
	if len(t.Multiaddrs) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Multiaddrs was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Multiaddrs))); err != nil {
		return err
	}
	for _, v := range t.Multiaddrs {
		if len(v) > cbg.ByteArrayMaxLen {
			return xerrors.Errorf("Byte array in field v was too long")
		}

		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(v))); err != nil {
			return err
		}

		if _, err := w.Write(v[:]); err != nil {
			return err
		}
	}
	return nil
}

func (t *MinerConstructorParams) UnmarshalCBOR(r io.Reader) error {
	*t = MinerConstructorParams{}

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

	// t.OwnerAddr (address.Address) (struct)

	{

		if err := t.OwnerAddr.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.OwnerAddr: %w", err)
		}

	}
	// t.WorkerAddr (address.Address) (struct)

	{

		if err := t.WorkerAddr.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.WorkerAddr: %w", err)
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

	// t.SealProofType (abi.RegisteredSealProof) (int64)
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

		t.SealProofType = abi.RegisteredSealProof(extraI)
	}
	// t.PeerId ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.PeerId: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.PeerId = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.PeerId[:]); err != nil {
		return err
	}
	// t.Multiaddrs ([][]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Multiaddrs: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Multiaddrs = make([][]uint8, extra)
	}

	for i := 0; i < int(extra); i++ {
		{
			var maj byte
			var extra uint64
			var err error

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Multiaddrs[i]: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Multiaddrs[i] = make([]uint8, extra)
			}

			if _, err := io.ReadFull(br, t.Multiaddrs[i][:]); err != nil {
				return err
			}
		}
	}

	return nil
}

var lengthBufExpert = []byte{129}

func (t *Expert) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufExpert); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.DataCount (int64) (int64)
	if t.DataCount >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.DataCount)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.DataCount-1)); err != nil {
			return err
		}
	}
	return nil
}

func (t *Expert) UnmarshalCBOR(r io.Reader) error {
	*t = Expert{}

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

	// t.DataCount (int64) (int64)
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

		t.DataCount = int64(extraI)
	}
	return nil
}

var lengthBufCreateExpertParams = []byte{131}

func (t *CreateExpertParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCreateExpertParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Owner (address.Address) (struct)
	if err := t.Owner.MarshalCBOR(w); err != nil {
		return err
	}

	// t.PeerId ([]uint8) (slice)
	if len(t.PeerId) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.PeerId was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.PeerId))); err != nil {
		return err
	}

	if _, err := w.Write(t.PeerId[:]); err != nil {
		return err
	}

	// t.Multiaddrs ([][]uint8) (slice)
	if len(t.Multiaddrs) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Multiaddrs was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Multiaddrs))); err != nil {
		return err
	}
	for _, v := range t.Multiaddrs {
		if len(v) > cbg.ByteArrayMaxLen {
			return xerrors.Errorf("Byte array in field v was too long")
		}

		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(v))); err != nil {
			return err
		}

		if _, err := w.Write(v[:]); err != nil {
			return err
		}
	}
	return nil
}

func (t *CreateExpertParams) UnmarshalCBOR(r io.Reader) error {
	*t = CreateExpertParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 3 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Owner (address.Address) (struct)

	{

		if err := t.Owner.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Owner: %w", err)
		}

	}
	// t.PeerId ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.PeerId: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.PeerId = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.PeerId[:]); err != nil {
		return err
	}
	// t.Multiaddrs ([][]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Multiaddrs: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.Multiaddrs = make([][]uint8, extra)
	}

	for i := 0; i < int(extra); i++ {
		{
			var maj byte
			var extra uint64
			var err error

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Multiaddrs[i]: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Multiaddrs[i] = make([]uint8, extra)
			}

			if _, err := io.ReadFull(br, t.Multiaddrs[i][:]); err != nil {
				return err
			}
		}
	}

	return nil
}

var lengthBufCreateExpertReturn = []byte{130}

func (t *CreateExpertReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCreateExpertReturn); err != nil {
		return err
	}

	// t.IDAddress (address.Address) (struct)
	if err := t.IDAddress.MarshalCBOR(w); err != nil {
		return err
	}

	// t.RobustAddress (address.Address) (struct)
	if err := t.RobustAddress.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *CreateExpertReturn) UnmarshalCBOR(r io.Reader) error {
	*t = CreateExpertReturn{}

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

	// t.IDAddress (address.Address) (struct)

	{

		if err := t.IDAddress.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.IDAddress: %w", err)
		}

	}
	// t.RobustAddress (address.Address) (struct)

	{

		if err := t.RobustAddress.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.RobustAddress: %w", err)
		}

	}
	return nil
}

var lengthBufDeleteExpertParams = []byte{129}

func (t *DeleteExpertParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufDeleteExpertParams); err != nil {
		return err
	}

	// t.Expert (address.Address) (struct)
	if err := t.Expert.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *DeleteExpertParams) UnmarshalCBOR(r io.Reader) error {
	*t = DeleteExpertParams{}

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

	// t.Expert (address.Address) (struct)

	{

		if err := t.Expert.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Expert: %w", err)
		}

	}
	return nil
}
