package market_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	address "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"

	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/market"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
	"github.com/filecoin-project/specs-actors/v2/support/mock"
	tutil "github.com/filecoin-project/specs-actors/v2/support/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cbgFalse = cbg.CborBool(false)
	cbgTrue  = cbg.CborBool(true)
)

func mustCbor(o cbor.Marshaler) []byte {
	buf := new(bytes.Buffer)
	if err := o.MarshalCBOR(buf); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func TestExports(t *testing.T) {
	mock.CheckActorExports(t, market.Actor{})
}

func TestRemoveAllError(t *testing.T) {
	marketActor := tutil.NewIDAddr(t, 100)
	builder := mock.NewBuilder(context.Background(), marketActor)
	rt := builder.Build(t)
	store := adt.AsStore(rt)

	smm, err := market.MakeEmptySetMultimap(store, builtin.DefaultHamtBitwidth)
	require.NoError(t, err)

	if err := smm.RemoveAll(42); err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
}

func TestMarketActor(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	minerAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}

	var st market.State

	t.Run("simple construction", func(t *testing.T) {
		actor := market.Actor{}
		receiver := tutil.NewIDAddr(t, 100)
		builder := mock.NewBuilder(context.Background(), receiver).
			WithCaller(builtin.SystemActorAddr, builtin.InitActorCodeID)

		rt := builder.Build(t)

		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, nil).(*abi.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)

		emptyBalanceTable, err := adt.StoreEmptyMap(store, adt.BalanceTableBitwidth)
		assert.NoError(t, err)

		emptyMap, err := adt.StoreEmptyMap(store, builtin.DefaultHamtBitwidth)
		assert.NoError(t, err)

		emptyProposalsArrayCid, err := adt.StoreEmptyArray(store, market.ProposalsAmtBitwidth)
		assert.NoError(t, err)

		emptyStatesArrayCid, err := adt.StoreEmptyArray(store, market.StatesAmtBitwidth)
		assert.NoError(t, err)

		emptyMultiMap, err := market.StoreEmptySetMultimap(store, builtin.DefaultHamtBitwidth)
		assert.NoError(t, err)

		var state market.State
		rt.GetState(&state)

		assert.Equal(t, emptyProposalsArrayCid, state.Proposals)
		assert.Equal(t, emptyStatesArrayCid, state.States)
		assert.Equal(t, emptyMap, state.PendingProposals)
		assert.Equal(t, emptyBalanceTable, state.EscrowTable)
		assert.Equal(t, emptyBalanceTable, state.LockedTable)
		assert.Equal(t, abi.DealID(0), state.NextID)
		assert.Equal(t, emptyMultiMap, state.DealOpsByEpoch)
		assert.Equal(t, emptyMap, state.Quotas)
		assert.Equal(t, abi.ChainEpoch(-1), state.LastCron)
	})

	t.Run("AddBalance", func(t *testing.T) {
		t.Run("adds to provider escrow funds", func(t *testing.T) {
			testCases := []struct {
				delta int64
				total int64
			}{
				{10, 10},
				{20, 30},
				{40, 70},
			}

			// Test adding provider funds from both worker and owner address
			for _, callerAddr := range []address.Address{owner, worker} {
				rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

				for _, tc := range testCases {
					rt.SetCaller(callerAddr, builtin.AccountActorCodeID)
					rt.SetReceived(abi.NewTokenAmount(tc.delta))
					rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)
					expectGetControlAddresses(rt, provider, owner, worker, coinbase)

					rt.Call(actor.AddBalance, &provider)

					rt.Verify()

					rt.GetState(&st)
					assert.Equal(t, abi.NewTokenAmount(tc.total), actor.getEscrowBalance(rt, provider))

					actor.checkState(rt)
				}
			}
		})

		t.Run("fails unless called by an account actor", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

			rt.SetReceived(abi.NewTokenAmount(10))
			rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)

			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.ExpectAbort(exitcode.SysErrForbidden, func() {
				rt.Call(actor.AddBalance, &provider)
			})

			rt.Verify()

			actor.checkState(rt)
		})

		t.Run("adds to non-provider escrow funds", func(t *testing.T) {
			testCases := []struct {
				delta int64
				total int64
			}{
				{10, 10},
				{20, 30},
				{40, 70},
			}

			// Test adding non-provider funds from both worker and client addresses
			for _, callerAddr := range []address.Address{client, worker} {
				rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

				for _, tc := range testCases {
					rt.SetCaller(callerAddr, builtin.AccountActorCodeID)
					rt.SetReceived(abi.NewTokenAmount(tc.delta))
					rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)

					rt.Call(actor.AddBalance, &callerAddr)

					rt.Verify()

					rt.GetState(&st)
					assert.Equal(t, abi.NewTokenAmount(tc.total), actor.getEscrowBalance(rt, callerAddr))

					actor.checkState(rt)
				}
			}
		})

		t.Run("fail when balance is zero", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

			rt.SetCaller(tutil.NewIDAddr(t, 101), builtin.AccountActorCodeID)
			rt.SetReceived(big.Zero())

			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.AddBalance, &provider)
			})
			rt.Verify()

			actor.checkState(rt)
		})
	})

	t.Run("WithdrawBalance", func(t *testing.T) {
		startEpoch := abi.ChainEpoch(10)
		// endEpoch := startEpoch + 200*builtin.EpochsInDay
		publishEpoch := abi.ChainEpoch(5)

		t.Run("fails with a negative withdraw amount", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

			params := market.WithdrawBalanceParams{
				ProviderOrClientAddress: provider,
				Amount:                  abi.NewTokenAmount(-1),
			}

			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.WithdrawBalance, &params)
			})

			rt.Verify()
			actor.checkState(rt)
		})

		t.Run("fails if withdraw from non provider funds is not initiated by the recipient", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			actor.addParticipantFunds(rt, client, abi.NewTokenAmount(20))

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, client))

			rt.ExpectValidateCallerAddr(client)
			params := market.WithdrawBalanceParams{
				ProviderOrClientAddress: client,
				Amount:                  abi.NewTokenAmount(1),
			}

			// caller is not the recipient
			rt.SetCaller(tutil.NewIDAddr(t, 909), builtin.AccountActorCodeID)
			rt.ExpectAbort(exitcode.SysErrForbidden, func() {
				rt.Call(actor.WithdrawBalance, &params)
			})
			rt.Verify()

			// verify there was no withdrawal
			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, client))

			actor.checkState(rt)
		})

		t.Run("fails if withdraw from provider funds is not initiated by the owner or worker", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			actor.addProviderFunds(rt, abi.NewTokenAmount(20), minerAddrs)

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, provider))

			// only signing parties can add balance for client AND provider.
			rt.ExpectValidateCallerAddr(owner, worker)
			params := market.WithdrawBalanceParams{
				ProviderOrClientAddress: provider,
				Amount:                  abi.NewTokenAmount(1),
			}

			// caller is not owner or worker
			rt.SetCaller(tutil.NewIDAddr(t, 909), builtin.AccountActorCodeID)
			expectGetControlAddresses(rt, provider, owner, worker, coinbase)

			rt.ExpectAbort(exitcode.SysErrForbidden, func() {
				rt.Call(actor.WithdrawBalance, &params)
			})
			rt.Verify()

			// verify there was no withdrawal
			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, provider))

			actor.checkState(rt)
		})

		t.Run("withdraws from provider escrow funds and sends to owner", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

			actor.addProviderFunds(rt, abi.NewTokenAmount(20), minerAddrs)

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, provider))

			// worker calls WithdrawBalance, balance is transferred to owner
			withdrawAmount := abi.NewTokenAmount(1)
			actor.withdrawProviderBalance(rt, withdrawAmount, withdrawAmount, minerAddrs)

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(19), actor.getEscrowBalance(rt, provider))

			actor.checkState(rt)
		})

		t.Run("withdraws from non-provider escrow funds", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			actor.addParticipantFunds(rt, client, abi.NewTokenAmount(20))

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, client))

			withdrawAmount := abi.NewTokenAmount(1)
			actor.withdrawClientBalance(rt, client, withdrawAmount, withdrawAmount)

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(19), actor.getEscrowBalance(rt, client))

			actor.checkState(rt)
		})

		t.Run("client withdrawing more than escrow balance limits to available funds", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			actor.addParticipantFunds(rt, client, abi.NewTokenAmount(20))

			// withdraw amount greater than escrow balance
			withdrawAmount := abi.NewTokenAmount(25)
			expectedAmount := abi.NewTokenAmount(20)
			actor.withdrawClientBalance(rt, client, withdrawAmount, expectedAmount)

			actor.assertAccountBalance(rt, client, big.Zero(), big.Zero())

			actor.checkState(rt)
		})

		t.Run("worker withdrawing more than escrow balance limits to available funds", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			actor.addProviderFunds(rt, abi.NewTokenAmount(20), minerAddrs)

			rt.GetState(&st)
			assert.Equal(t, abi.NewTokenAmount(20), actor.getEscrowBalance(rt, provider))

			// withdraw amount greater than escrow balance
			withdrawAmount := abi.NewTokenAmount(25)
			actualWithdrawn := abi.NewTokenAmount(20)
			actor.withdrawProviderBalance(rt, withdrawAmount, actualWithdrawn, minerAddrs)

			actor.assertAccountBalance(rt, provider, big.Zero(), big.Zero())

			actor.checkState(rt)
		})

		t.Run("balance after withdrawal must ALWAYS be greater than or equal to locked amount", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			rt.SetAddressActorType(expert, builtin.ExpertActorCodeID)

			// publish the deal so that client AND provider collateral is locked
			rt.SetEpoch(publishEpoch)
			dealId := actor.generateAndPublishDealWithFunds(rt, client, minerAddrs, startEpoch, startEpoch, big.Zero(), big.Zero())
			deal := actor.getDealProposal(rt, dealId)
			rt.GetState(&st)
			require.Equal(t, deal.StartEpoch, startEpoch)
			/* require.Equal(t, deal.ProviderCollateral, actor.getEscrowBalance(rt, provider))
			require.Equal(t, deal.ClientBalanceRequirement(), actor.getEscrowBalance(rt, client)) */

			withDrawAmt := abi.NewTokenAmount(1)
			withDrawableAmt := abi.NewTokenAmount(0)
			// client cannot withdraw any funds since all it's balance is locked
			actor.withdrawClientBalance(rt, client, withDrawAmt, withDrawableAmt)
			//  provider cannot withdraw any funds since all it's balance is locked
			actor.withdrawProviderBalance(rt, withDrawAmt, withDrawableAmt, minerAddrs)

			// add some more funds to the provider & ensure withdrawal is limited by the locked funds
			withDrawAmt = abi.NewTokenAmount(30)
			withDrawableAmt = abi.NewTokenAmount(25)
			actor.addProviderFunds(rt, withDrawableAmt, minerAddrs)
			actor.withdrawProviderBalance(rt, withDrawAmt, withDrawableAmt, minerAddrs)

			// add some more funds to the client & ensure withdrawal is limited by the locked funds
			actor.addParticipantFunds(rt, client, withDrawableAmt)
			actor.withdrawClientBalance(rt, client, withDrawAmt, withDrawableAmt)

			actor.checkState(rt)
		})

		t.Run("worker balance after withdrawal must account for slashed funds", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

			// publish deal
			rt.SetEpoch(publishEpoch)
			dealID := actor.generateAndPublishDealWithFunds(rt, client, minerAddrs, startEpoch, startEpoch, big.Zero(), big.Zero())

			// activate the deal
			actor.activateDeals(rt /* endEpoch+1, */, provider, publishEpoch, dealID)
			st := actor.getDealState(rt, dealID)
			require.EqualValues(t, publishEpoch, st.SectorStartEpoch)

			// slash the deal
			newEpoch := publishEpoch + 1
			rt.SetEpoch(newEpoch)
			actor.terminateDeals(rt, provider, dealID)
			st = actor.getDealState(rt, dealID)
			require.EqualValues(t, publishEpoch+1, st.SlashEpoch)

			// provider cannot withdraw any funds since all it's balance is locked
			withDrawAmt := abi.NewTokenAmount(1)
			actualWithdrawn := abi.NewTokenAmount(0)
			actor.withdrawProviderBalance(rt, withDrawAmt, actualWithdrawn, minerAddrs)

			// add some more funds to the provider & ensure withdrawal is limited by the locked funds
			actor.addProviderFunds(rt, abi.NewTokenAmount(25), minerAddrs)
			withDrawAmt = abi.NewTokenAmount(30)
			actualWithdrawn = abi.NewTokenAmount(25)

			actor.withdrawProviderBalance(rt, withDrawAmt, actualWithdrawn, minerAddrs)

			actor.checkState(rt)
		})
	})
}

func TestPublishStorageDeals(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAddr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	startEpoch := abi.ChainEpoch(42)
	control := tutil.NewIDAddr(t, 200)
	mAddr := &minerAddrs{owner, worker, coinbase, provider, expertAddr, []address.Address{control}}

	t.Run("simple deal", func(t *testing.T) {
		startEpoch := abi.ChainEpoch(1000)
		publishEpoch := abi.ChainEpoch(1)

		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(publishEpoch)

		// Publish from miner worker.
		deal1 := actor.generateRandDealAndAddFunds(rt, client, mAddr, startEpoch)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		_ = actor.publishDeals(rt, mAddr, publishDealReq{deal: deal1})

		// Publish from miner control address.
		deal2 := actor.generateRandDealAndAddFunds(rt, client, mAddr, startEpoch+1)
		rt.SetCaller(control, builtin.AccountActorCodeID)
		_ = actor.publishDeals(rt, mAddr, publishDealReq{deal: deal2})

		actor.checkState(rt)
	})

	t.Run("provider and client addresses are resolved before persisting state and sent to VerigReg actor for a verified deal", func(t *testing.T) {
		// provider addresses
		providerBls := tutil.NewBLSAddr(t, 101)
		providerResolved := tutil.NewIDAddr(t, 102)
		// client addresses
		clientBls := tutil.NewBLSAddr(t, 900)
		clientResolved := tutil.NewIDAddr(t, 333)
		mAddr := &minerAddrs{owner, worker, coinbase, providerBls, expertAddr, nil}

		rt, actor := basicMarketSetup(t, owner, providerResolved, worker, clientResolved, coinbase)
		// mappings for resolving address
		rt.AddIDAddress(providerBls, providerResolved)
		rt.AddIDAddress(clientBls, clientResolved)

		// generate deal and add required funds for deal
		startEpoch := abi.ChainEpoch(42)
		/* endEpoch := startEpoch + 200*builtin.EpochsInDay */
		deal := generateDealProposal(clientBls, mAddr.provider, startEpoch /* , endEpoch */)
		/* deal.VerifiedDeal = true

		// add funds for cient using it's BLS address -> will be resolved and persisted
		actor.addParticipantFunds(rt, clientBls, deal.ClientBalanceRequirement())
		require.EqualValues(t, deal.ClientBalanceRequirement(), actor.getEscrowBalance(rt, clientResolved)) */
		require.EqualValues(t, big.Zero(), actor.getEscrowBalance(rt, clientResolved))

		// add funds for provider using it's BLS address -> will be resolved and persisted
		rt.SetReceived(big.NewInt(10))
		rt.SetCaller(mAddr.owner, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)
		expectGetControlAddresses(rt, providerResolved, mAddr.owner, mAddr.worker, mAddr.coinbase)
		rt.Call(actor.AddBalance, &mAddr.provider)
		rt.Verify()
		/* rt.SetBalance(big.Add(rt.Balance(), deal.ProviderCollateral))
		require.EqualValues(t, deal.ProviderCollateral, actor.getEscrowBalance(rt, providerResolved)) */
		require.EqualValues(t, big.NewInt(10), actor.getEscrowBalance(rt, providerResolved))

		// publish deal using the BLS addresses
		rt.SetCaller(mAddr.worker, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)
		expectGetControlAddresses(rt, providerResolved, mAddr.owner, mAddr.worker, mAddr.coinbase)
		batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: deal.PieceCID}}}
		rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
		rt.ExpectSend(providerResolved, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)
		// expectQueryNetworkInfo(rt, actor)
		//  create a client proposal with a valid signature
		params := market.PublishStorageDealsParams{
			DataRef: market.PublishStorageDataRef{
				Expert: expertAddr.String(),
			},
		}
		buf := bytes.Buffer{}
		require.NoError(t, deal.MarshalCBOR(&buf), "failed to marshal deal proposal")
		sig := crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte("does not matter")}
		clientProposal := market.ClientDealProposal{Proposal: deal, ClientSignature: sig}
		params.Deals = append(params.Deals, clientProposal)
		// expect a call to verify the above signature
		rt.ExpectVerifySignature(sig, deal.Client, buf.Bytes(), nil)

		/* // request is sent to the VerigReg actor using the resolved address
		param := &verifreg.UseBytesParams{
			Address:  clientResolved,
			DealSize: big.NewIntUnsigned(uint64(deal.PieceSize)),
		}
		rt.ExpectSend(builtin.VerifiedRegistryActorAddr, builtin.MethodsVerifiedRegistry.UseBytes, param, abi.NewTokenAmount(0), nil, exitcode.Ok) */

		deal2 := deal
		deal2.Client = clientResolved
		deal2.Provider = providerResolved
		actor.expectGetRandom(rt, &deal2, abi.ChainEpoch(100))

		ret := rt.Call(actor.PublishStorageDeals, &params)
		rt.Verify()
		resp, ok := ret.(*market.PublishStorageDealsReturn)
		require.True(t, ok)
		dealId := resp.IDs[0]

		// assert that deal is persisted with the resolved addresses
		prop := actor.getDealProposal(rt, dealId)
		require.EqualValues(t, clientResolved, prop.Client)
		require.EqualValues(t, providerResolved, prop.Provider)

		ss := actor.checkState(rt)
		require.Equal(t, uint64(1), ss.PendingProposalCount)
		require.Equal(t, uint64(1), ss.DealOpCount)
		require.Equal(t, 1, len(ss.DealOpEpochStats))
		require.Len(t, ss.Quotas, 0)
		require.Equal(t, startEpoch, ss.Deals[0].StartEpoch)
		require.Equal(t, providerResolved, ss.Deals[0].Provider)
	})

	t.Run("publish a deal after activating a previous deal which has a start epoch far in the future", func(t *testing.T) {
		startEpoch := abi.ChainEpoch(1000)
		/* endEpoch := startEpoch + 200*builtin.EpochsInDay */
		publishEpoch := abi.ChainEpoch(1)

		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

		// publish the deal and activate it
		rt.SetEpoch(publishEpoch)
		deal1ID := actor.generateAndPublishRandDeal(rt, client, mAddr, startEpoch, startEpoch)
		actor.activateDeals(rt /* endEpoch, */, provider, publishEpoch, deal1ID)
		st := actor.getDealState(rt, deal1ID)
		require.EqualValues(t, publishEpoch, st.SectorStartEpoch)

		// now publish a second deal and activate it
		newEpoch := publishEpoch + 1
		rt.SetEpoch(newEpoch)
		deal2ID := actor.generateAndPublishRandDeal(rt, client, mAddr, startEpoch+1, startEpoch+1)
		actor.activateDeals(rt /* endEpoch+1, */, provider, newEpoch, deal2ID)

		ss := actor.checkState(rt)
		require.Len(t, ss.Quotas, 2) //
		proposal1 := actor.getDealProposal(rt, deal1ID)
		require.Equal(t, ss.Quotas[proposal1.PieceCID], uint64(market.DefaultInitialQuota-1))
		proposal2 := actor.getDealProposal(rt, deal2ID)
		require.Equal(t, ss.Quotas[proposal2.PieceCID], uint64(market.DefaultInitialQuota-1))
	})

	t.Run("publish a deal with enough collateral when circulating supply > 0", func(t *testing.T) {
		startEpoch := abi.ChainEpoch(1000)
		/* endEpoch := startEpoch + 200*builtin.EpochsInDay */
		publishEpoch := abi.ChainEpoch(1)

		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

		/*clientCollateral := abi.NewTokenAmount(10) // min is zero so this is placeholder

		// given power and circ supply cancel this should be 1*dealqapower / 100
		dealSize := abi.PaddedPieceSize(2048) // generateDealProposal's deal size
		providerCollateral := big.Div(
			big.Mul(big.NewInt(int64(dealSize)), market.ProviderCollateralSupplyTarget.Numerator),
			market.ProviderCollateralSupplyTarget.Denominator,
		)
		providerCollateral := market.DealMinProviderCollateral*/
		deal := actor.generateDealWithCollateralAndAddFunds(rt, client, mAddr, abi.NewTokenAmount(10), abi.NewTokenAmount(10), startEpoch /* , endEpoch */)
		rt.SetCirculatingSupply(actor.networkQAPower) // convenient for these two numbers to cancel out

		// publish the deal successfully
		rt.SetEpoch(publishEpoch)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, mAddr, publishDealReq{deal: deal})

		actor.checkState(rt)
	})

	t.Run("publish multiple deals for different clients and ensure balances are correct", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		client1 := tutil.NewIDAddr(t, 900)
		client2 := tutil.NewIDAddr(t, 901)
		client3 := tutil.NewIDAddr(t, 902)

		// generate first deal for
		deal1 := actor.generateRandDealAndAddFunds(rt, client1, mAddr, startEpoch /* , endEpoch */)

		// generate second deal
		deal2 := actor.generateRandDealAndAddFunds(rt, client2, mAddr, startEpoch /* , endEpoch */)

		// generate third deal
		deal3 := actor.generateRandDealAndAddFunds(rt, client3, mAddr, startEpoch /* , endEpoch */)

		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, mAddr, publishDealReq{deal: deal1}, publishDealReq{deal: deal2},
			publishDealReq{deal: deal3})

		/* // assert locked balance for all clients and provider
		providerLocked := big.Sum(deal1.ProviderCollateral, deal2.ProviderCollateral, deal3.ProviderCollateral)
		client1Locked := actor.getLockedBalance(rt, client1)
		client2Locked := actor.getLockedBalance(rt, client2)
		client3Locked := actor.getLockedBalance(rt, client3) */
		require.EqualValues(t, big.Zero(), actor.getLockedBalance(rt, client1))
		require.EqualValues(t, big.NewInt(10), actor.getEscrowBalance(rt, client1))
		require.EqualValues(t, big.Zero(), actor.getLockedBalance(rt, client2))
		require.EqualValues(t, big.NewInt(10), actor.getEscrowBalance(rt, client2))
		require.EqualValues(t, big.Zero(), actor.getLockedBalance(rt, client3))
		require.EqualValues(t, big.NewInt(10), actor.getEscrowBalance(rt, client3))
		require.EqualValues(t, big.Zero(), actor.getLockedBalance(rt, provider))
		require.EqualValues(t, big.NewInt(30), actor.getEscrowBalance(rt, provider))

		/* // assert locked funds dealStates
		rt.GetState(&st)
		totalClientCollateralLocked := big.Sum(deal3.ClientCollateral, deal1.ClientCollateral, deal2.ClientCollateral)
		require.EqualValues(t, totalClientCollateralLocked, st.TotalClientLockedCollateral)
		require.EqualValues(t, providerLocked, st.TotalProviderLockedCollateral)
		totalStorageFee := big.Sum(deal1.TotalStorageFee(), deal2.TotalStorageFee(), deal3.TotalStorageFee())
		require.EqualValues(t, totalStorageFee, st.TotalClientStorageFee)*/

		// publish two more deals for same clients with same provider
		deal4 := actor.generateRandDealAndAddFunds(rt, client3, mAddr, abi.ChainEpoch(1000) /* , abi.ChainEpoch(1000+200*builtin.EpochsInDay) */)
		deal5 := actor.generateRandDealAndAddFunds(rt, client3, mAddr, abi.ChainEpoch(100) /* , abi.ChainEpoch(100+200*builtin.EpochsInDay) */)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, mAddr, publishDealReq{deal: deal4}, publishDealReq{deal: deal5})

		require.EqualValues(t, big.NewInt(30), actor.getEscrowBalance(rt, client3))
		/* // assert locked balances for clients and provider
		rt.GetState(&st)
		providerLocked = big.Sum(providerLocked, deal4.ProviderCollateral, deal5.ProviderCollateral)
		require.EqualValues(t, providerLocked, actor.getLockedBalance(rt, provider))

		client3LockedUpdated := actor.getLockedBalance(rt, client3)
		require.EqualValues(t, big.Sum(client3Locked, deal4.ClientBalanceRequirement(), deal5.ClientBalanceRequirement()), client3LockedUpdated)

		client1Locked = actor.getLockedBalance(rt, client1)
		client2Locked = actor.getLockedBalance(rt, client2)
		require.EqualValues(t, deal1.ClientBalanceRequirement(), client1Locked)
		require.EqualValues(t, deal2.ClientBalanceRequirement(), client2Locked)

		// assert locked funds dealStates
		totalClientCollateralLocked = big.Sum(totalClientCollateralLocked, deal4.ClientCollateral, deal5.ClientCollateral)
		require.EqualValues(t, totalClientCollateralLocked, st.TotalClientLockedCollateral)
		require.EqualValues(t, providerLocked, st.TotalProviderLockedCollateral)

		totalStorageFee = big.Sum(totalStorageFee, deal4.TotalStorageFee(), deal5.TotalStorageFee())
		require.EqualValues(t, totalStorageFee, st.TotalClientStorageFee) */

		// PUBLISH DEALS with a different provider
		provider2 := tutil.NewIDAddr(t, 109)
		expert := tutil.NewIDAddr(t, 110)
		miner := &minerAddrs{owner, worker, coinbase, provider2, expert, nil}

		// generate first deal for second provider
		deal6 := actor.generateRandDealAndAddFunds(rt, client1, miner, abi.ChainEpoch(20) /* , abi.ChainEpoch(20+200*builtin.EpochsInDay) */)

		// generate second deal for second provider
		deal7 := actor.generateRandDealAndAddFunds(rt, client1, miner, abi.ChainEpoch(25) /* , abi.ChainEpoch(60+200*builtin.EpochsInDay) */)

		// publish both the deals for the second provider
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, miner, publishDealReq{deal: deal6}, publishDealReq{deal: deal7})

		/* // assertions
		rt.GetState(&st)
		provider2Locked := big.Add(deal6.ProviderCollateral, deal7.ProviderCollateral)
		require.EqualValues(t, provider2Locked, actor.getLockedBalance(rt, provider2))
		client1LockedUpdated := actor.getLockedBalance(rt, client1)
		require.EqualValues(t, big.Add(deal7.ClientBalanceRequirement(), big.Add(client1Locked, deal6.ClientBalanceRequirement())), client1LockedUpdated)

		// assert first provider's balance as well
		require.EqualValues(t, providerLocked, actor.getLockedBalance(rt, provider))

		totalClientCollateralLocked = big.Add(totalClientCollateralLocked, big.Add(deal6.ClientCollateral, deal7.ClientCollateral))
		require.EqualValues(t, totalClientCollateralLocked, st.TotalClientLockedCollateral)
		require.EqualValues(t, big.Add(providerLocked, provider2Locked), st.TotalProviderLockedCollateral)
		totalStorageFee = big.Add(totalStorageFee, big.Add(deal6.TotalStorageFee(), deal7.TotalStorageFee()))
		require.EqualValues(t, totalStorageFee, st.TotalClientStorageFee) */

		require.EqualValues(t, big.NewInt(20), actor.getEscrowBalance(rt, provider2))
		require.EqualValues(t, big.NewInt(30), actor.getEscrowBalance(rt, client1))

		actor.checkState(rt)
	})
}

func TestPublishStorageDealsFailures(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAddr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expertAddr, nil}

	currentEpoch := abi.ChainEpoch(5)
	startEpoch := abi.ChainEpoch(10)
	/* endEpoch := startEpoch + 200*builtin.EpochsInDay */

	// simple failures because of invalid deal params
	{
		tcs := map[string]struct {
			setup                      func(*mock.Runtime, *marketActorTestHarness, *market.DealProposal)
			exitCode                   exitcode.ExitCode
			signatureVerificationError error
		}{
			/* "deal end after deal start": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.StartEpoch = 10
					d.EndEpoch = 9
				},
				exitCode: exitcode.ErrIllegalArgument,
			}, */
			"current epoch greater than start epoch": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.StartEpoch = currentEpoch - 1
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			/* "deal duration greater than max deal duration": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.StartEpoch = abi.ChainEpoch(10)
					d.EndEpoch = d.StartEpoch + (540 * builtin.EpochsInDay) + 1
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"negative price per epoch": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.StoragePricePerEpoch = abi.NewTokenAmount(-1)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"price per epoch greater than total filecoin": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.StoragePricePerEpoch = big.Add(builtin.TotalEPK, big.NewInt(1))
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"negative provider collateral": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.ProviderCollateral = big.NewInt(-1)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"provider collateral greater than max collateral": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.ProviderCollateral = big.Add(builtin.TotalEPK, big.NewInt(1))
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"provider collateral less than bound": {
				setup: func(rt *mock.Runtime, h *marketActorTestHarness, d *market.DealProposal) {
					// with these two equal provider collatreal min is 5/100 * deal size
					rt.SetCirculatingSupply(h.networkQAPower)
					dealSize := big.NewInt(2048) // default deal size used
					providerMin := big.Div(
						big.Mul(dealSize, market.ProviderCollateralSupplyTarget.Numerator),
						market.ProviderCollateralSupplyTarget.Denominator,
					)
					providerMin = market.DealMinProviderCollateral
					d.ProviderCollateral = big.Sub(providerMin, big.NewInt(1))
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"negative client collateral": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.ClientCollateral = big.NewInt(-1)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"client collateral greater than max collateral": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.ClientCollateral = big.Add(builtin.TotalEPK, big.NewInt(1))
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"client does not have enough balance for collateral": {
				setup: func(rt *mock.Runtime, a *marketActorTestHarness, d *market.DealProposal) {
					a.addParticipantFunds(rt, client, big.Sub(d.ClientBalanceRequirement(), big.NewInt(1)))
					a.addProviderFunds(rt, d.ProviderCollateral, mAddrs)
				},
				exitCode: exitcode.ErrInsufficientFunds,
			},
			"provider does not have enough balance for collateral": {
				setup: func(rt *mock.Runtime, a *marketActorTestHarness, d *market.DealProposal) {
					a.addParticipantFunds(rt, client, d.ClientBalanceRequirement())
					a.addProviderFunds(rt, big.Sub(d.ProviderCollateral, big.NewInt(1)), mAddrs)
				},
				exitCode: exitcode.ErrInsufficientFunds,
			}, */
			"unable to resolve client address": {
				setup: func(_ *mock.Runtime, a *marketActorTestHarness, d *market.DealProposal) {
					d.Client = tutil.NewBLSAddr(t, 1)
				},
				exitCode: exitcode.ErrNotFound,
			},
			"signature is invalid": {
				setup: func(_ *mock.Runtime, a *marketActorTestHarness, d *market.DealProposal) {

				},
				exitCode:                   exitcode.ErrIllegalArgument,
				signatureVerificationError: errors.New("error"),
			},
			/* "no entry for client in locked  balance table": {
				setup: func(rt *mock.Runtime, a *marketActorTestHarness, d *market.DealProposal) {
					a.addProviderFunds(rt, d.ProviderCollateral, mAddrs)
				},
				exitCode: exitcode.ErrInsufficientFunds,
			},
			"no entry for provider in locked  balance table": {
				setup: func(rt *mock.Runtime, a *marketActorTestHarness, d *market.DealProposal) {
					a.addParticipantFunds(rt, client, d.ClientBalanceRequirement())
				},
				exitCode: exitcode.ErrInsufficientFunds,
			}, */
			"bad piece CID": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.PieceCID = tutil.MakeCID("random cid", nil)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"zero piece size": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.PieceSize = abi.PaddedPieceSize(0)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"piece size less than 128 bytes": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.PieceSize = abi.PaddedPieceSize(64)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
			"piece size is not a power of 2": {
				setup: func(_ *mock.Runtime, _ *marketActorTestHarness, d *market.DealProposal) {
					d.PieceSize = abi.PaddedPieceSize(254)
				},
				exitCode: exitcode.ErrIllegalArgument,
			},
		}

		for name, tc := range tcs {
			t.Run(name, func(t *testing.T) {
				_ = name
				rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
				dealProposal := generateDealProposal(client, provider, startEpoch /* , endEpoch */)
				rt.SetEpoch(currentEpoch)
				tc.setup(rt, actor, &dealProposal)
				params := mkPublishStorageParams(dealProposal)
				params.DataRef.Expert = expertAddr.String()

				rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
				rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)

				batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: dealProposal.PieceCID}}}
				rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
				rt.ExpectSend(provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)
				/* expectQueryNetworkInfo(rt, actor) */
				rt.SetCaller(worker, builtin.AccountActorCodeID)
				rt.ExpectVerifySignature(crypto.Signature{}, dealProposal.Client, mustCbor(&dealProposal), tc.signatureVerificationError)
				rt.ExpectAbort(tc.exitCode, func() {
					rt.Call(actor.PublishStorageDeals, params)
				})

				rt.Verify()
				actor.checkState(rt)
			})
		}
	}

	/* // fails when client or provider has some funds but not enough to cover a deal
	{
		t.Run("fail when client has some funds but not enough for a deal", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client)

			//
			actor.addParticipantFunds(rt, client, abi.NewTokenAmount(100))
			startEpoch := abi.ChainEpoch(42)
			deal1 := generateDealProposal(client, provider, startEpoch, startEpoch+200*builtin.EpochsInDay)
			actor.addProviderFunds(rt, deal1.ProviderCollateral, mAddrs)
			params := mkPublishStorageParams(deal1)

			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			expectQueryNetworkInfo(rt, actor)
			rt.SetCaller(worker, builtin.AccountActorCodeID)
			rt.ExpectVerifySignature(crypto.Signature{}, deal1.Client, mustCbor(&deal1), nil)
			rt.ExpectAbort(exitcode.ErrInsufficientFunds, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})

			rt.Verify()
			actor.checkState(rt)
		})

		t.Run("fail when provider has some funds but not enough for a deal", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client)

			actor.addProviderFunds(rt, abi.NewTokenAmount(1), mAddrs)
			deal1 := generateDealProposal(client, provider, startEpoch, endEpoch)
			actor.addParticipantFunds(rt, client, deal1.ClientBalanceRequirement())

			params := mkPublishStorageParams(deal1)

			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &miner.GetControlAddressesReturn{Worker: worker, Owner: owner}, 0)
			expectQueryNetworkInfo(rt, actor)
			rt.SetCaller(worker, builtin.AccountActorCodeID)
			rt.ExpectVerifySignature(crypto.Signature{}, deal1.Client, mustCbor(&deal1), nil)
			rt.ExpectAbort(exitcode.ErrInsufficientFunds, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})

			rt.Verify()
			actor.checkState(rt)
		})
	} */

	// fail when deals have different providers
	{
		t.Run("fail when deals have different providers", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			deal1 := actor.generateDealAndAddFunds(rt, client, mAddrs, startEpoch /* , endEpoch */)
			m2 := &minerAddrs{owner, worker, tutil.NewIDAddr(t, 999), tutil.NewIDAddr(t, 1000), tutil.NewIDAddr(t, 1001), nil}

			deal2 := actor.generateDealAndAddFunds(rt, client, m2, abi.ChainEpoch(1) /* , endEpoch */)

			params := mkPublishStorageParams(deal1, deal2)
			params.DataRef.Expert = expertAddr.String()

			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)

			batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{
				{CID: deal1.PieceCID},
				{CID: deal2.PieceCID},
			}}
			rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
			rt.ExpectSend(deal1.Provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)

			rt.SetCaller(worker, builtin.AccountActorCodeID)
			rt.ExpectVerifySignature(crypto.Signature{}, deal1.Client, mustCbor(&deal1), nil)
			rt.ExpectVerifySignature(crypto.Signature{}, deal2.Client, mustCbor(&deal2), nil)

			actor.expectGetRandom(rt, &deal1, abi.ChainEpoch(100))

			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})

			rt.Verify()
			actor.checkState(rt)
		})

		//  failures because of incorrect call params
		t.Run("fail when caller is not of signable type", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			params := mkPublishStorageParams(generateDealProposal(client, provider, startEpoch /* , endEpoch */))
			w := tutil.NewIDAddr(t, 1000)
			rt.SetCaller(w, builtin.StorageMinerActorCodeID)
			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			rt.ExpectAbort(exitcode.SysErrForbidden, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})
			actor.checkState(rt)
		})

		t.Run("fail when no deals in params", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			params := mkPublishStorageParams()
			rt.SetCaller(worker, builtin.AccountActorCodeID)
			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})
			actor.checkState(rt)
		})

		t.Run("fail to resolve provider address", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			deal := generateDealProposal(client, provider, startEpoch /* , endEpoch */)
			deal.Provider = tutil.NewBLSAddr(t, 100)

			params := mkPublishStorageParams(deal)
			rt.SetCaller(worker, builtin.AccountActorCodeID)
			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			rt.ExpectAbort(exitcode.ErrNotFound, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})
			actor.checkState(rt)
		})

		t.Run("caller is not the same as the worker address for miner", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			deal := generateDealProposal(client, provider, startEpoch /* , endEpoch */)
			params := mkPublishStorageParams(deal)
			rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
			rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: tutil.NewIDAddr(t, 999), Owner: owner, Coinbase: coinbase}, 0)
			rt.SetCaller(worker, builtin.AccountActorCodeID)
			rt.ExpectAbort(exitcode.ErrForbidden, func() {
				rt.Call(actor.PublishStorageDeals, params)
			})

			rt.Verify()
			actor.checkState(rt)
		})
	}

	t.Run("fails if provider is not a storage miner actor", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

		// deal provider will be a Storage Miner Actor.
		p2 := tutil.NewIDAddr(t, 505)
		rt.SetAddressActorType(p2, builtin.StoragePowerActorCodeID)
		deal := generateDealProposal(client, p2, abi.ChainEpoch(1) /* , abi.ChainEpoch(5) */)

		params := mkPublishStorageParams(deal)
		rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.PublishStorageDeals, params)
		})

		rt.Verify()
		actor.checkState(rt)
	})
}

func TestActivateDeals(t *testing.T) {

	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}

	startEpoch := abi.ChainEpoch(10)
	/* endEpoch := startEpoch + 200*builtin.EpochsInDay */
	currentEpoch := abi.ChainEpoch(5)
	/* sectorExpiry := endEpoch + 100 */

	t.Run("active deals multiple times with different providers", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		// provider 1 publishes deals1 and deals2 and deal3
		dealId1 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch, "deal1")
		dealId2 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+1, */, startEpoch, "deal2")
		dealId3 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+2, */, startEpoch, "deal3")

		// provider2 publishes deal4 and deal5
		provider2 := tutil.NewIDAddr(t, 401)
		mAddrs.provider = provider2
		dealId4 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch, "deal4")
		dealId5 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+1, */, startEpoch, "deal5")

		// provider1 activates deal 1 and deal2 but that does not activate deal3 to deal5
		ret := actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId1, dealId2)
		actor.assertDealsNotActivated(rt, currentEpoch, dealId3, dealId4, dealId5)
		resp := ret.(*market.ActivateDealsReturn)
		require.Len(t, resp.DealWins, 2)
		require.True(t, resp.DealWins[0].Bool)
		require.True(t, resp.DealWins[1].Bool)

		// provider3 activates deal5 but that does not activate deal3 or deal4
		ret = actor.activateDeals(rt /* sectorExpiry, */, provider2, currentEpoch, dealId5)
		actor.assertDealsNotActivated(rt, currentEpoch, dealId3, dealId4)
		resp = ret.(*market.ActivateDealsReturn)
		require.Len(t, resp.DealWins, 1)
		require.True(t, resp.DealWins[0].Bool)

		// provider1 activates deal3
		ret = actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId3)
		actor.assertDealsNotActivated(rt, currentEpoch, dealId4)
		resp = ret.(*market.ActivateDealsReturn)
		require.Len(t, resp.DealWins, 1)
		require.True(t, resp.DealWins[0].Bool)

		actor.checkState(rt)
	})
}

func TestActivateDealFailures(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}

	startEpoch := abi.ChainEpoch(10)
	endEpoch := startEpoch + 200*builtin.EpochsInDay
	sectorExpiry := endEpoch + 100

	// caller is not the provider
	{
		t.Run("fail when caller is not the provider of the deal", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			provider2 := tutil.NewIDAddr(t, 201)
			expert := tutil.NewIDAddr(t, 202)
			mAddrs2 := &minerAddrs{owner, worker, coinbase, provider2, expert, nil}
			dealId := actor.generateAndPublishDeal(rt, worker, mAddrs2, startEpoch, startEpoch)

			params := mkActivateDealParams(sectorExpiry, dealId)

			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.ExpectAbort(exitcode.ErrForbidden, func() {
				rt.Call(actor.ActivateDeals, params)
			})

			rt.Verify()
			actor.checkState(rt)
		})
	}

	// caller is not a StorageMinerActor
	{
		t.Run("fail when caller is not a StorageMinerActor", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.AccountActorCodeID)
			rt.ExpectAbort(exitcode.SysErrForbidden, func() {
				rt.Call(actor.ActivateDeals, &market.ActivateDealsParams{})
			})

			rt.Verify()
			actor.checkState(rt)
		})
	}

	// deal has not been published before
	{
		t.Run("fail when deal has not been published before", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			params := mkActivateDealParams(sectorExpiry, abi.DealID(42))

			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.ExpectAbort(exitcode.ErrNotFound, func() {
				rt.Call(actor.ActivateDeals, params)
			})

			rt.Verify()
			actor.checkState(rt)
		})
	}

	// deal has ALREADY been activated
	{
		t.Run("fail when deal has already been activated", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
			actor.activateDeals(rt /* sectorExpiry, */, provider, 0, dealId)

			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.ActivateDeals, mkActivateDealParams(sectorExpiry, dealId))
			})

			rt.Verify()
			actor.checkState(rt)
		})
	}

	// deal has invalid params
	{
		t.Run("fail when current epoch greater than start epoch of deal", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
			dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)

			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.SetEpoch(startEpoch + 1)
			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.ActivateDeals, mkActivateDealParams(sectorExpiry, dealId))
			})

			rt.Verify()
			actor.checkState(rt)
		})

		/* t.Run("fail when end epoch of deal greater than sector expiry", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client)
			dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch, endEpoch, startEpoch)

			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.ActivateDeals, mkActivateDealParams(endEpoch-1, dealId))
			})

			rt.Verify()
			actor.checkState(rt)
		}) */
	}

	// all fail if one fails
	{
		t.Run("fail to activate all deals if one deal fails", func(t *testing.T) {
			rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

			// activate deal1 so it fails later
			dealId1 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch, "deal1")
			actor.activateDeals(rt /* sectorExpiry, */, provider, 0, dealId1)

			dealId2 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+1, */, startEpoch, "deal2")

			rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
			rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.ActivateDeals, mkActivateDealParams(sectorExpiry, dealId1, dealId2))
			})
			rt.Verify()

			// no state for deal2 means deal2 activation has failed
			var st market.State
			rt.GetState(&st)

			states, err := market.AsDealStateArray(adt.AsStore(rt), st.States)
			require.NoError(t, err)

			_, found, err := states.Get(dealId2)
			require.NoError(t, err)
			require.False(t, found)

			actor.checkState(rt)

		})
	}

}

func TestOnMinerSectorsTerminate(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAdr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expertAdr, nil}

	startEpoch := abi.ChainEpoch(10)
	// endEpoch := startEpoch + 200*builtin.EpochsInDay
	currentEpoch := abi.ChainEpoch(5)
	/* sectorExpiry := endEpoch + 100 */

	t.Run("terminate multiple deals from multiple providers", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		// provider1 publishes deal1,2 and 3
		dealId1 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch, "deal1")
		dealId2 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+1, */, startEpoch, "deal2")
		dealId3 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+2, */, startEpoch, "deal3")
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId1, dealId2, dealId3)

		// provider2 publishes deal4 and deal5
		provider2 := tutil.NewIDAddr(t, 501)
		expertAddr2 := tutil.NewIDAddr(t, 502)
		maddrs2 := &minerAddrs{owner, worker, coinbase, provider2, expertAddr2, nil}
		dealId4 := actor.generateAndPublishRandDealWithLabel(rt, client, maddrs2, startEpoch, startEpoch, "deal4")
		dealId5 := actor.generateAndPublishRandDealWithLabel(rt, client, maddrs2, startEpoch, startEpoch, "deal5")
		actor.activateDeals(rt, provider2, currentEpoch, dealId4, dealId5)

		// provider1 terminates deal1 but that does not terminate deals2-5
		actor.terminateDeals(rt, provider, dealId1)
		actor.assertDealsTerminated(rt, currentEpoch, dealId1)
		actor.assertDeaslNotTerminated(rt, dealId2, dealId3, dealId4, dealId5)

		// provider2 terminates deal5 but that does not terminate delals 2-4
		actor.terminateDeals(rt, provider2, dealId5)
		actor.assertDealsTerminated(rt, currentEpoch, dealId5)
		actor.assertDeaslNotTerminated(rt, dealId2, dealId3, dealId4)

		// provider1 terminates deal2 and deal3
		actor.terminateDeals(rt, provider, dealId2, dealId3)
		actor.assertDealsTerminated(rt, currentEpoch, dealId2, dealId3)
		actor.assertDeaslNotTerminated(rt, dealId4)

		// provider2 terminates deal4
		actor.terminateDeals(rt, provider2, dealId4)
		actor.assertDealsTerminated(rt, currentEpoch, dealId4)

		ss := actor.checkState(rt)
		require.True(t, ss.DealOpCount == 10)
		require.True(t, len(ss.DealOpEpochStats) == 6) // different dealID to different termination epoch
		require.True(t, len(ss.DealOpEpochStats[startEpoch]) == 5)
	})

	t.Run("ignore deal proposal that does not exist", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		// deal1 will be terminated and the other deal will be ignored because it does not exist
		dealId1 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId1)

		actor.terminateDeals(rt, provider, dealId1, abi.DealID(42))
		st := actor.getDealState(rt, dealId1)
		require.EqualValues(t, currentEpoch, st.SlashEpoch)

		actor.checkState(rt)
	})

	/* t.Run("terminate valid deals along with expired deals - only valid deals are terminated", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		rt.SetEpoch(currentEpoch)

		// provider1 publishes deal1 and 2 and deal3 -> deal3 has the lowest endepoch
		dealId1 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch, endEpoch, startEpoch)
		dealId2 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch, endEpoch+1, startEpoch)
		dealId3 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch, endEpoch-1, startEpoch)
		actor.activateDeals(rt, sectorExpiry, provider, currentEpoch, dealId1, dealId2, dealId3)

		// set current epoch such that deal3 expires but the other two do not
		newEpoch := endEpoch - 1
		rt.SetEpoch(newEpoch)

		// terminating all three deals ONLY terminates deal1 and deal2 because deal3 has expired
		actor.terminateDeals(rt, provider, dealId1, dealId2, dealId3)
		actor.assertDealsTerminated(rt, newEpoch, dealId1, dealId2)
		actor.assertDeaslNotTerminated(rt, dealId3)

		actor.checkState(rt)
	}) */

	t.Run("terminating a deal the second time does not change it's slash epoch", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		dealId1 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId1)

		// terminating the deal so slash epoch is the current epoch
		actor.terminateDeals(rt, provider, dealId1)

		// set a new epoch and terminate again -> however slash epoch will still be the old epoch.
		newEpoch := currentEpoch + 1
		rt.SetEpoch(newEpoch)
		actor.terminateDeals(rt, provider, dealId1)
		st := actor.getDealState(rt, dealId1)
		require.EqualValues(t, currentEpoch, st.SlashEpoch)

		actor.checkState(rt)
	})

	t.Run("terminating new deals and an already terminated deal only terminates the new deals", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		// provider1 publishes deal1 and 2 and deal3 -> deal3 has the lowest endepoch
		dealId1 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch, "deal1")
		dealId2 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+1, */, startEpoch, "deal2")
		dealId3 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch-1, */, startEpoch, "deal3")
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId1, dealId2, dealId3)

		ss := actor.checkState(rt)
		require.True(t, len(ss.DealOpEpochStats[startEpoch]) == 3 && len(ss.DealOpEpochStats) == 1)

		// terminating the deal so slash epoch is the current epoch
		actor.terminateDeals(rt, provider, dealId1)
		ss = actor.checkState(rt)
		require.True(t, len(ss.DealOpEpochStats) == 2)
		_, ok := ss.DealOpEpochStats[currentEpoch+market.DealTerminateLatency+abi.ChainEpoch(dealId1)]
		require.True(t, ok)

		// set a new epoch and terminate again -> however slash epoch will still be the old epoch.
		newEpoch := currentEpoch + 1
		rt.SetEpoch(newEpoch)
		actor.terminateDeals(rt, provider, dealId1, dealId2, dealId3)
		ss = actor.checkState(rt)
		require.True(t, len(ss.DealOpEpochStats) == 4)
		_, ok = ss.DealOpEpochStats[newEpoch+market.DealTerminateLatency+abi.ChainEpoch(dealId2)]
		require.True(t, ok)
		_, ok = ss.DealOpEpochStats[newEpoch+market.DealTerminateLatency+abi.ChainEpoch(dealId3)]
		require.True(t, ok)

		st := actor.getDealState(rt, dealId1)
		require.EqualValues(t, currentEpoch, st.SlashEpoch)

		st2 := actor.getDealState(rt, dealId2)
		require.EqualValues(t, newEpoch, st2.SlashEpoch)

		st3 := actor.getDealState(rt, dealId3)
		require.EqualValues(t, newEpoch, st3.SlashEpoch)

		actor.checkState(rt)
	})

	/* t.Run("do not terminate deal if end epoch is equal to or less than current epoch", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		rt.SetEpoch(currentEpoch)

		// deal1 has endepoch equal to current epoch when terminate is called
		dealId1 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch, endEpoch, startEpoch)
		actor.activateDeals(rt, sectorExpiry, provider, currentEpoch, dealId1)
		rt.SetEpoch(endEpoch)
		actor.terminateDeals(rt, provider, dealId1)
		actor.assertDeaslNotTerminated(rt, dealId1)

		// deal2 has end epoch less than current epoch when terminate is called
		rt.SetEpoch(currentEpoch)
		dealId2 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch+1, endEpoch, startEpoch+1)
		actor.activateDeals(rt, sectorExpiry, provider, currentEpoch, dealId2)
		rt.SetEpoch(endEpoch + 1)
		actor.terminateDeals(rt, provider, dealId2)
		actor.assertDeaslNotTerminated(rt, dealId2)

		actor.checkState(rt)
	}) */

	t.Run("fail when caller is not a StorageMinerActor", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.SetCaller(provider, builtin.AccountActorCodeID)
		rt.ExpectAbort(exitcode.SysErrForbidden, func() {
			rt.Call(actor.OnMinerSectorsTerminate, &market.OnMinerSectorsTerminateParams{})
		})

		rt.Verify()
		actor.checkState(rt)

	})

	t.Run("fail when caller is not the provider of the deal", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId)

		params := mkTerminateDealParams(currentEpoch, dealId)

		provider2 := tutil.NewIDAddr(t, 501)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.SetCaller(provider2, builtin.StorageMinerActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalState, "caller t0501 is not the provider t0102 of deal 0", func() {
			rt.Call(actor.OnMinerSectorsTerminate, params)
		})

		rt.Verify()
		actor.checkState(rt)

	})

	t.Run("fail when termination occurs at a future epoch", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId)

		params := mkTerminateDealParams(currentEpoch+1, dealId)

		provider2 := tutil.NewIDAddr(t, 501)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.SetCaller(provider2, builtin.StorageMinerActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalState, "future termination", func() {
			rt.Call(actor.OnMinerSectorsTerminate, params)
		})

		rt.Verify()
		actor.checkState(rt)

	})

	t.Run("fail when deal has been published but not activated", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)

		params := mkTerminateDealParams(currentEpoch, dealId)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.OnMinerSectorsTerminate, params)
		})

		rt.Verify()
		actor.checkState(rt)

	})

	t.Run("termination of all deals should fail when one deal fails", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.SetEpoch(currentEpoch)

		// deal1 would terminate but deal2 will fail because deal2 has not been activated
		dealId1 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch, "deal1")
		actor.activateDeals(rt /* sectorExpiry, */, provider, currentEpoch, dealId1)
		dealId2 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, startEpoch /* endEpoch+1, */, startEpoch, "deal2")

		params := mkTerminateDealParams(currentEpoch, dealId1, dealId2)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.OnMinerSectorsTerminate, params)
		})

		rt.Verify()

		// verify deal1 has not been terminated
		actor.assertDeaslNotTerminated(rt, dealId1)

		actor.checkState(rt)
	})
}

func TestCronTick(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAddr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expertAddr, nil}

	startEpoch := abi.ChainEpoch(50)
	// endEpoch := startEpoch + 200*builtin.EpochsInDay

	/* t.Run("fail when deal is activated but proposal is not found", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch,  0,  startEpoch)

		// delete the deal proposal (this breaks state invariants)
		actor.deleteDealProposal(rt, dealId)

		// move the current epoch to the start epoch of the deal
		rt.SetEpoch(startEpoch)
		rt.ExpectAbort(exitcode.ErrNotFound, func() {
			actor.cronTick(rt)
		})
	}) */

	t.Run("fail when deal update epoch is in the future", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, 0, startEpoch)

		// update last updated to some time in the future
		actor.updateLastUpdated(rt, dealId, startEpoch+1000)

		// move the current epoch such that the deal's last updated field is set to the start epoch of the deal
		// and the next tick for it is scheduled at the endepoch.
		rt.SetEpoch(startEpoch)

		rt.ExpectAbort(exitcode.ErrIllegalState, func() {
			actor.cronTick(rt)
		})
	})

	t.Run("crontick for a deal at it's start epoch results in zero payment and no slashing", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, 0, startEpoch)

		ss := actor.checkState(rt)
		require.True(t, ss.DealOpCount == 1 && len(ss.DealOpEpochStats) == 1)

		// move the current epoch to startEpoch
		current := startEpoch
		rt.SetEpoch(current)
		pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		require.EqualValues(t, big.Zero(), pay)
		require.EqualValues(t, big.Zero(), slashed)

		// deal proposal and state should NOT be deleted
		require.NotNil(t, actor.getDealProposal(rt, dealId))
		require.NotNil(t, actor.getDealState(rt, dealId))

		ss = actor.checkState(rt)
		require.True(t, ss.DealOpCount == 0 && len(ss.DealOpEpochStats) == 0)
	})

	t.Run("slash a deal and make payment for another deal in the same epoch", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

		dealId1 := actor.publishAndActivateRandDeal(rt, client, mAddrs, startEpoch, 0, startEpoch)
		d1 := actor.getDealProposal(rt, dealId1)

		dealId2 := actor.publishAndActivateRandDeal(rt, client, mAddrs, startEpoch+1, 0, startEpoch+1)

		// slash deal1
		slashEpoch := abi.ChainEpoch(150)
		terminateEpoch := slashEpoch + market.DealTerminateLatency + abi.ChainEpoch(dealId1)
		rt.SetEpoch(slashEpoch)
		actor.terminateDeals(rt, provider, dealId1)

		// cron tick will slash deal1 and make payment for deal2
		current := abi.ChainEpoch(151)
		rt.SetEpoch(current)
		/* rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d1.ProviderCollateral, nil, exitcode.Ok) */
		actor.cronTick(rt)

		actor.assertDealDeleted(rt, dealId1, d1)
		s2 := actor.getDealState(rt, dealId2)
		require.EqualValues(t, current, s2.LastUpdatedEpoch)

		ss := actor.checkState(rt)
		require.True(t, ss.DealNotFound == 1 && ss.DealOpCount == 1) // deal terminated but cron exists

		// terminate tick will do nothing cause deal already deleted
		rt.SetEpoch(terminateEpoch)
		actor.cronTick(rt)
		ss = actor.checkState(rt)
		require.True(t, ss.DealNotFound == 0 && ss.DealOpCount == 0)
	})

	t.Run("cannot publish the same deal twice BEFORE a cron tick", func(t *testing.T) {
		// Publish a deal
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId1 := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		d1 := actor.getDealProposal(rt, dealId1)

		// now try to publish it again and it should fail because it will still be in pending state
		d2 := actor.generateDealAndAddFunds(rt, client, mAddrs, startEpoch /* , endEpoch */)
		params := mkPublishStorageParams(d2)
		params.DataRef.Expert = expertAddr.String()
		rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
		rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)
		batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: d2.PieceCID}}}
		rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
		rt.ExpectSend(d2.Provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)
		/* expectQueryNetworkInfo(rt, actor) */
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectVerifySignature(crypto.Signature{}, d2.Client, mustCbor(&d2), nil)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.PublishStorageDeals, params)
		})
		rt.Verify()

		// now a cron tick happens -> deal1 is no longer pending and then publishing the same deal again should work
		rt.SetEpoch(d1.StartEpoch - 1)
		actor.activateDeals(rt, provider, d1.StartEpoch-1, dealId1)
		rt.SetEpoch(d1.StartEpoch)
		actor.cronTick(rt)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, mAddrs, publishDealReq{deal: d2})

		actor.checkState(rt)
	})
}

func TestRandomCronEpochDuringPublish(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}

	startEpoch := abi.ChainEpoch(50)
	epochUndefined := abi.ChainEpoch(-1)
	// endEpoch := startEpoch + 200*builtin.EpochsInDay
	// sectorExpiry := endEpoch + 1

	t.Run("a random epoch in chosen as the cron processing epoch for a deal during publishing", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		processEpoch := startEpoch + 5
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, processEpoch)
		d := actor.getDealProposal(rt, dealId)

		// activate the deal
		rt.SetEpoch(startEpoch - 1)
		actor.activateDeals(rt /* sectorExpiry, */, provider, d.StartEpoch-1, dealId)

		// cron tick at deal start epoch does not do anything
		rt.SetEpoch(startEpoch)
		actor.cronTickNoChange(rt, client, provider)
		ds := actor.getDealState(rt, dealId)
		require.Equal(t, ds.LastUpdatedEpoch, epochUndefined)

		ss := actor.checkState(rt)
		require.Equal(t, uint64(1), ss.DealOpCount)

		// first cron tick at process epoch will make payment and schedule the deal for next epoch
		rt.SetEpoch(processEpoch)
		pay, _ := actor.cronTickAndAssertBalances(rt, client, provider, processEpoch, dealId)
		/* duration := big.Sub(big.NewInt(int64(processEpoch)), big.NewInt(int64(startEpoch)))
		require.EqualValues(t, big.Mul(duration, d.StoragePricePerEpoch), pay) */
		require.EqualValues(t, big.Zero(), pay)

		/* // payment at next epoch
		current := processEpoch + market.DealUpdatesInterval
		rt.SetEpoch(current)
		pay, _ = actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		duration = big.Sub(big.NewInt(int64(current)), big.NewInt(int64(processEpoch)))
		require.EqualValues(t, big.Mul(duration, d.StoragePricePerEpoch), pay)
		require.EqualValues(t, big.Zero(), pay) */

		ss = actor.checkState(rt)
		require.Zero(t, ss.DealOpCount)
		ds = actor.getDealState(rt, dealId)
		require.Equal(t, ds.LastUpdatedEpoch, processEpoch)
	})

	/* t.Run("deals are scheduled for expiry later than the end epoch", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch, endEpoch, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		rt.SetEpoch(startEpoch - 1)
		actor.activateDeals(rt, sectorExpiry, provider, d.StartEpoch-1, dealId)

		// a cron tick at end epoch -1 schedules the deal for later than end epoch
		curr := endEpoch - 1
		rt.SetEpoch(curr)
		duration := big.NewInt(int64(curr - startEpoch))
		pay, _ := actor.cronTickAndAssertBalances(rt, client, provider, curr, dealId)
		require.EqualValues(t, big.Mul(duration, d.StoragePricePerEpoch), pay)

		// cron tick at end epoch does NOT expire the deal
		rt.SetEpoch(endEpoch)
		actor.cronTickNoChange(rt, client, provider)
		require.NotNil(t, actor.getDealProposal(rt, dealId))

		// cron tick at nextEpoch expires the deal -> payment is ONLY for one epoch
		curr = curr + market.DealUpdatesInterval
		rt.SetEpoch(curr)
		pay, _ = actor.cronTickAndAssertBalances(rt, client, provider, curr, dealId)
		require.EqualValues(t, d.StoragePricePerEpoch, pay)
		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	})

	t.Run("deal is processed after it's end epoch -> should expire correctly", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		processEpoch := endEpoch + 100

		activationEpoch := startEpoch - 1
		rt.SetEpoch(activationEpoch)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, endEpoch, activationEpoch, sectorExpiry, processEpoch)
		d := actor.getDealProposal(rt, dealId)

		rt.SetEpoch(processEpoch)
		pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, processEpoch, dealId)
		require.EqualValues(t, big.Zero(), slashed)
		duration := big.Sub(big.NewInt(int64(endEpoch)), big.NewInt(int64(startEpoch)))
		require.EqualValues(t, big.Mul(duration, d.StoragePricePerEpoch), pay)

		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	}) */

	t.Run("activation after deal start epoch but before it is processed fails", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		processEpoch := startEpoch + 5
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, processEpoch)

		// activate the deal after the start epoch
		rt.SetEpoch(startEpoch + 1)

		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			actor.activateDeals(rt /* sectorExpiry, */, provider, startEpoch+1, dealId)
		})

		actor.checkState(rt)
	})

	t.Run("cron processing of deal after missed activation should fail and slash", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		processEpoch := startEpoch + 5
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, processEpoch)
		d := actor.getDealProposal(rt, dealId)

		rt.SetEpoch(processEpoch)

		// rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d.ProviderCollateral, nil, exitcode.Ok)
		actor.cronTick(rt)

		actor.assertDealDeleted(rt, dealId, d)

		ss := actor.checkState(rt)
		require.Zero(t, ss.DealOpCount)
	})

}

func TestLockedFundTrackingStates(t *testing.T) {
	t.Parallel()
	owner := tutil.NewIDAddr(t, 101)
	worker := tutil.NewIDAddr(t, 103)
	coinbase := tutil.NewIDAddr(t, 106)
	expert := tutil.NewIDAddr(t, 109)

	p1 := tutil.NewIDAddr(t, 201)
	p2 := tutil.NewIDAddr(t, 202)
	/* p3 := tutil.NewIDAddr(t, 203) */

	c1 := tutil.NewIDAddr(t, 104)
	c2 := tutil.NewIDAddr(t, 105)
	/* c3 := tutil.NewIDAddr(t, 106) */

	m1 := &minerAddrs{owner, worker, coinbase, p1, expert, nil}
	m2 := &minerAddrs{owner, worker, coinbase, p2, expert, nil}
	/* m3 := &minerAddrs{owner, worker, p3} */

	startEpoch := abi.ChainEpoch(50)
	endEpoch := startEpoch + 200*builtin.EpochsInDay
	/* sectorExpiry := endEpoch + 400 */

	var st market.State

	// assert values are zero
	rt, actor := basicMarketSetup(t, owner, p1, worker, c1, coinbase)
	rt.GetState(&st)
	/* require.True(t, st.TotalClientLockedCollateral.IsZero())
	require.True(t, st.TotalProviderLockedCollateral.IsZero())
	require.True(t, st.TotalClientStorageFee.IsZero()) */

	// Publish deal1, deal2 and deal3  with different client and provider
	dealId1 := actor.generateAndPublishRandDeal(rt, c1, m1, startEpoch /* endEpoch, */, startEpoch)
	/* d1 := actor.getDealProposal(rt, dealId1) */

	dealId2 := actor.generateAndPublishRandDeal(rt, c2, m2, startEpoch /* endEpoch, */, startEpoch)
	/*d2 := actor.getDealProposal(rt, dealId2)

	dealId3 := actor.generateAndPublishDeal(rt, c3, m3, startEpoch, endEpoch, startEpoch)
	d3 := actor.getDealProposal(rt, dealId3)

	 csf := big.Sum(d1.TotalStorageFee(), d2.TotalStorageFee(), d3.TotalStorageFee())
	plc := big.Sum(d1.ProviderCollateral, d2.ProviderCollateral, d3.ProviderCollateral)
	clc := big.Sum(d1.ClientCollateral, d2.ClientCollateral, d3.ClientCollateral)

	actor.assertLockedFundStates(rt, csf, plc, clc) */

	// activation dosen't change anything
	curr := startEpoch - 1
	rt.SetEpoch(curr)
	actor.activateDeals(rt /* sectorExpiry, */, p1, curr, dealId1)
	actor.activateDeals(rt /* sectorExpiry, */, p2, curr, dealId2)

	/* actor.assertLockedFundStates(rt, csf, plc, clc) */

	// make payment for p1 and p2, p3 times out as it has not been activated
	curr = 51 // startEpoch + 1
	rt.SetEpoch(curr)
	// rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d3.ProviderCollateral, nil, exitcode.Ok)
	actor.cronTick(rt)
	/* payment := big.Product(big.NewInt(2), d1.StoragePricePerEpoch)
	csf = big.Sub(big.Sub(csf, payment), d3.TotalStorageFee())
	plc = big.Sub(plc, d3.ProviderCollateral)
	clc = big.Sub(clc, d3.ClientCollateral)
	actor.assertLockedFundStates(rt, csf, plc, clc) */

	// deal1 and deal2 will now be charged at epoch curr + market.DealUpdatesInterval, so nothing changes before that.
	rt.SetEpoch(curr + market.DealUpdatesInterval - 1)
	actor.cronTick(rt)
	/* actor.assertLockedFundStates(rt, csf, plc, clc) */

	// one more round of payment for deal1 and deal2
	curr2 := curr + market.DealUpdatesInterval
	rt.SetEpoch(curr2)
	/* duration := big.NewInt(int64(curr2 - curr))
	payment = big.Product(big.NewInt(2), d1.StoragePricePerEpoch, duration)
	csf = big.Sub(csf, payment) */
	actor.cronTick(rt)
	/* actor.assertLockedFundStates(rt, csf, plc, clc) */

	// slash deal1 at 201
	slashEpoch := curr2 + 1
	rt.SetEpoch(slashEpoch)
	actor.terminateDeals(rt, m1.provider, dealId1)

	// cron tick to slash deal1 and expire deal2
	rt.SetEpoch(endEpoch)
	/* csf = big.Zero()
	clc = big.Zero()
	plc = big.Zero()
	rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d1.ProviderCollateral, nil, exitcode.Ok) */
	actor.cronTick(rt)
	/* actor.assertLockedFundStates(rt, csf, plc, clc) */

	actor.checkState(rt)
}

func TestCronTickTimedoutDeals(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAddr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expertAddr, nil}

	startEpoch := abi.ChainEpoch(50)
	/* endEpoch := startEpoch + 200*builtin.EpochsInDay */

	t.Run("timed out deal is slashed and deleted", func(t *testing.T) {
		// publish a deal but do NOT activate it
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		cEscrow := actor.getEscrowBalance(rt, client)

		// do a cron tick for it -> should time out and get slashed
		rt.SetEpoch(startEpoch)
		// rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d.ProviderCollateral, nil, exitcode.Ok)
		actor.cronTick(rt)

		require.Equal(t, cEscrow, actor.getEscrowBalance(rt, client))
		require.Equal(t, big.Zero(), actor.getLockedBalance(rt, client))

		actor.assertAccountBalance(rt, provider, big.NewInt(10), big.Zero())

		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	})

	t.Run("publishing timed out deal again should work after cron tick as it should no longer be pending", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		// publishing will fail as it will be in pending
		d2 := actor.generateDealAndAddFunds(rt, client, mAddrs, startEpoch /* , endEpoch */)
		params := mkPublishStorageParams(d2)
		params.DataRef.Expert = expertAddr.String()
		rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
		rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)
		batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: d2.PieceCID}}}
		rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
		rt.ExpectSend(d2.Provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)

		/* expectQueryNetworkInfo(rt, actor) */
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectVerifySignature(crypto.Signature{}, d2.Client, mustCbor(&d2), nil)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.PublishStorageDeals, params)
		})
		rt.Verify()

		// do a cron tick for it -> should time out and get slashed
		rt.SetEpoch(startEpoch)
		// rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d.ProviderCollateral, nil, exitcode.Ok)
		actor.cronTick(rt)
		actor.assertDealDeleted(rt, dealId, d)

		// now publishing should work
		actor.generateAndPublishDeal(rt, client, mAddrs, startEpoch /* endEpoch, */, startEpoch)

		actor.checkState(rt)
	})

	t.Run("timed out and verified deals are slashed, deleted AND sent to the Registry actor", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		// deal1 and deal2 are verified
		deal1 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, startEpoch /* , endEpoch */)
		deal1.Label = "deal1"
		// deal1.VerifiedDeal = true
		deal2 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, startEpoch /* , endEpoch+1 */)
		deal2.Label = "deal2"
		// deal2.VerifiedDeal = true

		// deal3 is NOT verified
		deal3 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, startEpoch /* , endEpoch+2 */)
		deal3.Label = "deal3"

		//  publishing verified deals
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		dealIds := actor.publishDeals(rt, mAddrs, publishDealReq{deal1, startEpoch},
			publishDealReq{deal2, startEpoch}, publishDealReq{deal3, startEpoch})

		// do a cron tick for it -> all should time out and get slashed
		// ONLY deal1 and deal2 should be sent to the Registry actor
		rt.SetEpoch(startEpoch)

		/* // expected sends to the registry actor
		param1 := &verifreg.RestoreBytesParams{
			Address:  deal1.Client,
			DealSize: big.NewIntUnsigned(uint64(deal1.PieceSize)),
		}
		param2 := &verifreg.RestoreBytesParams{
			Address:  deal2.Client,
			DealSize: big.NewIntUnsigned(uint64(deal2.PieceSize)),
		}

		rt.ExpectSend(builtin.VerifiedRegistryActorAddr, builtin.MethodsVerifiedRegistry.RestoreBytes, param1,
			abi.NewTokenAmount(0), nil, exitcode.Ok)
		rt.ExpectSend(builtin.VerifiedRegistryActorAddr, builtin.MethodsVerifiedRegistry.RestoreBytes, param2,
			abi.NewTokenAmount(0), nil, exitcode.Ok)

		expectedBurn := big.Mul(big.NewInt(3), deal1.ProviderCollateral)
		rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, expectedBurn, nil, exitcode.Ok) */
		actor.cronTick(rt)

		// a second cron tick for the same epoch should not change anything
		actor.cronTickNoChange(rt, client, provider)

		actor.assertAccountBalance(rt, provider, big.NewInt(10*3), big.Zero())
		actor.assertDealDeleted(rt, dealIds[0], &deal1)
		actor.assertDealDeleted(rt, dealIds[1], &deal2)
		actor.assertDealDeleted(rt, dealIds[2], &deal3)

		actor.checkState(rt)
	})
}

func TestCronTickDealSlashing(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}

	// hairy edge cases
	{

		tcs := map[string]struct {
			dealStart abi.ChainEpoch
			// dealEnd          abi.ChainEpoch
			activationEpoch  abi.ChainEpoch
			terminationEpoch abi.ChainEpoch
			cronTickEpoch    abi.ChainEpoch
			// payment          abi.TokenAmount
		}{
			"deal is slashed after the startepoch and then the first crontick happens": {
				dealStart: abi.ChainEpoch(10),
				// dealEnd:          abi.ChainEpoch(10 + 200*builtin.EpochsInDay),
				activationEpoch:  abi.ChainEpoch(5),
				terminationEpoch: abi.ChainEpoch(15),
				cronTickEpoch:    abi.ChainEpoch(16),
				// payment:          abi.NewTokenAmount(50), // (15 - 10) * 10 as deal storage fee is 10 per epoch
			},
			"deal is slashed at the startepoch and then the first crontick happens": {
				dealStart: abi.ChainEpoch(10),
				// dealEnd:          abi.ChainEpoch(10 + 200*builtin.EpochsInDay),
				activationEpoch:  abi.ChainEpoch(5),
				terminationEpoch: abi.ChainEpoch(10),
				cronTickEpoch:    abi.ChainEpoch(11),
				// payment:          abi.NewTokenAmount(0), // (10 - 10) * 10
			},
			"deal is slashed before the startepoch and then the first crontick happens": {
				dealStart: abi.ChainEpoch(10),
				// dealEnd:          abi.ChainEpoch(10 + 200*builtin.EpochsInDay),
				activationEpoch:  abi.ChainEpoch(5),
				terminationEpoch: abi.ChainEpoch(6),
				cronTickEpoch:    abi.ChainEpoch(10),
				// payment:          abi.NewTokenAmount(0), // (10 - 10) * 10
			},
			"deal is terminated at the activation epoch and then the first crontick happens": {
				dealStart: abi.ChainEpoch(10),
				// dealEnd:          abi.ChainEpoch(10 + 200*builtin.EpochsInDay),
				activationEpoch:  abi.ChainEpoch(5),
				terminationEpoch: abi.ChainEpoch(5),
				cronTickEpoch:    abi.ChainEpoch(10),
				// payment:          abi.NewTokenAmount(0), // (10 - 10) * 10
			},
			"deal is slashed and then deal expiry happens on crontick, but slashing still occurs": {
				dealStart: abi.ChainEpoch(10),
				// dealEnd:          abi.ChainEpoch(10 + 200*builtin.EpochsInDay),
				activationEpoch:  abi.ChainEpoch(5),
				terminationEpoch: abi.ChainEpoch(15),
				cronTickEpoch:    abi.ChainEpoch(25), // deal has expired
				// payment:          abi.NewTokenAmount(50),
			},
			"deal is slashed just BEFORE the end epoch": {
				dealStart: abi.ChainEpoch(10),
				// dealEnd:          abi.ChainEpoch(10 + 200*builtin.EpochsInDay),
				activationEpoch:  abi.ChainEpoch(5),
				terminationEpoch: abi.ChainEpoch(19),
				cronTickEpoch:    abi.ChainEpoch(19),
				// payment:          abi.NewTokenAmount(90), // (19 - 10) * 10
			},
		}

		for n, tc := range tcs {
			t.Run(n, func(t *testing.T) {
				rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

				// publish and activate
				rt.SetEpoch(tc.activationEpoch)
				dealId := actor.publishAndActivateDeal(rt, client, mAddrs, tc.dealStart, tc.activationEpoch, tc.dealStart)
				d := actor.getDealProposal(rt, dealId)

				// terminate
				rt.SetEpoch(tc.terminationEpoch)
				actor.terminateDeals(rt, provider, dealId)

				//  cron tick
				rt.SetEpoch(tc.cronTickEpoch)

				pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, tc.cronTickEpoch, dealId)
				require.EqualValues(t, big.Zero(), pay)
				// require.EqualValues(t, d.ProviderCollateral, slashed)
				require.EqualValues(t, big.Zero(), slashed)
				actor.assertDealDeleted(rt, dealId, d)

				// // if there has been no payment, provider will have zero balance and hence should be slashed
				// if tc.payment.Equals(big.Zero()) {
				actor.assertAccountBalance(rt, provider, big.NewInt(10), big.Zero())
				// client balances should not change
				cLocked := actor.getLockedBalance(rt, client)
				cEscrow := actor.getEscrowBalance(rt, client)
				actor.cronTick(rt)
				require.EqualValues(t, cEscrow, actor.getEscrowBalance(rt, client))
				require.EqualValues(t, cLocked, actor.getLockedBalance(rt, client))
				// } else {
				// 	// running cron tick again dosen't do anything
				// 	actor.cronTickNoChange(rt, client, provider)
				// }
				actor.checkState(rt)
			})
		}
	}

	startEpoch := abi.ChainEpoch(50)
	endEpoch := abi.ChainEpoch(50 + 10*builtin.EpochsInDay)

	t.Run("deal is slashed AT the end epoch -> should NOT be slashed and should be considered expired", func(t *testing.T) {
		t.Parallel()
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, 0, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		// set current epoch to deal end epoch and attempt to slash it -> should not be slashed
		// as deal is considered to be expired.
		current := endEpoch
		rt.SetEpoch(current)
		actor.terminateDeals(rt, provider, dealId)

		// on the next cron tick, it will be processed as expired
		current = endEpoch + 300
		rt.SetEpoch(current)
		pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		/* duration := big.NewInt(int64(endEpoch - startEpoch)) // end - start
		require.EqualValues(t, big.Mul(duration, d.StoragePricePerEpoch), pay) */
		require.EqualValues(t, big.Zero(), pay)
		require.EqualValues(t, big.Zero(), slashed)

		// deal should be deleted as it should have expired
		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	})

	t.Run("deal is correctly processed twice in the same crontick and slashed", func(t *testing.T) {
		t.Parallel()
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, 0, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		// move the current epoch to startEpoch so next cron epoch will be start + Interval
		current := startEpoch
		rt.SetEpoch(current)
		pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		require.EqualValues(t, big.Zero(), pay)
		require.EqualValues(t, big.Zero(), slashed)

		// set slash epoch of deal
		slashEpoch := current + 10
		rt.SetEpoch(slashEpoch)
		actor.terminateDeals(rt, provider, dealId)

		current2 := current + 10 + market.DealTerminateLatency + abi.ChainEpoch(dealId) - 1
		rt.SetEpoch(current2)
		actor.cronTickAndAssertBalances(rt, client, provider, current2, dealId)
		actor.getDealState(rt, dealId)

		// deal should be deleted as it should have expired
		current3 := current2 + 1
		rt.SetEpoch(current3)
		actor.cronTickAndAssertBalances(rt, client, provider, current3, dealId)
		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	})

	// end-end tests for slashing
	t.Run("slash multiple deals in the same epoch", func(t *testing.T) {
		t.Parallel()
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

		// three deals for slashing
		dealId1 := actor.publishAndActivateRandDealWithLabel(rt, client, mAddrs, startEpoch, 0, startEpoch, "deal1")
		d1 := actor.getDealProposal(rt, dealId1)

		dealId2 := actor.publishAndActivateRandDealWithLabel(rt, client, mAddrs, startEpoch, 0, startEpoch, "deal2")
		d2 := actor.getDealProposal(rt, dealId2)

		dealId3 := actor.publishAndActivateRandDealWithLabel(rt, client, mAddrs, startEpoch, 0, startEpoch, "deal3")
		d3 := actor.getDealProposal(rt, dealId3)

		// set slash epoch of deal at 151
		current := abi.ChainEpoch(151)
		rt.SetEpoch(current)
		actor.terminateDeals(rt, provider, dealId1, dealId2, dealId3)

		// process slashing of deals
		current = 300
		rt.SetEpoch(current)
		/* totalSlashed := big.Sum(d1.ProviderCollateral, d2.ProviderCollateral, d3.ProviderCollateral)
		rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, totalSlashed, nil, exitcode.Ok) */

		actor.cronTick(rt)

		actor.assertDealDeleted(rt, dealId1, d1)
		actor.assertDealDeleted(rt, dealId2, d2)
		actor.assertDealDeleted(rt, dealId3, d3)

		actor.checkState(rt)
	})

	/* t.Run("regular payments till deal is slashed and then slashing is processed", func(t *testing.T) {
		t.Parallel()
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, endEpoch, 0, sectorExpiry, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		// move the current epoch to startEpoch + 5 so payment is made
		current := startEpoch + 5
		rt.SetEpoch(current)
		// assert payment
		pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		require.EqualValues(t, pay, big.Mul(big.NewInt(5), d.StoragePricePerEpoch))
		require.EqualValues(t, big.Zero(), slashed)

		// Setting the current epoch to before the next schedule will NOT make any changes as the deal
		// is still not scheduled
		current2 := current + market.DealUpdatesInterval - 1
		rt.SetEpoch(current2)
		actor.cronTickNoChange(rt, client, provider)

		// a second cron tick for the same epoch should not change anything
		actor.cronTickNoChange(rt, client, provider)

		//  make another payment
		current3 := current2 + 1
		rt.SetEpoch(current3)
		duration := big.NewInt(int64(current3 - current))
		pay, slashed = actor.cronTickAndAssertBalances(rt, client, provider, current3, dealId)
		require.EqualValues(t, pay, big.Mul(duration, d.StoragePricePerEpoch))
		require.EqualValues(t, big.Zero(), slashed)

		// a second cron tick for the same epoch should not change anything
		actor.cronTickNoChange(rt, client, provider)

		// now terminate the deal
		slashEpoch := current3 + 1
		rt.SetEpoch(slashEpoch)
		actor.terminateDeals(rt, provider, dealId)

		// Setting the epoch to anything less than next schedule will not make any change even though the deal is slashed
		current4 := current3 + market.DealUpdatesInterval - 1
		rt.SetEpoch(current4)
		actor.cronTickNoChange(rt, client, provider)

		// next epoch for cron schedule  -> payment will be made and deal will be slashed
		current5 := current4 + 1
		rt.SetEpoch(current5)
		duration = big.NewInt(int64(slashEpoch - current3))
		pay, slashed = actor.cronTickAndAssertBalances(rt, client, provider, current5, dealId)
		require.EqualValues(t, pay, big.Mul(duration, d.StoragePricePerEpoch))
		require.EqualValues(t, d.ProviderCollateral, slashed)

		// deal should be deleted as it should have expired
		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	})

	// expired deals should NOT be slashed
	t.Run("regular payments till deal expires and then we attempt to slash it but it will NOT be slashed", func(t *testing.T) {
		t.Parallel()
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		dealId := actor.publishAndActivateDeal(rt, client, mAddrs, startEpoch, endEpoch, 0, sectorExpiry, startEpoch)
		d := actor.getDealProposal(rt, dealId)

		// move the current epoch to startEpoch + 5 so payment is made and assert payment
		current := startEpoch + 5 // 55
		rt.SetEpoch(current)
		pay, slashed := actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		require.EqualValues(t, pay, big.Mul(big.NewInt(5), d.StoragePricePerEpoch))
		require.EqualValues(t, big.Zero(), slashed)

		//  Setting the current epoch to 155 will make another payment
		current2 := current + market.DealUpdatesInterval
		rt.SetEpoch(current2)
		duration := big.NewInt(int64(current2 - current))
		pay, slashed = actor.cronTickAndAssertBalances(rt, client, provider, current2, dealId)
		require.EqualValues(t, pay, big.Mul(duration, d.StoragePricePerEpoch))
		require.EqualValues(t, big.Zero(), slashed)

		// set current epoch to deal end epoch and attempt to slash it -> should not be slashed
		// as deal is considered to be expired.
		rt.SetEpoch(endEpoch)
		actor.terminateDeals(rt, provider, dealId)

		// next epoch for cron schedule is endEpoch + 300 ->
		// setting epoch to higher than that will cause deal to be expired, payment will be made
		// and deal will NOT be slashed
		current = endEpoch + 300
		rt.SetEpoch(current)
		pay, slashed = actor.cronTickAndAssertBalances(rt, client, provider, current, dealId)
		duration = big.NewInt(int64(endEpoch - current2))
		require.EqualValues(t, big.Mul(duration, d.StoragePricePerEpoch), pay)
		require.EqualValues(t, big.Zero(), slashed)

		// deal should be deleted as it should have expired
		actor.assertDealDeleted(rt, dealId, d)

		actor.checkState(rt)
	}) */
}

func TestMarketActorDeals(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAddr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	minerAddrs := &minerAddrs{owner, worker, coinbase, provider, expertAddr, nil}

	var st market.State

	// Test adding provider funds from both worker and owner address
	rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
	actor.addProviderFunds(rt, abi.NewTokenAmount(20000000), minerAddrs)
	rt.GetState(&st)
	assert.Equal(t, abi.NewTokenAmount(20000000), actor.getEscrowBalance(rt, provider))

	actor.addParticipantFunds(rt, client, abi.NewTokenAmount(20000000))

	dealProposal := generateDealProposal(client, provider, abi.ChainEpoch(1) /* , abi.ChainEpoch(200*builtin.EpochsInDay) */)
	params := &market.PublishStorageDealsParams{Deals: []market.ClientDealProposal{{Proposal: dealProposal}}, DataRef: market.PublishStorageDataRef{Expert: expertAddr.String()}}

	// First attempt at publishing the deal should work
	{
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, minerAddrs, publishDealReq{deal: dealProposal})
	}

	// Second attempt at publishing the same deal should fail
	{
		rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
		rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)
		batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: dealProposal.PieceCID}}}
		rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
		rt.ExpectSend(provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)

		/* expectQueryNetworkInfo(rt, actor) */
		rt.ExpectVerifySignature(crypto.Signature{}, client, mustCbor(&params.Deals[0].Proposal), nil)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "file already published", func() {
			rt.Call(actor.PublishStorageDeals, params)
		})

		rt.Verify()
	}

	dealProposal.Label = "foo"
	// Same deal with a different label should fail
	{
		rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
		rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)
		batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: dealProposal.PieceCID}}}
		rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
		rt.ExpectSend(provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)

		rt.ExpectVerifySignature(crypto.Signature{}, client, mustCbor(&params.Deals[0].Proposal), nil)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "file already published", func() {
			rt.Call(actor.PublishStorageDeals, params)
		})

		rt.Verify()
	}
	actor.checkState(rt)
}

func TestMaxDealLabelSize(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expertAddr := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	minerAddrs := &minerAddrs{owner, worker, coinbase, provider, expertAddr, nil}

	var st market.State

	// Test adding provider funds from both worker and owner address
	rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
	actor.addProviderFunds(rt, abi.NewTokenAmount(20000000), minerAddrs)
	rt.GetState(&st)
	assert.Equal(t, abi.NewTokenAmount(20000000), actor.getEscrowBalance(rt, provider))

	actor.addParticipantFunds(rt, client, abi.NewTokenAmount(20000000))

	dealProposal := generateDealProposal(client, provider, abi.ChainEpoch(1) /* , abi.ChainEpoch(200*builtin.EpochsInDay) */)
	dealProposal.Label = string(make([]byte, market.DealMaxLabelSize))
	params := &market.PublishStorageDealsParams{Deals: []market.ClientDealProposal{{Proposal: dealProposal}}, DataRef: market.PublishStorageDataRef{Expert: expertAddr.String()}}

	// Label at max size should work.
	{
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		actor.publishDeals(rt, minerAddrs, publishDealReq{deal: dealProposal})
	}

	params.Deals[0].Proposal.Label = string(make([]byte, market.DealMaxLabelSize+1))
	// Label greater than max size should fail.
	{
		rt.ExpectValidateCallerType(builtin.AccountActorCodeID, builtin.MultisigActorCodeID)
		rt.ExpectSend(provider, builtin.MethodsMiner.ControlAddresses, nil, abi.NewTokenAmount(0), &builtin.GetControlAddressesReturn{Worker: worker, Owner: owner, Coinbase: coinbase}, 0)
		batchPids := builtin.BatchPieceCIDParams{PieceCIDs: []builtin.CheckedCID{{CID: dealProposal.PieceCID}}}
		rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
		rt.ExpectSend(provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)

		/* expectQueryNetworkInfo(rt, actor) */
		rt.ExpectVerifySignature(crypto.Signature{}, client, mustCbor(&params.Deals[0].Proposal), nil)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.PublishStorageDeals, params)
		})

		rt.Verify()
	}
	actor.checkState(rt)
}

func TestComputeDataCommitment(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}
	start := abi.ChainEpoch(10)

	t.Run("successfully compute cid", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId1 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, start /* end, */, start, "deal1")
		d1 := actor.getDealProposal(rt, dealId1)

		dealId2 := actor.generateAndPublishRandDealWithLabel(rt, client, mAddrs, start /* end+1, */, start, "deal2")
		d2 := actor.getDealProposal(rt, dealId2)

		param := &market.ComputeDataCommitmentParams{DealIDs: []abi.DealID{dealId1, dealId2}, SectorType: 1}

		p1 := abi.PieceInfo{Size: d1.PieceSize, PieceCID: d1.PieceCID}
		p2 := abi.PieceInfo{Size: d2.PieceSize, PieceCID: d2.PieceCID}

		c := tutil.MakeCID("100", &market.PieceCIDPrefix)

		rt.ExpectComputeUnsealedSectorCID(1, []abi.PieceInfo{p1, p2}, c, nil)
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)

		ret := rt.Call(actor.ComputeDataCommitment, param)
		val, ok := ret.(*cbg.CborCid)
		require.True(t, ok)
		require.Equal(t, c, *(*cid.Cid)(val))
		rt.Verify()

		actor.checkState(rt)
	})

	t.Run("fail when deal proposal is absent", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		param := &market.ComputeDataCommitmentParams{DealIDs: []abi.DealID{1}, SectorType: 1}
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrNotFound, func() {
			rt.Call(actor.ComputeDataCommitment, param)
		})
		actor.checkState(rt)
	})

	t.Run("fail when syscall returns an error", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start /* end, */, start)
		d := actor.getDealProposal(rt, dealId)
		param := &market.ComputeDataCommitmentParams{DealIDs: []abi.DealID{dealId}, SectorType: 1}

		pi := abi.PieceInfo{Size: d.PieceSize, PieceCID: d.PieceCID}

		rt.ExpectComputeUnsealedSectorCID(1, []abi.PieceInfo{pi}, cid.Cid{}, errors.New("error"))
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.ComputeDataCommitment, param)
		})
		actor.checkState(rt)
	})
}

func TestVerifyDealsForActivation(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	mAddrs := &minerAddrs{owner, worker, coinbase, provider, expert, nil}
	start := abi.ChainEpoch(10)

	t.Run("verify deal and get deal weight for unverified deal proposal", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start, start)
		d := actor.getDealProposal(rt, dealId)

		resp := actor.verifyDealsForActivation(rt, provider, []market.SectorDeals{{
			DealIDs: []abi.DealID{dealId},
		}})
		var pss []uint64
		for _, size := range resp.Sectors[0].PieceSizes {
			pss = append(pss, uint64(size))
		}
		require.EqualValues(t, d.PieceSize, sum(pss...))

		actor.checkState(rt)
	})

	t.Run("verify deal and get deal weight for verified deal proposal", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		deal := actor.generateDealAndAddFunds(rt, client, mAddrs, start)
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		dealIds := actor.publishDeals(rt, mAddrs, publishDealReq{deal: deal})

		resp := actor.verifyDealsForActivation(rt, provider, []market.SectorDeals{{
			DealIDs: dealIds,
		}})
		require.EqualValues(t, deal.PieceSize, resp.Sectors[0].PieceSizes[0])

		actor.checkState(rt)
	})

	t.Run("verification and weights for verified and unverified deals", func(T *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)

		vd1 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, start /* , end */)
		vd1.Label = "vd1"
		// vd1.VerifiedDeal = true

		vd2 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, start /* , end+1 */)
		vd2.Label = "vd2"
		// vd2.VerifiedDeal = true

		d1 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, start /* , end+2 */)
		d1.Label = "d1"
		d2 := actor.generateRandDealAndAddFunds(rt, client, mAddrs, start /* , end+3 */)
		d2.Label = "d2"

		rt.SetCaller(worker, builtin.AccountActorCodeID)
		dealIds := actor.publishDeals(rt, mAddrs, publishDealReq{deal: vd1}, publishDealReq{deal: vd2},
			publishDealReq{deal: d1}, publishDealReq{deal: d2})

		resp := actor.verifyDealsForActivation(rt, provider, []market.SectorDeals{{
			DealIDs: dealIds,
		}})
		var pss []uint64
		for _, size := range resp.Sectors[0].PieceSizes {
			pss = append(pss, uint64(size))
		}
		require.EqualValues(t, d1.PieceSize+d2.PieceSize+vd1.PieceSize+vd2.PieceSize, sum(pss...))

		actor.checkState(rt)
	})

	t.Run("fail when caller is not a StorageMinerActor", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start /* end, */, start)

		param := &market.VerifyDealsForActivationParams{Sectors: []market.SectorDeals{{
			DealIDs: []abi.DealID{dealId},
		}}}
		rt.SetCaller(worker, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.SysErrForbidden, func() {
			rt.Call(actor.VerifyDealsForActivation, param)
		})
		actor.checkState(rt)
	})

	t.Run("fail when deal proposal is not found", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		param := &market.VerifyDealsForActivationParams{Sectors: []market.SectorDeals{{
			DealIDs: []abi.DealID{1},
		}}}
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrNotFound, func() {
			rt.Call(actor.VerifyDealsForActivation, param)
		})
		actor.checkState(rt)
	})

	t.Run("fail when caller is not the provider", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start, start)

		param := &market.VerifyDealsForActivationParams{Sectors: []market.SectorDeals{{
			DealIDs: []abi.DealID{dealId},
		}}}
		provider2 := tutil.NewIDAddr(t, 205)
		rt.SetCaller(provider2, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrForbidden, func() {
			rt.Call(actor.VerifyDealsForActivation, param)
		})
		actor.checkState(rt)
	})

	t.Run("fail when current epoch is greater than proposal start epoch", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start, start)

		rt.SetEpoch(start + 1)
		param := &market.VerifyDealsForActivationParams{Sectors: []market.SectorDeals{{
			DealIDs: []abi.DealID{dealId},
		}}}
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.VerifyDealsForActivation, param)
		})
		actor.checkState(rt)
	})

	/* t.Run("fail when deal end epoch is greater than sector expiration", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start, end, start)

		param := &market.VerifyDealsForActivationParams{Sectors: []market.SectorDeals{{
			SectorExpiry: end - 1,
			DealIDs:      []abi.DealID{dealId},
		}}}
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.VerifyDealsForActivation, param)
		})
		actor.checkState(rt)
	}) */

	t.Run("fail when the same deal ID is passed multiple times", func(t *testing.T) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		dealId := actor.generateAndPublishDeal(rt, client, mAddrs, start /* end, */, start)

		param := &market.VerifyDealsForActivationParams{Sectors: []market.SectorDeals{{
			DealIDs: []abi.DealID{dealId, dealId},
		}}}
		rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
		rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "multiple times", func() {
			rt.Call(actor.VerifyDealsForActivation, param)
		})
		actor.checkState(rt)
	})
}

func TestResetQuotas(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	coinbase := tutil.NewIDAddr(t, 106)

	setupFunc := func(grantedCode exitcode.ExitCode) (*mock.Runtime, *marketActorTestHarness) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)
		rt.ExpectSend(builtin.GovernActorAddr,
			builtin.MethodsGovern.ValidateGranted,
			&builtin.ValidateGrantedParams{
				Caller: owner,
				Method: builtin.MethodsMarket.ResetQuotas,
			},
			big.Zero(),
			nil,
			grantedCode,
		)
		return rt, actor
	}

	t.Run("empty params", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.Ok)
		rt.SetCaller(owner, builtin.AccountActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "empty params", func() {
			rt.Call(actor.ResetQuotas, &market.ResetQuotasParams{})
		})
	})

	t.Run("method not granted", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.ErrForbidden)
		rt.SetCaller(owner, builtin.AccountActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrForbidden, "method not granted", func() {
			rt.Call(actor.ResetQuotas, &market.ResetQuotasParams{})
		})
	})

	t.Run("piece cid not found", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.Ok)
		rt.SetCaller(owner, builtin.AccountActorCodeID)

		var st market.State
		rt.GetState(&st)

		quotas, err := adt.AsMap(adt.AsStore(rt), st.Quotas, builtin.DefaultHamtBitwidth)
		require.NoError(t, err)
		keys, err := quotas.CollectKeys()
		require.NoError(t, err)
		require.True(t, len(keys) == 0)

		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "piece cid not found", func() {
			rt.Call(actor.ResetQuotas, &market.ResetQuotasParams{
				NewQuotas: []market.NewQuota{
					{tutil.MakeCID("piece cid", &market.PieceCIDPrefix), 99},
				},
			})
		})
	})

	t.Run("illegal new quota", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.Ok)
		rt.SetCaller(owner, builtin.AccountActorCodeID)

		pieceCID := tutil.MakeCID("piece cid", &market.PieceCIDPrefix)

		var st market.State
		rt.GetState(&st)

		quotas, err := adt.AsMap(adt.AsStore(rt), st.Quotas, builtin.DefaultHamtBitwidth)
		require.NoError(t, err)
		in := cbg.CborInt(1000)
		err = quotas.Put(abi.CidKey(pieceCID), &in)
		require.NoError(t, err)

		var out cbg.CborInt
		found, err := quotas.Get(abi.CidKey(pieceCID), &out)
		require.NoError(t, err)
		require.True(t, found && in == out)

		st.Quotas, err = quotas.Root()
		require.NoError(t, err)
		rt.ReplaceState(&st)

		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "negative quota not allowed", func() {
			rt.Call(actor.ResetQuotas, &market.ResetQuotasParams{
				NewQuotas: []market.NewQuota{
					{pieceCID, -1},
				},
			})
		})
	})

	t.Run("successful reset quota", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.Ok)
		rt.SetCaller(owner, builtin.AccountActorCodeID)

		pieceCID := tutil.MakeCID("piece cid", &market.PieceCIDPrefix)

		var st market.State
		rt.GetState(&st)

		quotas, err := adt.AsMap(adt.AsStore(rt), st.Quotas, builtin.DefaultHamtBitwidth)
		require.NoError(t, err)
		in := cbg.CborInt(1000)
		err = quotas.Put(abi.CidKey(pieceCID), &in)
		require.NoError(t, err)

		st.Quotas, err = quotas.Root()
		require.NoError(t, err)
		rt.ReplaceState(&st)

		rt.Call(actor.ResetQuotas, &market.ResetQuotasParams{
			NewQuotas: []market.NewQuota{
				{pieceCID, 2000},
			},
		})

		rt.Verify()
		actor.checkState(rt)

		// check
		rt.GetState(&st)
		quotas, err = adt.AsMap(adt.AsStore(rt), st.Quotas, builtin.DefaultHamtBitwidth)
		require.NoError(t, err)

		var out cbg.CborInt
		found, err := quotas.Get(abi.CidKey(pieceCID), &out)
		require.NoError(t, err)
		require.True(t, found && out == 2000)
	})
}

func TestSetInitialQuota(t *testing.T) {
	owner := tutil.NewIDAddr(t, 101)
	provider := tutil.NewIDAddr(t, 102)
	worker := tutil.NewIDAddr(t, 103)
	client := tutil.NewIDAddr(t, 104)
	// expert := tutil.NewIDAddr(t, 105)
	coinbase := tutil.NewIDAddr(t, 106)
	// minerAddrs := &minerAddrs{owner, worker,coinbase, provider, expert}

	setupFunc := func(grantedCode exitcode.ExitCode) (*mock.Runtime, *marketActorTestHarness) {
		rt, actor := basicMarketSetup(t, owner, provider, worker, client, coinbase)
		rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)
		rt.ExpectSend(builtin.GovernActorAddr,
			builtin.MethodsGovern.ValidateGranted,
			&builtin.ValidateGrantedParams{
				Caller: owner,
				Method: builtin.MethodsMarket.SetInitialQuota,
			},
			big.Zero(),
			nil,
			grantedCode,
		)
		return rt, actor
	}

	t.Run("non-positive quota", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.Ok)
		rt.SetCaller(owner, builtin.AccountActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrIllegalArgument, "non-positive quota to set", func() {
			quota := cbg.CborInt(0)
			rt.Call(actor.SetInitialQuota, &quota)
		})
	})

	t.Run("method not granted", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.ErrForbidden)
		rt.SetCaller(owner, builtin.AccountActorCodeID)
		rt.ExpectAbortContainsMessage(exitcode.ErrForbidden, "method not granted", func() {
			quota := cbg.CborInt(1)
			rt.Call(actor.SetInitialQuota, &quota)
		})
	})

	t.Run("successfull set initial quota", func(t *testing.T) {
		rt, actor := setupFunc(exitcode.Ok)
		rt.SetCaller(owner, builtin.AccountActorCodeID)

		var st market.State
		rt.GetState(&st)

		require.True(t, st.InitialQuota == market.DefaultInitialQuota)

		quota := cbg.CborInt(market.DefaultInitialQuota + 100)
		rt.Call(actor.SetInitialQuota, &quota)

		rt.Verify()
		actor.checkState(rt)

		rt.GetState(&st)
		require.True(t, st.InitialQuota == int64(quota))
	})
}

type marketActorTestHarness struct {
	market.Actor
	t testing.TB

	networkQAPower       abi.StoragePower
	networkBaselinePower abi.StoragePower
}

func (h *marketActorTestHarness) constructAndVerify(rt *mock.Runtime) {
	rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)
	ret := rt.Call(h.Constructor, nil)
	assert.Nil(h.t, ret)
	rt.Verify()
}

func (h *marketActorTestHarness) verifyDealsForActivation(rt *mock.Runtime, provider address.Address,
	sectorDeals []market.SectorDeals) *market.VerifyDealsForActivationReturn {
	param := &market.VerifyDealsForActivationParams{Sectors: sectorDeals}
	rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)
	rt.SetCaller(provider, builtin.StorageMinerActorCodeID)

	ret := rt.Call(h.VerifyDealsForActivation, param)
	rt.Verify()

	val, ok := ret.(*market.VerifyDealsForActivationReturn)
	require.True(h.t, ok)
	require.NotNil(h.t, val)
	return val
}

type minerAddrs struct {
	owner    address.Address
	worker   address.Address
	coinbase address.Address
	provider address.Address
	expert   address.Address
	control  []address.Address
}

// addProviderFunds is a helper method to setup provider market funds
func (h *marketActorTestHarness) addProviderFunds(rt *mock.Runtime, amount abi.TokenAmount, minerAddrs *minerAddrs) {
	rt.SetReceived(amount)
	rt.SetAddressActorType(minerAddrs.provider, builtin.StorageMinerActorCodeID)
	rt.SetCaller(minerAddrs.owner, builtin.AccountActorCodeID)
	rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)

	expectGetControlAddresses(rt, minerAddrs.provider, minerAddrs.owner, minerAddrs.worker, minerAddrs.coinbase)

	rt.Call(h.AddBalance, &minerAddrs.provider)

	rt.Verify()

	rt.SetBalance(big.Add(rt.Balance(), amount))
}

// addParticipantFunds is a helper method to setup non-provider storage market participant funds
func (h *marketActorTestHarness) addParticipantFunds(rt *mock.Runtime, addr address.Address, amount abi.TokenAmount) {
	rt.SetReceived(amount)
	rt.SetCaller(addr, builtin.AccountActorCodeID)
	rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)

	rt.Call(h.AddBalance, &addr)

	rt.Verify()

	rt.SetBalance(big.Add(rt.Balance(), amount))
}

func (h *marketActorTestHarness) withdrawProviderBalance(rt *mock.Runtime, withDrawAmt, expectedSend abi.TokenAmount, miner *minerAddrs) {
	rt.SetCaller(miner.worker, builtin.AccountActorCodeID)
	rt.ExpectValidateCallerAddr(miner.owner, miner.worker)
	expectGetControlAddresses(rt, miner.provider, miner.owner, miner.worker, miner.coinbase)

	params := market.WithdrawBalanceParams{
		ProviderOrClientAddress: miner.provider,
		Amount:                  withDrawAmt,
	}

	rt.ExpectSend(miner.owner, builtin.MethodSend, nil, expectedSend, nil, exitcode.Ok)
	rt.Call(h.WithdrawBalance, &params)
	rt.Verify()
}

func (h *marketActorTestHarness) withdrawClientBalance(rt *mock.Runtime, client address.Address, withDrawAmt, expectedSend abi.TokenAmount) {
	rt.SetCaller(client, builtin.AccountActorCodeID)
	rt.ExpectSend(client, builtin.MethodSend, nil, expectedSend, nil, exitcode.Ok)
	rt.ExpectValidateCallerAddr(client)

	params := market.WithdrawBalanceParams{
		ProviderOrClientAddress: client,
		Amount:                  withDrawAmt,
	}

	rt.Call(h.WithdrawBalance, &params)
	rt.Verify()
}

func (h *marketActorTestHarness) cronTickNoChange(rt *mock.Runtime, client, provider address.Address) {
	var st market.State
	rt.GetState(&st)
	epochCid := st.DealOpsByEpoch

	// fetch current client and provider escrow balances
	cLocked := h.getLockedBalance(rt, client)
	cEscrow := h.getEscrowBalance(rt, client)
	pLocked := h.getLockedBalance(rt, provider)
	pEscrow := h.getEscrowBalance(rt, provider)

	h.cronTick(rt)

	rt.GetState(&st)
	require.True(h.t, epochCid.Equals(st.DealOpsByEpoch))

	require.EqualValues(h.t, cEscrow, h.getEscrowBalance(rt, client))
	require.EqualValues(h.t, cLocked, h.getLockedBalance(rt, client))
	require.EqualValues(h.t, pEscrow, h.getEscrowBalance(rt, provider))
	require.EqualValues(h.t, pLocked, h.getLockedBalance(rt, provider))
}

// if this is the first crontick for the deal, it's next tick will be scheduled at `desiredNextEpoch`
// if this is not the first crontick, the `desiredNextEpoch` param is ignored.
func (h *marketActorTestHarness) cronTickAndAssertBalances(rt *mock.Runtime, client, provider address.Address,
	currentEpoch abi.ChainEpoch, dealId abi.DealID) (payment abi.TokenAmount, amountSlashed abi.TokenAmount) {
	// fetch current client and provider escrow balances
	cLocked := h.getLockedBalance(rt, client)
	cEscrow := h.getEscrowBalance(rt, client)
	pLocked := h.getLockedBalance(rt, provider)
	pEscrow := h.getEscrowBalance(rt, provider)
	amountSlashed = big.Zero()

	s := h.getDealState(rt, dealId)
	/* d := h.getDealProposal(rt, dealId)

	// end epoch for payment calc
	paymentEnd := d.EndEpoch
	if s.SlashEpoch != -1 {
		rt.ExpectSend(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, d.ProviderCollateral, nil, exitcode.Ok)
		amountSlashed = d.ProviderCollateral

		if s.SlashEpoch < d.StartEpoch {
			paymentEnd = d.StartEpoch
		} else {
			paymentEnd = s.SlashEpoch
		}
	} else if currentEpoch < paymentEnd {
		paymentEnd = currentEpoch
	}

	// start epoch for payment calc
	paymentStart := d.StartEpoch
	if s.LastUpdatedEpoch != -1 {
		paymentStart = s.LastUpdatedEpoch
	}
	duration := paymentEnd - paymentStart
	payment = big.Mul(big.NewInt(int64(duration)), d.StoragePricePerEpoch) */
	payment = big.Zero()

	// expected updated amounts
	updatedClientEscrow := big.Sub(cEscrow, payment)
	updatedProviderEscrow := big.Add(pEscrow, payment)
	updatedProviderEscrow = big.Sub(updatedProviderEscrow, amountSlashed)
	updatedClientLocked := big.Sub(cLocked, payment)
	updatedProviderLocked := pLocked
	/* // if the deal has expired or been slashed, locked amount will be zero for provider and client.
	isDealExpired := paymentEnd == d.EndEpoch */
	if /* isDealExpired || */ s.SlashEpoch != -1 {
		updatedClientLocked = big.Zero()
		updatedProviderLocked = big.Zero()
	}

	h.cronTick(rt)

	require.EqualValues(h.t, updatedClientEscrow, h.getEscrowBalance(rt, client))
	require.EqualValues(h.t, updatedClientLocked, h.getLockedBalance(rt, client))
	require.Equal(h.t, updatedProviderLocked, h.getLockedBalance(rt, provider))
	require.Equal(h.t, updatedProviderEscrow.Int64(), h.getEscrowBalance(rt, provider).Int64())

	return
}

func (h *marketActorTestHarness) cronTick(rt *mock.Runtime) {
	rt.ExpectValidateCallerAddr(builtin.CronActorAddr)
	rt.SetCaller(builtin.CronActorAddr, builtin.CronActorCodeID)
	param := abi.EmptyValue{}

	rt.Call(h.CronTick, &param)
	rt.Verify()
}

type publishDealReq struct {
	deal                 market.DealProposal
	requiredProcessEpoch abi.ChainEpoch
}

func (h *marketActorTestHarness) expectGetRandom(rt *mock.Runtime, deal *market.DealProposal, requiredProcessEpoch abi.ChainEpoch) {
	dealBuf := bytes.Buffer{}
	epochBuf := bytes.Buffer{}

	diff := uint64(requiredProcessEpoch - deal.StartEpoch)
	require.NoError(h.t, deal.MarshalCBOR(&dealBuf))
	require.NoError(h.t, binary.Write(&epochBuf, binary.BigEndian, diff))
	rt.ExpectGetRandomnessBeacon(crypto.DomainSeparationTag_MarketDealCronSeed, rt.Epoch()-1, dealBuf.Bytes(), epochBuf.Bytes())
}

func (h *marketActorTestHarness) publishDeals(rt *mock.Runtime, minerAddrs *minerAddrs, publishDealReqs ...publishDealReq) []abi.DealID {
	for _, pdr := range publishDealReqs {
		h.expectGetRandom(rt, &pdr.deal, pdr.requiredProcessEpoch)
	}

	rt.ExpectValidateCallerType(builtin.CallerTypesSignable...)
	rt.ExpectSend(
		minerAddrs.provider,
		builtin.MethodsMiner.ControlAddresses,
		nil,
		big.Zero(),
		&builtin.GetControlAddressesReturn{Owner: minerAddrs.owner, Worker: minerAddrs.worker, Coinbase: minerAddrs.coinbase, ControlAddrs: minerAddrs.control},
		exitcode.Ok,
	)
	batchPids := builtin.BatchPieceCIDParams{}
	for _, dr := range publishDealReqs {
		batchPids.PieceCIDs = append(batchPids.PieceCIDs, builtin.CheckedCID{CID: dr.deal.PieceCID})
	}
	rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchCheckData, &batchPids, big.Zero(), nil, exitcode.Ok)
	rt.ExpectSend(minerAddrs.provider, builtin.MethodsMiner.StoredAny, &batchPids, big.Zero(), &cbgFalse, exitcode.Ok)
	/* expectQueryNetworkInfo(rt, h) */

	params := market.PublishStorageDealsParams{
		DataRef: market.PublishStorageDataRef{
			Expert: minerAddrs.expert.String(),
		},
	}

	for _, pdr := range publishDealReqs {
		//  create a client proposal with a valid signature
		buf := bytes.Buffer{}
		require.NoError(h.t, pdr.deal.MarshalCBOR(&buf), "failed to marshal deal proposal")
		sig := crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte("does not matter")}
		clientProposal := market.ClientDealProposal{Proposal: pdr.deal, ClientSignature: sig}
		params.Deals = append(params.Deals, clientProposal)

		// expect a call to verify the above signature
		rt.ExpectVerifySignature(sig, pdr.deal.Client, buf.Bytes(), nil)
	}

	ret := rt.Call(h.PublishStorageDeals, &params)
	rt.Verify()

	resp, ok := ret.(*market.PublishStorageDealsReturn)
	require.True(h.t, ok, "unexpected type returned from call to PublishStorageDeals")
	require.Len(h.t, resp.IDs, len(publishDealReqs))

	// assert state after publishing the deals
	dealIds := resp.IDs
	for i, deaId := range dealIds {
		expected := publishDealReqs[i].deal
		p := h.getDealProposal(rt, deaId)

		require.Equal(h.t, expected.StartEpoch, p.StartEpoch)
		/* require.Equal(h.t, expected.EndEpoch, p.EndEpoch) */
		require.Equal(h.t, expected.PieceCID, p.PieceCID)
		require.Equal(h.t, expected.PieceSize, p.PieceSize)
		require.Equal(h.t, expected.Client, p.Client)
		require.Equal(h.t, expected.Provider, p.Provider)
		require.Equal(h.t, expected.Label, p.Label)
		require.Equal(h.t, expected.PieceSize, p.PieceSize)
		/* require.Equal(h.t, expected.VerifiedDeal, p.VerifiedDeal)
		require.Equal(h.t, expected.StoragePricePerEpoch, p.StoragePricePerEpoch)
		require.Equal(h.t, expected.ClientCollateral, p.ClientCollateral)
		require.Equal(h.t, expected.ProviderCollateral, p.ProviderCollateral) */
	}

	return resp.IDs
}

func (h *marketActorTestHarness) assertDealsNotActivated(rt *mock.Runtime, epoch abi.ChainEpoch, dealIDs ...abi.DealID) {
	var st market.State
	rt.GetState(&st)

	states, err := market.AsDealStateArray(adt.AsStore(rt), st.States)
	require.NoError(h.t, err)

	for _, d := range dealIDs {
		_, found, err := states.Get(d)
		require.NoError(h.t, err)
		require.False(h.t, found)
	}
}

func (h *marketActorTestHarness) activateDeals(rt *mock.Runtime /* sectorExpiry abi.ChainEpoch, */, provider address.Address,
	currentEpoch abi.ChainEpoch, dealIDs ...abi.DealID) interface{} {
	rt.SetCaller(provider, builtin.StorageMinerActorCodeID)
	rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)

	batchPids := builtin.BatchPieceCIDParams{}
	for _, d := range dealIDs {
		proposal := h.getDealProposal(rt, d)
		batchPids.PieceCIDs = append(batchPids.PieceCIDs, builtin.CheckedCID{CID: proposal.PieceCID})
	}
	rt.ExpectSend(builtin.ExpertFundActorAddr, builtin.MethodsExpertFunds.BatchStoreData, &batchPids, big.Zero(), nil, exitcode.Ok)

	params := &market.ActivateDealsParams{DealIDs: dealIDs /* SectorExpiry: sectorExpiry */}

	ret := rt.Call(h.ActivateDeals, params)
	rt.Verify()

	for _, d := range dealIDs {
		s := h.getDealState(rt, d)
		require.EqualValues(h.t, currentEpoch, s.SectorStartEpoch)
	}
	return ret
}

func (h *marketActorTestHarness) getDealProposal(rt *mock.Runtime, dealID abi.DealID) *market.DealProposal {
	var st market.State
	rt.GetState(&st)

	deals, err := market.AsDealProposalArray(adt.AsStore(rt), st.Proposals)
	require.NoError(h.t, err)

	d, found, err := deals.Get(dealID)
	require.NoError(h.t, err)
	require.True(h.t, found)
	require.NotNil(h.t, d)

	return d
}

func (h *marketActorTestHarness) assertAccountBalance(rt *mock.Runtime, addr address.Address, escrow, locked abi.TokenAmount) {
	var st market.State
	rt.GetState(&st)

	et, err := adt.AsBalanceTable(adt.AsStore(rt), st.EscrowTable)
	require.NoError(h.t, err)

	b, err := et.Get(addr)
	require.NoError(h.t, err)
	require.Equal(h.t, escrow, b)

	lt, err := adt.AsBalanceTable(adt.AsStore(rt), st.LockedTable)
	require.NoError(h.t, err)
	b, err = lt.Get(addr)
	require.NoError(h.t, err)
	require.Equal(h.t, locked, b)
}

func (h *marketActorTestHarness) getEscrowBalance(rt *mock.Runtime, addr address.Address) abi.TokenAmount {
	var st market.State
	rt.GetState(&st)

	et, err := adt.AsBalanceTable(adt.AsStore(rt), st.EscrowTable)
	require.NoError(h.t, err)

	bal, err := et.Get(addr)
	require.NoError(h.t, err)

	return bal
}

func (h *marketActorTestHarness) getLockedBalance(rt *mock.Runtime, addr address.Address) abi.TokenAmount {
	var st market.State
	rt.GetState(&st)

	lt, err := adt.AsBalanceTable(adt.AsStore(rt), st.LockedTable)
	require.NoError(h.t, err)

	bal, err := lt.Get(addr)
	require.NoError(h.t, err)

	return bal
}

func (h *marketActorTestHarness) getDealState(rt *mock.Runtime, dealID abi.DealID) *market.DealState {
	var st market.State
	rt.GetState(&st)

	states, err := market.AsDealStateArray(adt.AsStore(rt), st.States)
	require.NoError(h.t, err)

	s, found, err := states.Get(dealID)
	require.NoError(h.t, err)
	require.True(h.t, found)
	require.NotNil(h.t, s)

	return s
}

/* func (h *marketActorTestHarness) assertLockedFundStates(rt *mock.Runtime, storageFee, providerCollateral, clientCollateral abi.TokenAmount) {
	var st market.State
	rt.GetState(&st)

	require.Equal(h.t, clientCollateral, st.TotalClientLockedCollateral)
	require.Equal(h.t, providerCollateral, st.TotalProviderLockedCollateral)
	require.Equal(h.t, storageFee, st.TotalClientStorageFee)
} */

func (h *marketActorTestHarness) assertDealDeleted(rt *mock.Runtime, dealId abi.DealID, p *market.DealProposal) {
	var st market.State
	rt.GetState(&st)

	proposals, err := market.AsDealProposalArray(adt.AsStore(rt), st.Proposals)
	require.NoError(h.t, err)
	_, found, err := proposals.Get(dealId)
	require.NoError(h.t, err)
	require.False(h.t, found)

	states, err := market.AsDealStateArray(adt.AsStore(rt), st.States)
	require.NoError(h.t, err)
	_, found, err = states.Get(dealId)
	require.NoError(h.t, err)
	require.False(h.t, found)

	pcid, err := p.Cid()
	require.NoError(h.t, err)
	pending, err := adt.AsMap(adt.AsStore(rt), st.PendingProposals, builtin.DefaultHamtBitwidth)
	require.NoError(h.t, err)
	found, err = pending.Get(abi.CidKey(pcid), nil)
	require.NoError(h.t, err)
	require.False(h.t, found)
}

func (h *marketActorTestHarness) assertDealsTerminated(rt *mock.Runtime, epoch abi.ChainEpoch, dealIds ...abi.DealID) {
	for _, d := range dealIds {
		s := h.getDealState(rt, d)
		require.EqualValues(h.t, epoch, s.SlashEpoch)
	}
}

func (h *marketActorTestHarness) assertDeaslNotTerminated(rt *mock.Runtime, dealIds ...abi.DealID) {
	for _, d := range dealIds {
		s := h.getDealState(rt, d)
		require.EqualValues(h.t, abi.ChainEpoch(-1), s.SlashEpoch)
	}
}

func (h *marketActorTestHarness) terminateDeals(rt *mock.Runtime, minerAddr address.Address, dealIds ...abi.DealID) {
	rt.SetCaller(minerAddr, builtin.StorageMinerActorCodeID)
	rt.ExpectValidateCallerType(builtin.StorageMinerActorCodeID)

	params := mkTerminateDealParams(rt.Epoch(), dealIds...)

	ret := rt.Call(h.OnMinerSectorsTerminate, params)
	rt.Verify()
	require.Nil(h.t, ret)
}

func (h *marketActorTestHarness) publishAndActivateDeal(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch, currentEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch) abi.DealID {
	deal := h.generateDealAndAddFunds(rt, client, minerAddrs, startEpoch)
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	h.activateDeals(rt, minerAddrs.provider, currentEpoch, dealIds[0])
	return dealIds[0]
}

func (h *marketActorTestHarness) publishAndActivateRandDeal(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch, currentEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch) abi.DealID {
	deal := h.generateRandDealAndAddFunds(rt, client, minerAddrs, startEpoch)
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	h.activateDeals(rt, minerAddrs.provider, currentEpoch, dealIds[0])
	return dealIds[0]
}

func (h *marketActorTestHarness) publishAndActivateDealWithLabel(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch /* endEpoch, */, currentEpoch /* , sectorExpiry */ abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch, lable string) abi.DealID {
	deal := h.generateDealAndAddFunds(rt, client, minerAddrs, startEpoch /* , endEpoch */)
	deal.Label = lable
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	h.activateDeals(rt /* sectorExpiry, */, minerAddrs.provider, currentEpoch, dealIds[0])
	return dealIds[0]
}

func (h *marketActorTestHarness) publishAndActivateRandDealWithLabel(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch /* endEpoch, */, currentEpoch /* , sectorExpiry */ abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch, lable string) abi.DealID {
	deal := h.generateRandDealAndAddFunds(rt, client, minerAddrs, startEpoch /* , endEpoch */)
	deal.Label = lable
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	h.activateDeals(rt /* sectorExpiry, */, minerAddrs.provider, currentEpoch, dealIds[0])
	return dealIds[0]
}

func (h *marketActorTestHarness) updateLastUpdated(rt *mock.Runtime, dealId abi.DealID, newLastUpdated abi.ChainEpoch) {
	var st market.State
	// rt.StateTransaction(&st, func() {
	// 	states, err := market.AsDealStateArray(adt.AsStore(rt), st.States)
	// 	require.NoError(h.t, err)
	// 	s, found, err := states.Get(dealId)
	// 	require.True(h.t, found)
	// 	require.NoError(h.t, err)
	// 	require.NotNil(h.t, s)

	// 	require.NoError(h.t, states.Set(dealId, &market.DealState{s.SectorStartEpoch, newLastUpdated, s.SlashEpoch}))
	// 	st.States, err = states.Root()
	// 	require.NoError(h.t, err)
	// })
	rt.GetState(&st)
	states, err := market.AsDealStateArray(adt.AsStore(rt), st.States)
	require.NoError(h.t, err)
	s, found, err := states.Get(dealId)
	require.True(h.t, found)
	require.NoError(h.t, err)
	require.NotNil(h.t, s)

	require.NoError(h.t, states.Set(dealId, &market.DealState{s.SectorStartEpoch, newLastUpdated, s.SlashEpoch}))
	st.States, err = states.Root()
	require.NoError(h.t, err)
	rt.ReplaceState(&st)
}

func (h *marketActorTestHarness) deleteDealProposal(rt *mock.Runtime, dealId abi.DealID) {
	var st market.State

	// rt.StateTransaction(&st, func() {
	// 	deals, err := market.AsDealProposalArray(adt.AsStore(rt), st.Proposals)
	// 	require.NoError(h.t, err)
	// 	require.NoError(h.t, deals.Delete(uint64(dealId)))
	// 	st.Proposals, err = deals.Root()
	// 	require.NoError(h.t, err)
	// })
	rt.GetState(&st)
	deals, err := market.AsDealProposalArray(adt.AsStore(rt), st.Proposals)
	require.NoError(h.t, err)
	require.NoError(h.t, deals.Delete(dealId))
	st.Proposals, err = deals.Root()
	require.NoError(h.t, err)
	rt.ReplaceState(&st)
}

func (h *marketActorTestHarness) setLastCron(rt *mock.Runtime, epoch abi.ChainEpoch) {
	var st market.State
	rt.StateTransaction(&st, func() {
		st.LastCron = epoch
	})
}

func (h *marketActorTestHarness) generateAndPublishDeal(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch) abi.DealID {

	deal := h.generateDealAndAddFunds(rt, client, minerAddrs, startEpoch)
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	return dealIds[0]
}

func (h *marketActorTestHarness) generateAndPublishRandDeal(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch) abi.DealID {

	deal := h.generateRandDealAndAddFunds(rt, client, minerAddrs, startEpoch)
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	return dealIds[0]
}

func (h *marketActorTestHarness) generateAndPublishDealWithFunds(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch,
	clientFunds, providerFunds abi.TokenAmount) abi.DealID {
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	deal := generateDealProposal(client, minerAddrs.provider, startEpoch)
	if !providerFunds.IsZero() {
		h.addProviderFunds(rt, providerFunds, minerAddrs)
	}
	if !clientFunds.IsZero() {
		h.addParticipantFunds(rt, client, clientFunds)
	}
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	return dealIds[0]
}

func (h *marketActorTestHarness) generateAndPublishDealWithLabel(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch, label string) abi.DealID {

	deal := h.generateDealAndAddFunds(rt, client, minerAddrs, startEpoch)
	deal.Label = label
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	return dealIds[0]
}

func (h *marketActorTestHarness) generateAndPublishRandDealWithLabel(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch, requiredProcessEpoch abi.ChainEpoch, label string) abi.DealID {

	deal := h.generateRandDealAndAddFunds(rt, client, minerAddrs, startEpoch)
	deal.Label = label
	rt.SetCaller(minerAddrs.worker, builtin.AccountActorCodeID)
	dealIds := h.publishDeals(rt, minerAddrs, publishDealReq{deal: deal, requiredProcessEpoch: requiredProcessEpoch})
	return dealIds[0]
}

func (h *marketActorTestHarness) generateDealAndAddFunds(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch) market.DealProposal {
	deal4 := generateDealProposal(client, minerAddrs.provider, startEpoch)
	h.addProviderFunds(rt, big.NewInt(10), minerAddrs)
	h.addParticipantFunds(rt, client, big.NewInt(10))

	return deal4
}

func (h *marketActorTestHarness) generateRandDealAndAddFunds(rt *mock.Runtime, client address.Address, minerAddrs *minerAddrs,
	startEpoch abi.ChainEpoch) market.DealProposal {
	var buf [10]byte
	rand.Read(buf[:])
	pieceCid := tutil.MakeCID(string(buf[:]), &market.PieceCIDPrefix)
	pieceSize := abi.PaddedPieceSize(2048)
	deal := market.DealProposal{PieceCID: pieceCid, PieceSize: pieceSize, Client: client, Provider: minerAddrs.provider, Label: "label", StartEpoch: startEpoch}
	h.addProviderFunds(rt, big.NewInt(10), minerAddrs)
	h.addParticipantFunds(rt, client, big.NewInt(10))

	return deal
}

func (h *marketActorTestHarness) generateDealWithCollateralAndAddFunds(rt *mock.Runtime, client address.Address,
	minerAddrs *minerAddrs, providerFunds, clientFunds abi.TokenAmount, startEpoch /* , endEpoch */ abi.ChainEpoch) market.DealProposal {
	deal := generateDealProposalWithCollateral(client, minerAddrs.provider, /* providerCollateral, clientCollateral, */
		startEpoch /* , endEpoch */)
	h.addProviderFunds(rt, providerFunds, minerAddrs)
	h.addParticipantFunds(rt, client, clientFunds)

	return deal
}

func (h *marketActorTestHarness) checkState(rt *mock.Runtime) *market.StateSummary {
	var st market.State
	rt.GetState(&st)
	sum, msgs := market.CheckStateInvariants(&st, rt.AdtStore(), rt.Balance(), rt.Epoch())
	assert.True(h.t, msgs.IsEmpty(), strings.Join(msgs.Messages(), "\n"))
	return sum
}

var pieceCid1 = tutil.MakeCID("1", &market.PieceCIDPrefix)

func generateDealProposalWithCollateral(client, provider address.Address, startEpoch abi.ChainEpoch) market.DealProposal {
	pieceCid := pieceCid1
	pieceSize := abi.PaddedPieceSize(2048)
	return market.DealProposal{PieceCID: pieceCid, PieceSize: pieceSize, Client: client, Provider: provider, Label: "label", StartEpoch: startEpoch}
}

func generateDealProposal(client, provider address.Address, startEpoch abi.ChainEpoch) market.DealProposal {
	/* clientCollateral := big.NewInt(10)
	providerCollateral := big.NewInt(10) */

	return generateDealProposalWithCollateral(client, provider, startEpoch)
}

func basicMarketSetup(t *testing.T, owner, provider, worker, client, coinbase address.Address) (*mock.Runtime, *marketActorTestHarness) {
	builder := mock.NewBuilder(context.Background(), builtin.StorageMarketActorAddr).
		WithCaller(builtin.SystemActorAddr, builtin.SystemActorCodeID).
		WithBalance(big.Mul(big.NewInt(10), big.NewInt(1e18)), big.Zero()).
		WithActorType(owner, builtin.AccountActorCodeID).
		WithActorType(worker, builtin.AccountActorCodeID).
		WithActorType(coinbase, builtin.AccountActorCodeID).
		WithActorType(provider, builtin.StorageMinerActorCodeID).
		WithActorType(client, builtin.AccountActorCodeID)

	rt := builder.Build(t)
	power := abi.NewStoragePower(1 << 50)
	actor := marketActorTestHarness{
		t:                    t,
		networkQAPower:       power,
		networkBaselinePower: power,
	}
	actor.constructAndVerify(rt)

	return rt, &actor
}

func mkPublishStorageParams(proposals ...market.DealProposal) *market.PublishStorageDealsParams {
	m := &market.PublishStorageDealsParams{}
	for _, p := range proposals {
		m.Deals = append(m.Deals, market.ClientDealProposal{Proposal: p})
	}
	return m
}

func mkActivateDealParams(sectorExpiry abi.ChainEpoch, dealIds ...abi.DealID) *market.ActivateDealsParams {
	return &market.ActivateDealsParams{ /* SectorExpiry: sectorExpiry, */ DealIDs: dealIds}
}

func mkTerminateDealParams(epoch abi.ChainEpoch, dealIds ...abi.DealID) *market.OnMinerSectorsTerminateParams {
	return &market.OnMinerSectorsTerminateParams{Epoch: epoch, DealIDs: dealIds}
}

func expectGetControlAddresses(rt *mock.Runtime, provider address.Address, owner, worker, coinbase address.Address, controls ...address.Address) {
	result := &builtin.GetControlAddressesReturn{Owner: owner, Worker: worker, Coinbase: coinbase, ControlAddrs: controls}
	rt.ExpectSend(
		provider,
		builtin.MethodsMiner.ControlAddresses,
		nil,
		big.Zero(),
		result,
		exitcode.Ok,
	)
}

/*
func expectQueryNetworkInfo(rt *mock.Runtime, h *marketActorTestHarness) {
	currentPower := power.CurrentTotalPowerReturn{
		QualityAdjPower: h.networkQAPower,
	}
	currentReward := reward.ThisEpochRewardReturn{
		ThisEpochBaselinePower: h.networkBaselinePower,
	}
	rt.ExpectSend(
		builtin.RewardActorAddr,
		builtin.MethodsReward.ThisEpochReward,
		nil,
		big.Zero(),
		&currentReward,
		exitcode.Ok,
	)

	rt.ExpectSend(
		builtin.StoragePowerActorAddr,
		builtin.MethodsPower.CurrentTotalPower,
		nil,
		big.Zero(),
		&currentPower,
		exitcode.Ok,
	)
}
*/

func sum(s ...uint64) (sum uint64) {
	for _, a := range s {
		sum += a
	}
	return sum
}
