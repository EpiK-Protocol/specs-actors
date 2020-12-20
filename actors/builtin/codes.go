package builtin

import (
	"sort"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// The built-in actor code IDs
var (
	SystemActorCodeID         cid.Cid
	InitActorCodeID           cid.Cid
	CronActorCodeID           cid.Cid
	AccountActorCodeID        cid.Cid
	StoragePowerActorCodeID   cid.Cid
	StorageMinerActorCodeID   cid.Cid
	StorageMarketActorCodeID  cid.Cid
	PaymentChannelActorCodeID cid.Cid
	MultisigActorCodeID       cid.Cid
	RewardActorCodeID         cid.Cid
	GovernActorCodeID         cid.Cid
	CallerTypesSignable       []cid.Cid
	ExpertActorCodeID         cid.Cid
	ExpertFundActorCodeID     cid.Cid
	VoteActorCodeID           cid.Cid
	RetrievalActorCodeID      cid.Cid
	KnowledgeActorCodeID      cid.Cid
	CallerTypesGoverned       []cid.Cid
)

var builtinActors map[cid.Cid]*actorInfo

type actorInfo struct {
	name   string
	signer bool
}

func init() {
	builder := cid.V1Builder{Codec: cid.Raw, MhType: mh.IDENTITY}
	builtinActors = make(map[cid.Cid]*actorInfo)

	for id, info := range map[*cid.Cid]*actorInfo{ //nolint:nomaprange
		&SystemActorCodeID:         {name: "epk/1/system"},
		&InitActorCodeID:           {name: "epk/1/init"},
		&CronActorCodeID:           {name: "epk/1/cron"},
		&StoragePowerActorCodeID:   {name: "epk/1/storagepower"},
		&StorageMinerActorCodeID:   {name: "epk/1/storageminer"},
		&StorageMarketActorCodeID:  {name: "epk/1/storagemarket"},
		&PaymentChannelActorCodeID: {name: "epk/1/paymentchannel"},
		&RewardActorCodeID:         {name: "epk/1/reward"},
		&GovernActorCodeID:         {name: "epk/1/govern"},
		&AccountActorCodeID:        {name: "epk/1/account", signer: true},
		&MultisigActorCodeID:       {name: "epk/1/multisig", signer: true},
		&ExpertActorCodeID:         {name: "epk/1/expert"},
		&ExpertFundActorCodeID:     {name: "epk/1/expertfund"},
		&VoteActorCodeID:           {name: "epk/1/vote"},
		&KnowledgeActorCodeID:      {name: "epk/1/knowledge"},
		&RetrievalActorCodeID:      {name: "epk/1/retrieval"},
	} {
		c, err := builder.Sum([]byte(info.name))
		if err != nil {
			panic(err)
		}
		*id = c
		builtinActors[c] = info
	}

	// Set of actor code types that can represent external signing parties.
	for id, info := range builtinActors { //nolint:nomaprange
		if info.signer {
			CallerTypesSignable = append(CallerTypesSignable, id)
		}
	}
	sort.Slice(CallerTypesSignable, func(i, j int) bool {
		return CallerTypesSignable[i].KeyString() < CallerTypesSignable[j].KeyString()
	})

	CallerTypesGoverned = []cid.Cid{
		StoragePowerActorCodeID,
		StorageMinerActorCodeID,
		StorageMarketActorCodeID,
		ExpertActorCodeID,
		ExpertFundActorCodeID,
		VoteActorCodeID,
		RetrievalActorCodeID,
		KnowledgeActorCodeID,
	}
}

// IsBuiltinActor returns true if the code belongs to an actor defined in this repo.
func IsBuiltinActor(code cid.Cid) bool {
	_, isBuiltin := builtinActors[code]
	return isBuiltin
}

// ActorNameByCode returns the (string) name of the actor given a cid code.
func ActorNameByCode(code cid.Cid) string {
	if !code.Defined() {
		return "<undefined>"
	}

	info, ok := builtinActors[code]
	if !ok {
		return "<unknown>"
	}
	return info.name
}

// Tests whether a code CID represents an actor that can be an external principal: i.e. an account or multisig.
// We could do something more sophisticated here: https://github.com/filecoin-project/specs-actors/issues/178
func IsPrincipal(code cid.Cid) bool {
	info, ok := builtinActors[code]
	if !ok {
		return false
	}
	return info.signer
}
