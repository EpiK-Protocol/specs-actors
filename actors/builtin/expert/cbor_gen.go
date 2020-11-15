// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package expert

import (
	"fmt"
	"io"

	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf

var lengthBufState = []byte{130}

func (t *State) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Info (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Info); err != nil {
		return xerrors.Errorf("failed to write cid field t.Info: %w", err)
	}

	// t.Datas (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Datas); err != nil {
		return xerrors.Errorf("failed to write cid field t.Datas: %w", err)
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

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Info (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Info: %w", err)
		}

		t.Info = c

	}
	// t.Datas (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Datas: %w", err)
		}

		t.Datas = c

	}
	return nil
}

var lengthBufExpertInfo = []byte{131}

func (t *ExpertInfo) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufExpertInfo); err != nil {
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

func (t *ExpertInfo) UnmarshalCBOR(r io.Reader) error {
	*t = ExpertInfo{}

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

var lengthBufGetControlAddressReturn = []byte{129}

func (t *GetControlAddressReturn) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufGetControlAddressReturn); err != nil {
		return err
	}

	// t.Owner (address.Address) (struct)
	if err := t.Owner.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *GetControlAddressReturn) UnmarshalCBOR(r io.Reader) error {
	*t = GetControlAddressReturn{}

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

var lengthBufChangePeerIDParams = []byte{129}

func (t *ChangePeerIDParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufChangePeerIDParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.NewID ([]uint8) (slice)
	if len(t.NewID) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.NewID was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.NewID))); err != nil {
		return err
	}

	if _, err := w.Write(t.NewID[:]); err != nil {
		return err
	}
	return nil
}

func (t *ChangePeerIDParams) UnmarshalCBOR(r io.Reader) error {
	*t = ChangePeerIDParams{}

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

	// t.NewID ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.NewID: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.NewID = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.NewID[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufChangeMultiaddrsParams = []byte{129}

func (t *ChangeMultiaddrsParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufChangeMultiaddrsParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.NewMultiaddrs ([][]uint8) (slice)
	if len(t.NewMultiaddrs) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.NewMultiaddrs was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.NewMultiaddrs))); err != nil {
		return err
	}
	for _, v := range t.NewMultiaddrs {
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

func (t *ChangeMultiaddrsParams) UnmarshalCBOR(r io.Reader) error {
	*t = ChangeMultiaddrsParams{}

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

	// t.NewMultiaddrs ([][]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.NewMultiaddrs: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}

	if extra > 0 {
		t.NewMultiaddrs = make([][]uint8, extra)
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
				return fmt.Errorf("t.NewMultiaddrs[i]: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.NewMultiaddrs[i] = make([]uint8, extra)
			}

			if _, err := io.ReadFull(br, t.NewMultiaddrs[i][:]); err != nil {
				return err
			}
		}
	}

	return nil
}

var lengthBufChangeAddressParams = []byte{129}

func (t *ChangeAddressParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufChangeAddressParams); err != nil {
		return err
	}

	// t.NewOwner (address.Address) (struct)
	if err := t.NewOwner.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *ChangeAddressParams) UnmarshalCBOR(r io.Reader) error {
	*t = ChangeAddressParams{}

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

	// t.NewOwner (address.Address) (struct)

	{

		if err := t.NewOwner.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.NewOwner: %w", err)
		}

	}
	return nil
}

var lengthBufExpertDataParams = []byte{130}

func (t *ExpertDataParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufExpertDataParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.PieceID (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.PieceID); err != nil {
		return xerrors.Errorf("failed to write cid field t.PieceID: %w", err)
	}

	// t.Bounty (string) (string)
	if len(t.Bounty) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.Bounty was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.Bounty))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.Bounty)); err != nil {
		return err
	}
	return nil
}

func (t *ExpertDataParams) UnmarshalCBOR(r io.Reader) error {
	*t = ExpertDataParams{}

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

	// t.PieceID (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.PieceID: %w", err)
		}

		t.PieceID = c

	}
	// t.Bounty (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.Bounty = string(sval)
	}
	return nil
}

var lengthBufDataOnChainInfo = []byte{130}

func (t *DataOnChainInfo) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufDataOnChainInfo); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.PieceID (string) (string)
	if len(t.PieceID) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.PieceID was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.PieceID))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.PieceID)); err != nil {
		return err
	}

	// t.Bounty (string) (string)
	if len(t.Bounty) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.Bounty was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.Bounty))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.Bounty)); err != nil {
		return err
	}
	return nil
}

func (t *DataOnChainInfo) UnmarshalCBOR(r io.Reader) error {
	*t = DataOnChainInfo{}

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

	// t.PieceID (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.PieceID = string(sval)
	}
	// t.Bounty (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.Bounty = string(sval)
	}
	return nil
}
