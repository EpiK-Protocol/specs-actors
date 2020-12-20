package govern

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/filecoin-project/specs-actors/v2/actors/runtime"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
	"github.com/ipfs/go-cid"
)

type Actor struct{}

func (a Actor) Exports() []interface{} {
	return []interface{}{
		builtin.MethodConstructor: a.Constructor,
		2:                         a.Grant,
		3:                         a.Revoke,
		4:                         a.ValidateGranted,
	}
}

func (a Actor) Code() cid.Cid {
	return builtin.GovernActorCodeID
}

func (a Actor) IsSingleton() bool {
	return true
}

func (a Actor) State() cbor.Er { /* return new(State) */ return nil }

var _ runtime.VMActor = Actor{}

func (a Actor) Constructor(rt runtime.Runtime, supervisor address.Address) *abi.EmptyValue {
	emptyMap, err := adt.MakeEmptyMap(adt.AsStore(rt)).Root()
	builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to construct state")

	st := ConstructState(emptyMap, supervisor)
	rt.StateCreate(st)
	return nil
}

func (a Actor) ValidateGranted(rt runtime.Runtime, params *builtin.ValidateGrantedParams) *abi.EmptyValue {
	rt.ValidateImmediateCallerType(builtin.CallerTypesGoverned...)

	governor, ok := rt.ResolveAddress(params.Caller)
	builtin.RequireParam(rt, ok, "failed to resolve governor %s", params.Caller)

	codeID, ok := rt.GetActorCodeCID(rt.Caller())
	builtin.RequireParam(rt, ok, "failed to get actor code %s", rt.Caller())

	var st State
	rt.StateReadonly(&st)
	store := adt.AsStore(rt)

	governors, err := adt.AsMap(store, st.Governors)
	builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to load governor")

	granted, err := st.IsGranted(store, governors, governor, codeID, params.Method)
	builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to check granted")

	builtin.RequireParam(rt, granted, "forbidden")

	return nil
}

type GrantOrRevokeParams struct {
	Governor    address.Address
	Authorities []Authority
}

type Authority struct {
	ActorCodeID cid.Cid
	Methods     []abi.MethodNum
}

// Grant all privileges to specified Governor if Authorities is empty.
//
// Grant all privileges on ActorCodeID to specified Governor if Methods is empty.
func (a Actor) Grant(rt runtime.Runtime, params *GrantOrRevokeParams) *abi.EmptyValue {

	governor, targetCodeMethods := checkGrantOrRevokeParams(rt, params)
	builtin.RequireParam(rt, len(targetCodeMethods) != 0, "no priviledge to grant")

	var st State
	rt.StateTransaction(&st, func() {
		builtin.RequireParam(rt, st.Supervisor == rt.Caller(), "forbidden")

		store := adt.AsStore(rt)

		governors, err := adt.AsMap(store, st.Governors)
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to load governors")

		err = st.GrantOrRevoke(store, governors, governor, targetCodeMethods, true)
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to grant")

		st.Governors, err = governors.Root()
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to flush governors")
	})
	return nil
}

// Revoke all privileges from specified Governor if Authorities is empty.
//
// Revoke all privileges on ActorCodeID from specified Governor if Methods is empty.
func (a Actor) Revoke(rt runtime.Runtime, params *GrantOrRevokeParams) *abi.EmptyValue {

	governor, targetCodeMethods := checkGrantOrRevokeParams(rt, params)
	builtin.RequireParam(rt, len(targetCodeMethods) != 0, "no priviledge to revoke")

	var st State
	rt.StateTransaction(&st, func() {
		builtin.RequireParam(rt, st.Supervisor == rt.Caller(), "forbidden")

		store := adt.AsStore(rt)

		governors, err := adt.AsMap(store, st.Governors)
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to load governors")

		err = st.GrantOrRevoke(store, governors, governor, targetCodeMethods, false)
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to revoke")

		st.Governors, err = governors.Root()
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalState, "failed to flush governors")
	})
	return nil
}

func checkGrantOrRevokeParams(rt runtime.Runtime, params *GrantOrRevokeParams) (address.Address, map[cid.Cid][]abi.MethodNum) {
	governor, ok := rt.ResolveAddress(params.Governor)
	builtin.RequireParam(rt, ok, "failed to resovle governor")

	target := make(map[cid.Cid][]abi.MethodNum)
	// builtin.RequireParam(rt, len(params.Authorities) != 0, "no authorities")

	if len(params.Authorities) != 0 {
		seenCodeID := make(map[cid.Cid]struct{})
		seenMethod := make(map[abi.MethodNum]struct{})
		for _, info := range params.Authorities {

			_, ok := seenCodeID[info.ActorCodeID]
			builtin.RequireParam(rt, !ok, "duplicated actor code %s", info.ActorCodeID)
			seenCodeID[info.ActorCodeID] = struct{}{}

			governedMethods, ok := GovernedActors[info.ActorCodeID]
			builtin.RequireParam(rt, ok, "actor code %s not found", info.ActorCodeID)

			if len(info.Methods) != 0 {
				for _, method := range info.Methods {
					_, ok = seenMethod[method]
					builtin.RequireParam(rt, !ok, "duplicated method %s", method)
					seenMethod[method] = struct{}{}

					_, ok = governedMethods[method]
					builtin.RequireParam(rt, ok, "method %d of actor code %s not found", method, info.ActorCodeID)
				}
				target[info.ActorCodeID] = info.Methods
			} else {
				// fill in all priviledges on this actor
				for method := range governedMethods {
					target[info.ActorCodeID] = append(target[info.ActorCodeID], method)
				}
			}
		}
	} else {
		// fill in all priviledges on all actors
		for code, methods := range GovernedActors {
			for method := range methods {
				target[code] = append(target[code], method)
			}
		}
	}
	return governor, target
}
