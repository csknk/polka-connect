package main

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey"
)

var (
	blockHashStringWestend = "0xd4cba21f5ee0078c21af2e5887cd3c7248ee1ed74bb938530dd034361859a456"
	blockHashString        = "0xd8030c1a1cdb40f8c53f6a0d1db9e6b63ff500ddeda7c6074100d2012689f387"
)

func (c *Connection) getExtrinsic(blockHashStr string, index uint8) error {
	// Make a types.Hash object from a hexstring
	blockHash, err := types.NewHashFromHexString(blockHashString)
	if err != nil {
		return fmt.Errorf("hash from hex string %s: %w", blockHashStr, err)
	}

	// Get the block
	block, err := c.Api.RPC.Chain.GetBlock(blockHash)
	if err != nil {
		return fmt.Errorf("error getting block for hash %s: %w", blockHashString, err)
	}
	if block.Block.Header.Number == 0 {
		return fmt.Errorf("can't get data for block hash %s - it may not exist", blockHashString)
	}

	for i, extrinsic := range block.Block.Extrinsics {
		h, _ := types.GetHash(extrinsic)
		fmt.Printf("%#x\n", h)

		who := extrinsic.Signature.Signer.AsID
		fmt.Printf("extrinsic %d signed by: %#x\n", i, who)
	}

	// The metadata is required to properly decode the data in this block.
	// TODO check this.
	// <-------------------------------------------------------------------------------------------------
	meta, err := c.getMetadata(blockHash)
	if err != nil {
		return fmt.Errorf("error getting meta data latest: %w", err)
	}

	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return fmt.Errorf("error creating storage key: %w", err)
	}

	raw, err := c.Api.RPC.State.GetStorageRaw(key, blockHash)
	if err != nil {
		return fmt.Errorf("error retrieving raw storage data for key %v: %w", key, err)
	}

	events := types.EventRecords{}
	err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
	if err != nil {
		return fmt.Errorf("error decoding event records for %v: %w", *raw, err)
	}

	for _, event := range events.Balances_Transfer {
		send, _ := subkey.SS58Address(event.From[:], 0) // 0 is the network identifier byte
		to, _ := subkey.SS58Address(event.To[:], 0)
		fmt.Printf("from: %+v\n", send)
		fmt.Printf("to: %+v\n", to)
		fmt.Printf("value : %+v\n", event.Value)
		fmt.Printf("phase : %+v\n", event.Phase)
		fmt.Printf("topics : %+v\n", event.Topics)

		ext := block.Block.Extrinsics[int(event.Phase.AsApplyExtrinsic)]

		//		fmt.Printf("ext : %+v\n", ext)

		fmt.Printf("nonce : %+v\n", ext.Signature.Nonce)

		fmt.Printf("tip : %+v\n", ext.Signature.Tip)

		extBytes, err := types.EncodeToHexString(ext)
		if err != nil {
			panic(err)
		}
		fmt.Println(extBytes)

		resInter := Fee{}
		err = c.Api.Client.Call(&resInter, "payment_queryInfo", ext, blockHash.Hex())
		if err != nil {
			panic(err)
		}

		fmt.Println("PartialFee: ", resInter.PartialFee)
	}
	return nil
}

type Fee struct {
	Weight     types.Weight
	Class      string
	PartialFee string
}

func (c *Connection) FilterBlockForRequiredExtrinsics(blockHashBytes, accountID []byte) error {
	blockHash := types.NewHash(blockHashBytes)

	block, err := c.GetBlockByHash(blockHash)
	if err != nil {
		return fmt.Errorf("error getting block for hash %s: %#x", blockHash, err)
	}
	if block.Block.Header.Number == 0 {
		return fmt.Errorf("can't get data for block hash %s - it may not exist", blockHashString)
	}

	meta, err := c.getMetadata(blockHash)
	if err != nil {
		return fmt.Errorf("error getting meta data latest: %w", err)
	}
	// NOTE: Assumes that the transfer was made using transfer_keep_alive call. It's possible that the
	// transfer used "Balances.transfer" so we should allow either.
	callIndex, err := meta.FindCallIndex("Balances.transfer_keep_alive")
	if err != nil {
		return fmt.Errorf("error getting callIndex: %w", err)
	}

	for i, extrinsic := range block.Block.Extrinsics {
		fmt.Println("extrinsic ", i)

		if extrinsic.Method.CallIndex != callIndex {
			continue
		}

		signerPubKey := []byte(extrinsic.Signature.Signer.AsID[:])
		fmt.Println(signerPubKey)

		fmt.Printf("processing extrinsic %v in block %#x\n", i, blockHashBytes)
		h, _ := types.GetHash(extrinsic)
		fmt.Printf("from GetHash: %#x\n", h)
		who := extrinsic.Signature.Signer.AsID
		fmt.Printf("extrinsic %d signed by: %#x\n", i, who)

		decodedArgs, err := DecodeExtrinsicArgs(&extrinsic)
		if err != nil {
			return err
		}

		fmt.Println(decodedArgs)
	}
	timestamp, err := c.GetBlockTimestamp(block, blockHash)
	if err != nil {
		return fmt.Errorf("timestamp: %w", err)
	}
	fmt.Println("timestamp: ", timestamp)

	return nil
}

func (c *Connection) DecodeEvents(blockHashBytes []byte) error {
	blockHash := types.NewHash(blockHashBytes)

	block, err := c.GetBlockByHash(blockHash)
	if err != nil {
		return fmt.Errorf("error getting block for hash %s: %#x", blockHash, err)
	}
	if block.Block.Header.Number == 0 {
		return fmt.Errorf("can't get data for block hash %s - it may not exist", blockHashString)
	}

	meta, err := c.getMetadata(blockHash)
	if err != nil {
		return fmt.Errorf("error getting meta data latest: %w", err)
	}
	_ = meta // TODO: remove
	/*
		// TODO: unable to decode field 4 event #2 with EventID [5 8], field Balances_Withdraw:
		// expected more bytes, but could not decode any more - problem with custom
		// EventBalancesWithdraw struct 0- check Rust ref for the required data types...
		//

		key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
		if err != nil {
			return fmt.Errorf("error creating storage key: %w", err)
		}

		raw, err := c.Api.RPC.State.GetStorageRaw(key, blockHash)
		if err != nil {
			return fmt.Errorf("error retrieving raw storage data for key %v: %w", key, err)
		}

		//	events := types.EventRecords{}
		events := EventRecords{}
		err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
		if err != nil {
			return fmt.Errorf("error decoding event records for %v: %w", *raw, err)
		}
		for _, event := range events.Balances_Transfer {
			fmt.Println("event.Value = ", event.Value)

		}
	*/
	return nil
}

func DecodeExtrinsicArgs(extrinsic *types.Extrinsic) (*ExtrinsicArgs, error) {
	hash, err := types.GetHash(extrinsic)
	if err != nil {
		return nil, fmt.Errorf("problem getting extrinsic hash: %w", err)
	}
	txHash := hash[:]

	argsDecoder := scale.NewDecoder(bytes.NewReader(extrinsic.Method.Args))
	nCalls, err := argsDecoder.DecodeUintCompact()
	if err != nil {
		return nil, fmt.Errorf("failed to decode call count for extrinsic %#x: %w", txHash, err)
	}

	accountID := types.AccountID{}
	err = argsDecoder.Decode(&accountID)
	if err != nil {
		return nil, fmt.Errorf("problem accountID for extrinsic %#x: %w", txHash, err)
	}

	amount, err := argsDecoder.DecodeUintCompact()
	if err != nil {
		return nil, fmt.Errorf("problem decoding amount for extrinsic %#x: %w", txHash, err)
	}

	return &ExtrinsicArgs{
		Amount:         *amount,
		NCalls:         *nCalls,
		ReceiverPubKey: []byte(accountID[:]),
		TxHash:         txHash,
	}, nil
}

type ExtrinsicArgs struct {
	NCalls         big.Int
	Amount         big.Int
	ReceiverPubKey []byte
	TxHash         []byte
}

func (ext ExtrinsicArgs) String() string {
	return fmt.Sprintf("NCalls: %d\nAmount: %d\nReceiver PubKey: %#x\nTxHash: %#x\n",
		ext.NCalls.Int64(),
		ext.Amount.Int64(),
		ext.ReceiverPubKey,
		ext.TxHash,
	)
}

// EventRecords is a default set of possible event records that can be used as a target for
// `func (e EventRecordsRaw) Decode(...`
type EventRecords struct {
	Claims_Claimed                     []types.EventClaimsClaimed                     //nolint:stylecheck,golint
	Balances_Endowed                   []types.EventBalancesEndowed                   //nolint:stylecheck,golint
	Balances_DustLost                  []types.EventBalancesDustLost                  //nolint:stylecheck,golint
	Balances_Transfer                  []types.EventBalancesTransfer                  //nolint:stylecheck,golint
	Balances_BalanceSet                []types.EventBalancesBalanceSet                //nolint:stylecheck,golint
	Balances_Deposit                   []types.EventBalancesDeposit                   //nolint:stylecheck,golint
	Balances_Reserved                  []types.EventBalancesReserved                  //nolint:stylecheck,golint
	Balances_Unreserved                []types.EventBalancesUnreserved                //nolint:stylecheck,golint
	Balances_ReservedRepatriated       []types.EventBalancesReserveRepatriated        //nolint:stylecheck,golint
	Balances_Withdraw                  []EventBalancesWithdraw                        //nolint:stylecheck,golint
	Grandpa_NewAuthorities             []types.EventGrandpaNewAuthorities             //nolint:stylecheck,golint
	Grandpa_Paused                     []types.EventGrandpaPaused                     //nolint:stylecheck,golint
	Grandpa_Resumed                    []types.EventGrandpaResumed                    //nolint:stylecheck,golint
	ImOnline_HeartbeatReceived         []types.EventImOnlineHeartbeatReceived         //nolint:stylecheck,golint
	ImOnline_AllGood                   []types.EventImOnlineAllGood                   //nolint:stylecheck,golint
	ImOnline_SomeOffline               []types.EventImOnlineSomeOffline               //nolint:stylecheck,golint
	Indices_IndexAssigned              []types.EventIndicesIndexAssigned              //nolint:stylecheck,golint
	Indices_IndexFreed                 []types.EventIndicesIndexFreed                 //nolint:stylecheck,golint
	Indices_IndexFrozen                []types.EventIndicesIndexFrozen                //nolint:stylecheck,golint
	Offences_Offence                   []types.EventOffencesOffence                   //nolint:stylecheck,golint
	Session_NewSession                 []types.EventSessionNewSession                 //nolint:stylecheck,golint
	Staking_EraPayout                  []types.EventStakingEraPayout                  //nolint:stylecheck,golint
	Staking_Reward                     []types.EventStakingReward                     //nolint:stylecheck,golint
	Staking_Slash                      []types.EventStakingSlash                      //nolint:stylecheck,golint
	Staking_OldSlashingReportDiscarded []types.EventStakingOldSlashingReportDiscarded //nolint:stylecheck,golint
	Staking_StakingElection            []types.EventStakingStakingElection            //nolint:stylecheck,golint
	Staking_SolutionStored             []types.EventStakingSolutionStored             //nolint:stylecheck,golint
	Staking_Bonded                     []types.EventStakingBonded                     //nolint:stylecheck,golint
	Staking_Unbonded                   []types.EventStakingUnbonded                   //nolint:stylecheck,golint
	Staking_Withdrawn                  []types.EventStakingWithdrawn                  //nolint:stylecheck,golint
	System_ExtrinsicSuccess            []types.EventSystemExtrinsicSuccess            //nolint:stylecheck,golint
	System_ExtrinsicFailed             []types.EventSystemExtrinsicFailed             //nolint:stylecheck,golint
	System_CodeUpdated                 []types.EventSystemCodeUpdated                 //nolint:stylecheck,golint
	System_NewAccount                  []types.EventSystemNewAccount                  //nolint:stylecheck,golint
	System_KilledAccount               []types.EventSystemKilledAccount               //nolint:stylecheck,golint
	Assets_Issued                      []types.EventAssetIssued                       //nolint:stylecheck,golint
	Assets_Transferred                 []types.EventAssetTransferred                  //nolint:stylecheck,golint
	Assets_Destroyed                   []types.EventAssetDestroyed                    //nolint:stylecheck,golint
	Democracy_Proposed                 []types.EventDemocracyProposed                 //nolint:stylecheck,golint
	Democracy_Tabled                   []types.EventDemocracyTabled                   //nolint:stylecheck,golint
	Democracy_ExternalTabled           []types.EventDemocracyExternalTabled           //nolint:stylecheck,golint
	Democracy_Started                  []types.EventDemocracyStarted                  //nolint:stylecheck,golint
	Democracy_Passed                   []types.EventDemocracyPassed                   //nolint:stylecheck,golint
	Democracy_NotPassed                []types.EventDemocracyNotPassed                //nolint:stylecheck,golint
	Democracy_Cancelled                []types.EventDemocracyCancelled                //nolint:stylecheck,golint
	Democracy_Executed                 []types.EventDemocracyExecuted                 //nolint:stylecheck,golint
	Democracy_Delegated                []types.EventDemocracyDelegated                //nolint:stylecheck,golint
	Democracy_Undelegated              []types.EventDemocracyUndelegated              //nolint:stylecheck,golint
	Democracy_Vetoed                   []types.EventDemocracyVetoed                   //nolint:stylecheck,golint
	Democracy_PreimageNoted            []types.EventDemocracyPreimageNoted            //nolint:stylecheck,golint
	Democracy_PreimageUsed             []types.EventDemocracyPreimageUsed             //nolint:stylecheck,golint
	Democracy_PreimageInvalid          []types.EventDemocracyPreimageInvalid          //nolint:stylecheck,golint
	Democracy_PreimageMissing          []types.EventDemocracyPreimageMissing          //nolint:stylecheck,golint
	Democracy_PreimageReaped           []types.EventDemocracyPreimageReaped           //nolint:stylecheck,golint
	Democracy_Unlocked                 []types.EventDemocracyUnlocked                 //nolint:stylecheck,golint
	Democracy_Blacklisted              []types.EventDemocracyBlacklisted              //nolint:stylecheck,golint
	Council_Proposed                   []types.EventCollectiveProposed                //nolint:stylecheck,golint
	Council_Voted                      []types.EventCollectiveVoted                   //nolint:stylecheck,golint
	Council_Approved                   []types.EventCollectiveApproved                //nolint:stylecheck,golint
	Council_Disapproved                []types.EventCollectiveDisapproved             //nolint:stylecheck,golint
	Council_Executed                   []types.EventCollectiveExecuted                //nolint:stylecheck,golint
	Council_MemberExecuted             []types.EventCollectiveMemberExecuted          //nolint:stylecheck,golint
	Council_Closed                     []types.EventCollectiveClosed                  //nolint:stylecheck,golint
	TechnicalCommittee_Proposed        []types.EventTechnicalCommitteeProposed        //nolint:stylecheck,golint
	TechnicalCommittee_Voted           []types.EventTechnicalCommitteeVoted           //nolint:stylecheck,golint
	TechnicalCommittee_Approved        []types.EventTechnicalCommitteeApproved        //nolint:stylecheck,golint
	TechnicalCommittee_Disapproved     []types.EventTechnicalCommitteeDisapproved     //nolint:stylecheck,golint
	TechnicalCommittee_Executed        []types.EventTechnicalCommitteeExecuted        //nolint:stylecheck,golint
	TechnicalCommittee_MemberExecuted  []types.EventTechnicalCommitteeMemberExecuted  //nolint:stylecheck,golint
	TechnicalCommittee_Closed          []types.EventTechnicalCommitteeClosed          //nolint:stylecheck,golint
	TechnicalMembership_MemberAdded    []types.EventTechnicalMembershipMemberAdded    //nolint:stylecheck,golint
	TechnicalMembership_MemberRemoved  []types.EventTechnicalMembershipMemberRemoved  //nolint:stylecheck,golint
	TechnicalMembership_MembersSwapped []types.EventTechnicalMembershipMembersSwapped //nolint:stylecheck,golint
	TechnicalMembership_MembersReset   []types.EventTechnicalMembershipMembersReset   //nolint:stylecheck,golint
	TechnicalMembership_KeyChanged     []types.EventTechnicalMembershipKeyChanged     //nolint:stylecheck,golint
	TechnicalMembership_Dummy          []types.EventTechnicalMembershipDummy          //nolint:stylecheck,golint
	Elections_NewTerm                  []types.EventElectionsNewTerm                  //nolint:stylecheck,golint
	Elections_EmptyTerm                []types.EventElectionsEmptyTerm                //nolint:stylecheck,golint
	Elections_ElectionError            []types.EventElectionsElectionError            //nolint:stylecheck,golint
	Elections_MemberKicked             []types.EventElectionsMemberKicked             //nolint:stylecheck,golint
	Elections_MemberRenounced          []types.EventElectionsMemberRenounced          //nolint:stylecheck,golint
	Elections_VoterReported            []types.EventElectionsVoterReported            //nolint:stylecheck,golint
	Identity_IdentitySet               []types.EventIdentitySet                       //nolint:stylecheck,golint
	Identity_IdentityCleared           []types.EventIdentityCleared                   //nolint:stylecheck,golint
	Identity_IdentityKilled            []types.EventIdentityKilled                    //nolint:stylecheck,golint
	Identity_JudgementRequested        []types.EventIdentityJudgementRequested        //nolint:stylecheck,golint
	Identity_JudgementUnrequested      []types.EventIdentityJudgementUnrequested      //nolint:stylecheck,golint
	Identity_JudgementGiven            []types.EventIdentityJudgementGiven            //nolint:stylecheck,golint
	Identity_RegistrarAdded            []types.EventIdentityRegistrarAdded            //nolint:stylecheck,golint
	Identity_SubIdentityAdded          []types.EventIdentitySubIdentityAdded          //nolint:stylecheck,golint
	Identity_SubIdentityRemoved        []types.EventIdentitySubIdentityRemoved        //nolint:stylecheck,golint
	Identity_SubIdentityRevoked        []types.EventIdentitySubIdentityRevoked        //nolint:stylecheck,golint
	Society_Founded                    []types.EventSocietyFounded                    //nolint:stylecheck,golint
	Society_Bid                        []types.EventSocietyBid                        //nolint:stylecheck,golint
	Society_Vouch                      []types.EventSocietyVouch                      //nolint:stylecheck,golint
	Society_AutoUnbid                  []types.EventSocietyAutoUnbid                  //nolint:stylecheck,golint
	Society_Unbid                      []types.EventSocietyUnbid                      //nolint:stylecheck,golint
	Society_Unvouch                    []types.EventSocietyUnvouch                    //nolint:stylecheck,golint
	Society_Inducted                   []types.EventSocietyInducted                   //nolint:stylecheck,golint
	Society_SuspendedMemberJudgement   []types.EventSocietySuspendedMemberJudgement   //nolint:stylecheck,golint
	Society_CandidateSuspended         []types.EventSocietyCandidateSuspended         //nolint:stylecheck,golint
	Society_MemberSuspended            []types.EventSocietyMemberSuspended            //nolint:stylecheck,golint
	Society_Challenged                 []types.EventSocietyChallenged                 //nolint:stylecheck,golint
	Society_Vote                       []types.EventSocietyVote                       //nolint:stylecheck,golint
	Society_DefenderVote               []types.EventSocietyDefenderVote               //nolint:stylecheck,golint
	Society_NewMaxMembers              []types.EventSocietyNewMaxMembers              //nolint:stylecheck,golint
	Society_Unfounded                  []types.EventSocietyUnfounded                  //nolint:stylecheck,golint
	Society_Deposit                    []types.EventSocietyDeposit                    //nolint:stylecheck,golint
	Recovery_RecoveryCreated           []types.EventRecoveryCreated                   //nolint:stylecheck,golint
	Recovery_RecoveryInitiated         []types.EventRecoveryInitiated                 //nolint:stylecheck,golint
	Recovery_RecoveryVouched           []types.EventRecoveryVouched                   //nolint:stylecheck,golint
	Recovery_RecoveryClosed            []types.EventRecoveryClosed                    //nolint:stylecheck,golint
	Recovery_AccountRecovered          []types.EventRecoveryAccountRecovered          //nolint:stylecheck,golint
	Recovery_RecoveryRemoved           []types.EventRecoveryRemoved                   //nolint:stylecheck,golint
	Vesting_VestingUpdated             []types.EventVestingVestingUpdated             //nolint:stylecheck,golint
	Vesting_VestingCompleted           []types.EventVestingVestingCompleted           //nolint:stylecheck,golint
	Scheduler_Scheduled                []types.EventSchedulerScheduled                //nolint:stylecheck,golint
	Scheduler_Canceled                 []types.EventSchedulerCanceled                 //nolint:stylecheck,golint
	Scheduler_Dispatched               []types.EventSchedulerDispatched               //nolint:stylecheck,golint
	Proxy_ProxyExecuted                []types.EventProxyProxyExecuted                //nolint:stylecheck,golint
	Proxy_AnonymousCreated             []types.EventProxyAnonymousCreated             //nolint:stylecheck,golint
	Proxy_Announced                    []types.EventProxyAnnounced                    //nolint:stylecheck,golint
	Sudo_Sudid                         []types.EventSudoSudid                         //nolint:stylecheck,golint
	Sudo_KeyChanged                    []types.EventSudoKeyChanged                    //nolint:stylecheck,golint
	Sudo_SudoAsDone                    []types.EventSudoAsDone                        //nolint:stylecheck,golint
	Treasury_Proposed                  []types.EventTreasuryProposed                  //nolint:stylecheck,golint
	Treasury_Spending                  []types.EventTreasurySpending                  //nolint:stylecheck,golint
	Treasury_Awarded                   []types.EventTreasuryAwarded                   //nolint:stylecheck,golint
	Treasury_Rejected                  []types.EventTreasuryRejected                  //nolint:stylecheck,golint
	Treasury_Burnt                     []types.EventTreasuryBurnt                     //nolint:stylecheck,golint
	Treasury_Rollover                  []types.EventTreasuryRollover                  //nolint:stylecheck,golint
	Treasury_Deposit                   []types.EventTreasuryDeposit                   //nolint:stylecheck,golint
	Treasury_NewTip                    []types.EventTreasuryNewTip                    //nolint:stylecheck,golint
	Treasury_TipClosing                []types.EventTreasuryTipClosing                //nolint:stylecheck,golint
	Treasury_TipClosed                 []types.EventTreasuryTipClosed                 //nolint:stylecheck,golint
	Treasury_TipRetracted              []types.EventTreasuryTipRetracted              //nolint:stylecheck,golint
	Treasury_BountyProposed            []types.EventTreasuryBountyProposed            //nolint:stylecheck,golint
	Treasury_BountyRejected            []types.EventTreasuryBountyRejected            //nolint:stylecheck,golint
	Treasury_BountyBecameActive        []types.EventTreasuryBountyBecameActive        //nolint:stylecheck,golint
	Treasury_BountyAwarded             []types.EventTreasuryBountyAwarded             //nolint:stylecheck,golint
	Treasury_BountyClaimed             []types.EventTreasuryBountyClaimed             //nolint:stylecheck,golint
	Treasury_BountyCanceled            []types.EventTreasuryBountyCanceled            //nolint:stylecheck,golint
	Treasury_BountyExtended            []types.EventTreasuryBountyExtended            //nolint:stylecheck,golint
	Contracts_Instantiated             []types.EventContractsInstantiated             //nolint:stylecheck,golint
	Contracts_Evicted                  []types.EventContractsEvicted                  //nolint:stylecheck,golint
	Contracts_Restored                 []types.EventContractsRestored                 //nolint:stylecheck,golint
	Contracts_CodeStored               []types.EventContractsCodeStored               //nolint:stylecheck,golint
	Contracts_ScheduleUpdated          []types.EventContractsScheduleUpdated          //nolint:stylecheck,golint
	Contracts_ContractExecution        []types.EventContractsContractExecution        //nolint:stylecheck,golint
	Utility_BatchInterrupted           []types.EventUtilityBatchInterrupted           //nolint:stylecheck,golint
	Utility_BatchCompleted             []types.EventUtilityBatchCompleted             //nolint:stylecheck,golint
	Multisig_NewMultisig               []types.EventMultisigNewMultisig               //nolint:stylecheck,golint
	Multisig_MultisigApproval          []types.EventMultisigApproval                  //nolint:stylecheck,golint
	Multisig_MultisigExecuted          []types.EventMultisigExecuted                  //nolint:stylecheck,golint
	Multisig_MultisigCancelled         []types.EventMultisigCancelled                 //nolint:stylecheck,golint
}

// EventBalancesTransfer is emitted when a transfer succeeded (from, to, value)
type EventBalancesWithdraw struct {
	Phase  types.Phase
	From   types.AccountID
	To     types.AccountID
	Value  types.U128
	Topics []types.Hash
}
