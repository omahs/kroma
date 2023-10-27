package e2e

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/kroma-network/kroma/bindings/bindings"
	"github.com/kroma-network/kroma/bindings/predeploys"
	"github.com/kroma-network/kroma/components/node/client"
	"github.com/kroma-network/kroma/components/node/eth"
	"github.com/kroma-network/kroma/components/node/metrics"
	rollupNode "github.com/kroma-network/kroma/components/node/node"
	"github.com/kroma-network/kroma/components/node/p2p"
	"github.com/kroma-network/kroma/components/node/rollup"
	"github.com/kroma-network/kroma/components/node/rollup/derive"
	"github.com/kroma-network/kroma/components/node/rollup/driver"
	"github.com/kroma-network/kroma/components/node/sources"
	"github.com/kroma-network/kroma/components/node/testlog"
	"github.com/kroma-network/kroma/components/node/withdrawals"
	val "github.com/kroma-network/kroma/components/validator"
	chal "github.com/kroma-network/kroma/components/validator/challenge"
	"github.com/kroma-network/kroma/e2e/testdata"
	kpprof "github.com/kroma-network/kroma/utils/service/pprof"
)

var enableParallelTesting bool = true

// Init testing to enable test flags
var _ = func() bool {
	testing.Init()
	return true
}()

var verboseGethNodes bool

func init() {
	flag.BoolVar(&verboseGethNodes, "gethlogs", true, "Enable logs on geth nodes")
	flag.Parse()
	if os.Getenv("E2E_DISABLE_PARALLEL") == "true" {
		enableParallelTesting = false
	}
}

func parallel(t *testing.T) {
	t.Helper()
	if enableParallelTesting {
		t.Parallel()
	}
}

func TestL2OutputSubmitter(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	cfg.NonFinalizedOutputs = true // speed up the time till we see checkpoint outputs

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]

	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.Nil(t, err)
	rollupClient := sources.NewRollupClient(client.NewBaseRPCClient(rollupRPCClient))

	// OutputOracle is already deployed
	l2OutputOracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
	require.Nil(t, err)

	initialOutputBlockNumber, err := l2OutputOracle.LatestBlockNumber(&bind.CallOpts{})
	require.Nil(t, err)

	// Wait until the second output submission from L2. The output submitter submits outputs from the
	// unsafe portion of the chain which gets reorged on startup. The sequencer has an out of date view
	// when it creates it's first block and uses and old L1 Origin. It then does not submit a batch
	// for that block and subsequently reorgs to match what the syncer derives when running the
	// reconciliation process.
	l2Sync := sys.Clients["syncer"]
	_, err = waitForL2Block(big.NewInt(6), l2Sync, 10*time.Duration(cfg.DeployConfig.L2BlockTime)*time.Second)
	require.Nil(t, err)

	// Wait for batch submitter to update L2 output oracle.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		l2ooBlockNumber, err := l2OutputOracle.LatestBlockNumber(&bind.CallOpts{})
		require.Nil(t, err)

		// Wait for the L2 output oracle to have been changed from the initial
		// timestamp set in the contract constructor.
		if l2ooBlockNumber.Cmp(initialOutputBlockNumber) > 0 {
			// Retrieve the l2 output committed at this updated timestamp.
			committedL2Output, err := l2OutputOracle.GetL2OutputAfter(&bind.CallOpts{}, l2ooBlockNumber)
			require.NotEqual(t, [32]byte{}, committedL2Output.OutputRoot, "Empty L2 Output")
			require.Nil(t, err)

			// Fetch the corresponding L2 block and assert the committed L2
			// output matches the block's state root.
			//
			// NOTE: This assertion will change once the L2 output format is
			// finalized.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			l2Output, err := rollupClient.OutputAtBlock(ctx, l2ooBlockNumber.Uint64())
			require.Nil(t, err)
			require.Equal(t, l2Output.OutputRoot[:], committedL2Output.OutputRoot[:])
			break
		}

		select {
		case <-ctx.Done():
			t.Fatalf("State root oracle not updated")
		case <-ticker.C:
		}
	}
}

func TestValidationReward(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	cfg.DeployConfig.FinalizationPeriodSeconds = 32
	cfg.DeployConfig.L2OutputOracleSubmissionInterval = 16
	cfg.DeployConfig.ColosseumSegmentsLengths = "5,5"
	cfg.DeployConfig.ValidatorPoolRoundDuration = 16

	sys, err := cfg.Start()
	require.NoError(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	validatorVault, err := bindings.NewValidatorRewardVault(predeploys.ValidatorRewardVaultAddr, l2Sync)
	require.NoError(t, err)

	rewardDivider, err := validatorVault.REWARDDIVIDER(&bind.CallOpts{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, rewardDivider.Uint64(), uint64(1))

	// Send a transaction to pay a fee.
	_, err = cfg.SendTransferTx(l2Seq, l2Sync)
	require.NoError(t, err)

	l2RewardedCh := make(chan *bindings.ValidatorRewardVaultRewarded, 1)
	rewardedSub, err := validatorVault.WatchRewarded(&bind.WatchOpts{}, l2RewardedCh, nil, nil)
	require.NoError(t, err)
	defer rewardedSub.Unsubscribe()

	timeout := time.Minute * 2
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case evt := <-l2RewardedCh:
			vaultBalance, err := l2Sync.PendingBalanceAt(ctx, predeploys.ValidatorRewardVaultAddr)
			require.NoError(t, err)
			reward := new(big.Int).Div(vaultBalance, rewardDivider)
			require.Equal(t, 0, reward.Cmp(evt.Amount))
			return
		case <-ctx.Done():
			t.Fatalf("not rewarded to validator")
		}
	}
}

// TestSystemE2E sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that L1 deposits are reflected on L2.
// All nodes are run in process (but are the full nodes, not mocked or stubbed).
func TestSystemE2E(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	log := testlog.Logger(t, log.LvlInfo)
	log.Info("genesis", "l2", sys.RollupConfig.Genesis.L2, "l1", sys.RollupConfig.Genesis.L1, "l2_time", sys.RollupConfig.Genesis.L2Time)

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Transactor Account
	ethPrivKey := sys.cfg.Secrets.Alice

	// Send Transaction & wait for success
	fromAddr := sys.cfg.Secrets.Addresses().Alice

	// Find deposit contract
	depositContract, err := bindings.NewKromaPortal(predeploys.DevKromaPortalAddr, l1Client)
	require.Nil(t, err)

	// Create signer
	opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, cfg.L1ChainIDBig())
	require.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Finally send TX
	mintAmount := big.NewInt(1_000_000_000_000)
	opts.Value = mintAmount
	tx, err := depositContract.DepositTransaction(opts, fromAddr, common.Big0, 1_000_000, false, nil)
	require.Nil(t, err, "with deposit tx")

	receipt, err := waitForTransaction(tx.Hash(), l1Client, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for deposit tx on L1")

	reconstructedDep, err := derive.UnmarshalDepositLogEvent(receipt.Logs[0])
	require.NoError(t, err, "Could not reconstruct L2 Deposit")
	tx = types.NewTx(reconstructedDep)
	receipt, err = waitForL2Transaction(tx.Hash(), l2Sync, 6*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.NoError(t, err)
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

	// Confirm balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, mintAmount, diff, "Did not get expected balance change")

	// Submit TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx = types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     1, // Already have deposit
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.Nil(t, err, "Sending L2 tx to sequencer")

	_, err = waitForL2Transaction(tx.Hash(), l2Seq, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on sequencer")

	receipt, err = waitForL2Transaction(tx.Hash(), l2Sync, 10*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on syncer")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "TX should have succeeded")

	// Verify blocks match after batch submission on syncers and sequencers
	syncBlock, err := l2Sync.BlockByNumber(context.Background(), receipt.BlockNumber)
	require.Nil(t, err)
	seqBlock, err := l2Seq.BlockByNumber(context.Background(), receipt.BlockNumber)
	require.Nil(t, err)
	require.Equal(t, syncBlock.NumberU64(), seqBlock.NumberU64(), "Syncer and sequencer blocks not the same after including a batch tx")
	require.Equal(t, syncBlock.ParentHash(), seqBlock.ParentHash(), "Syncer and sequencer blocks parent hashes not the same after including a batch tx")
	require.Equal(t, syncBlock.Hash(), seqBlock.Hash(), "Syncer and sequencer blocks not the same after including a batch tx")

	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.Nil(t, err)
	rollupClient := sources.NewRollupClient(client.NewBaseRPCClient(rollupRPCClient))
	// basic check that sync status works
	seqStatus, err := rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.LessOrEqual(t, seqBlock.NumberU64(), seqStatus.UnsafeL2.Number)
	// basic check that version endpoint works
	seqVersion, err := rollupClient.Version(context.Background())
	require.Nil(t, err)
	require.NotEqual(t, "", seqVersion)
}

// TestConfirmationDepth runs the rollup with both sequencer and syncer not immediately processing the tip of the chain.
func TestConfirmationDepth(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	cfg.DeployConfig.SequencerWindowSize = 4
	cfg.DeployConfig.MaxSequencerDrift = 10 * cfg.DeployConfig.L1BlockTime
	seqConfDepth := uint64(2)
	syncConfDepth := uint64(5)
	cfg.Nodes["sequencer"].Driver.SequencerConfDepth = seqConfDepth
	cfg.Nodes["sequencer"].Driver.SyncerConfDepth = 0
	cfg.Nodes["syncer"].Driver.SyncerConfDepth = syncConfDepth

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	log := testlog.Logger(t, log.LvlInfo)
	log.Info("genesis", "l2", sys.RollupConfig.Genesis.L2, "l1", sys.RollupConfig.Genesis.L1, "l2_time", sys.RollupConfig.Genesis.L2Time)

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Wait enough time for the sequencer to submit a block with distance from L1 head, submit it,
	// and for the slower syncer to read a full sequence window and cover confirmation depth for reading and some margin
	<-time.After(time.Duration((cfg.DeployConfig.SequencerWindowSize+syncConfDepth+3)*cfg.DeployConfig.L1BlockTime) * time.Second)

	// within a second, get both L1 and L2 syncer and sequencer block heads
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l1Head, err := l1Client.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	l2SeqHead, err := l2Seq.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	l2SyncHead, err := l2Sync.BlockByNumber(ctx, nil)
	require.NoError(t, err)

	seqInfo, err := derive.L1InfoDepositTxData(l2SeqHead.Transactions()[0].Data())
	require.NoError(t, err)
	require.LessOrEqual(t, seqInfo.Number+seqConfDepth, l1Head.NumberU64(), "the seq L2 head block should have an origin older than the L1 head block by at least the sequencer conf depth")

	syncInfo, err := derive.L1InfoDepositTxData(l2SyncHead.Transactions()[0].Data())
	require.NoError(t, err)
	require.LessOrEqual(t, syncInfo.Number+syncConfDepth, l1Head.NumberU64(), "the syncer L2 head block should have an origin older than the L1 head block by at least the syncer conf depth")
}

// TestPendingGasLimit tests the configuration of the gas limit of the pending block,
// and if it does not conflict with the regular gas limit on the syncer or sequencer.
func TestPendingGasLimit(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)

	// configure the L2 gas limit to be high, and the pending gas limits to be lower for resource saving.
	cfg.DeployConfig.L2GenesisBlockGasLimit = 30_000_000
	cfg.GethOptions["sequencer"] = []GethOption{
		func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
			ethCfg.Miner.GasCeil = 10_000_000
			return nil
		},
	}
	cfg.GethOptions["syncer"] = []GethOption{
		func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
			ethCfg.Miner.GasCeil = 9_000_000
			return nil
		},
	}

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	log := testlog.Logger(t, log.LvlInfo)
	log.Info("genesis", "l2", sys.RollupConfig.Genesis.L2, "l1", sys.RollupConfig.Genesis.L1, "l2_time", sys.RollupConfig.Genesis.L2Time)

	l2Sync := sys.Clients["syncer"]
	l2Seq := sys.Clients["sequencer"]

	checkGasLimit := func(client *ethclient.Client, number *big.Int, expected uint64) *types.Header {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		header, err := client.HeaderByNumber(ctx, number)
		cancel()
		require.NoError(t, err)
		require.Equal(t, expected, header.GasLimit)
		return header
	}

	// check if the gaslimits are matching the expected values,
	// and that the syncer/sequencer can use their locally configured gas limit for the pending block.
	for {
		checkGasLimit(l2Seq, big.NewInt(-1), 10_000_000)
		checkGasLimit(l2Sync, big.NewInt(-1), 9_000_000)
		checkGasLimit(l2Seq, nil, 30_000_000)
		latestSyncHeader := checkGasLimit(l2Sync, nil, 30_000_000)

		// Stop once the syncer passes genesis:
		// this implies we checked a new block from the sequencer, on both sequencer and syncer nodes.
		if latestSyncHeader.Number.Uint64() > 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// TestFinalize tests if L2 finalizes after sufficient time after L1 finalizes
func TestFinalize(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]

	// as configured in the extra geth lifecycle in testing setup
	const finalizedDistance = 8
	// Wait enough time for L1 to finalize and L2 to confirm its data in finalized L1 blocks
	timeout := time.Duration((finalizedDistance+6)*cfg.DeployConfig.L1BlockTime) * time.Second * timeoutMultiplier
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// fetch the finalizes head of geth
	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "timeout")
			return
		default:
			<-time.After(time.Second)
		}

		// poll until the finalized block number is greater than 0
		l2Finalized, err := waitForL2Block(big.NewInt(int64(rpc.FinalizedBlockNumber)), l2Seq, time.Second)
		require.NoError(t, err)
		if l2Finalized.NumberU64() > 0 {
			break
		}
	}
}

func TestMintOnRevertedDeposit(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}
	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]
	l2Sync := sys.Clients["syncer"]

	// Find deposit contract
	depositContract, err := bindings.NewKromaPortal(predeploys.DevKromaPortalAddr, l1Client)
	require.Nil(t, err)
	l1Node := sys.Nodes["l1"]

	// create signer
	ks := l1Node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, ks.Accounts()[0], cfg.L1ChainIDBig())
	require.Nil(t, err)
	fromAddr := opts.From

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	startBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	cancel()
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	startNonce, err := l2Sync.NonceAt(ctx, fromAddr, nil)
	require.NoError(t, err)
	cancel()

	toAddr := common.Address{0xff, 0xff}
	mintAmount := big.NewInt(9_000_000)
	opts.Value = mintAmount
	value := new(big.Int).Mul(common.Big2, startBalance) // trigger a revert by transferring more than we have available
	tx, err := depositContract.DepositTransaction(opts, toAddr, value, 1_000_000, false, nil)
	require.Nil(t, err, "with deposit tx")

	receipt, err := waitForTransaction(tx.Hash(), l1Client, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for deposit tx on L1")

	reconstructedDep, err := derive.UnmarshalDepositLogEvent(receipt.Logs[0])
	require.NoError(t, err, "Could not reconstruct L2 Deposit")
	tx = types.NewTx(reconstructedDep)
	receipt, err = waitForL2Transaction(tx.Hash(), l2Sync, 10*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.NoError(t, err)
	require.Equal(t, receipt.Status, types.ReceiptStatusFailed)

	// Confirm balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	endBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	cancel()
	require.Nil(t, err)
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	toAddrBalance, err := l2Sync.BalanceAt(ctx, toAddr, nil)
	require.NoError(t, err)
	cancel()

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, mintAmount, diff, "Did not get expected balance change")
	require.Equal(t, common.Big0.Int64(), toAddrBalance.Int64(), "The recipient account balance should be zero")

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	endNonce, err := l2Sync.NonceAt(ctx, fromAddr, nil)
	require.NoError(t, err)
	cancel()
	require.Equal(t, startNonce+1, endNonce, "Nonce of deposit sender should increment on L2, even if the deposit fails")
}

func TestMissingBatchE2E(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}
	// Note this test zeroes the balance of the batch-submitter to make the batches unable to go into L1.
	// The test logs may look scary, but this is expected:
	// 'unable to publish transaction    role=batcher   err="insufficient funds for gas * price + value"'

	cfg := DefaultSystemConfig(t)
	// small sequence window size so the test does not take as long
	cfg.DeployConfig.SequencerWindowSize = 4

	// Specifically set batch submitter balance to stop batches from being included
	cfg.Premine[cfg.Secrets.Addresses().Batcher] = big.NewInt(0)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     0,
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.Nil(t, err, "Sending L2 tx to sequencer")

	// Let it show up on the unsafe chain
	receipt, err := waitForL2Transaction(tx.Hash(), l2Seq, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on sequencer")

	// Wait until the block it was first included in shows up in the safe chain on the syncer
	_, err = waitForL2Block(receipt.BlockNumber, l2Sync, time.Duration((sys.RollupConfig.SeqWindowSize+4)*cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for block on syncer")

	// Assert that the transaction is not found on the syncer
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = l2Sync.TransactionReceipt(ctx, tx.Hash())
	require.Equal(t, ethereum.NotFound, err, "Found transaction in syncer when it should not have been included")

	// Wait a short time for the L2 reorg to occur on the sequencer as well.
	// The proper thing to do is to wait until the sequencer marks this block safe.
	<-time.After(2 * time.Second)

	// Assert that the reconciliation process did an L2 reorg on the sequencer to remove the invalid block
	ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	block, err := l2Seq.BlockByNumber(ctx2, receipt.BlockNumber)
	require.Nil(t, err, "Get block from sequencer")
	require.NotEqual(t, block.Hash(), receipt.BlockHash, "L2 Sequencer did not reorg out transaction on it's safe chain")
}

func L1InfoFromState(ctx context.Context, contract *bindings.L1Block, l2Number *big.Int) (derive.L1BlockInfo, error) {
	var err error
	var out derive.L1BlockInfo
	opts := bind.CallOpts{
		BlockNumber: l2Number,
		Context:     ctx,
	}

	out.Number, err = contract.Number(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get number: %w", err)
	}

	out.Time, err = contract.Timestamp(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get timestamp: %w", err)
	}

	out.BaseFee, err = contract.Basefee(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get timestamp: %w", err)
	}

	blockHashBytes, err := contract.Hash(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get block hash: %w", err)
	}
	out.BlockHash = common.BytesToHash(blockHashBytes[:])

	out.SequenceNumber, err = contract.SequenceNumber(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get sequence number: %w", err)
	}

	overhead, err := contract.L1FeeOverhead(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get l1 fee overhead: %w", err)
	}
	out.L1FeeOverhead = eth.Bytes32(common.BigToHash(overhead))

	scalar, err := contract.L1FeeScalar(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get l1 fee scalar: %w", err)
	}
	out.L1FeeScalar = eth.Bytes32(common.BigToHash(scalar))

	batcherHash, err := contract.BatcherHash(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get batch sender: %w", err)
	}
	out.BatcherAddr = common.BytesToAddress(batcherHash[:])

	validatorRewardScalar, err := contract.ValidatorRewardScalar(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get validator reward scalar: %w", err)
	}
	out.ValidatorRewardScalar = eth.Bytes32(common.BigToHash(validatorRewardScalar))

	return out, nil
}

// TestSystemMockP2P sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that
// the nodes can sync L2 blocks before they are confirmed on L1.
func TestSystemMockP2P(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	// Disable batcher, so we don't sync from L1
	cfg.DisableBatcher = true
	// disable at the start, so we don't miss any gossiped blocks.
	cfg.Nodes["sequencer"].Driver.SequencerStopped = true

	// connect the nodes
	cfg.P2PTopology = map[string][]string{
		"syncer": {"sequencer"},
	}

	var published, received []common.Hash
	seqTracer, syncTracer := new(FnTracer), new(FnTracer)
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayload) {
		published = append(published, payload.BlockHash)
	}
	syncTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
		received = append(received, payload.BlockHash)
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["syncer"].Tracer = syncTracer

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	// Enable the sequencer now that everyone is ready to receive payloads.
	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.NoError(t, err)
	require.NoError(t, rollupRPCClient.Call(nil, "admin_startProposer", sys.L2GenesisCfg.ToBlock().Hash()))

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     0,
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.Nil(t, err, "Sending L2 tx to sequencer")

	// Wait for tx to be mined on the L2 sequencer chain
	receiptSeq, err := waitForL2Transaction(tx.Hash(), l2Seq, 10*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on sequencer")

	// Wait until the block it was first included in shows up in the safe chain on the syncer
	receiptSync, err := waitForL2Transaction(tx.Hash(), l2Sync, 10*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on syncer")

	require.Equal(t, receiptSeq, receiptSync)

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received))
	require.ElementsMatch(t, received, published[:len(received)])

	// Verify that the tx was received via p2p
	require.Contains(t, received, receiptSync.BlockHash)
}

// TestSystemRPCAltSync sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that
// the nodes can sync L2 blocks before they are confirmed on L1.
//
// Test steps:
// 1. Spin up the nodes (P2P is disabled on the syncer)
// 2. Send a transaction to the sequencer.
// 3. Wait for the TX to be mined on the sequencer chain.
// 5. Wait for the syncer to detect a gap in the payload queue vs. the unsafe head
// 6. Wait for the RPC sync method to grab the block from the sequencer over RPC and insert it into the syncer's unsafe chain.
// 7. Wait for the syncer to sync the unsafe chain into the safe chain.
// 8. Verify that the TX is included in the syncer's safe chain.
func TestSystemRPCAltSync(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	// the default is nil, but this may change in the future.
	// This test must ensure the blocks are not synced via Gossip, but instead via the alt RPC based sync.
	cfg.P2PTopology = nil
	// Disable batcher, so there will not be any L1 data to sync from
	cfg.DisableBatcher = true

	var published, received []string
	seqTracer, syncTracer := new(FnTracer), new(FnTracer)
	// The sequencer still publishes the blocks to the tracer, even if they do not reach the network due to disabled P2P
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayload) {
		published = append(published, payload.ID().String())
	}
	// Blocks are now received via the RPC based alt-sync method
	syncTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
		received = append(received, payload.ID().String())
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["syncer"].Tracer = syncTracer

	sys, err := cfg.Start(SystemConfigOption{
		key:  "afterRollupNodeStart",
		role: "sequencer",
		action: func(sCfg *SystemConfig, system *System) {
			rpc, _ := system.Nodes["sequencer"].Attach() // never errors
			cfg.Nodes["syncer"].L2Sync = &rollupNode.PreparedL2SyncEndpoint{
				Client: client.NewBaseRPCClient(rpc),
			}
		},
	})
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit a TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     0,
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.Nil(t, err, "Sending L2 tx to sequencer")

	// Wait for tx to be mined on the L2 sequencer chain
	receiptSeq, err := waitForTransaction(tx.Hash(), l2Seq, 6*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on sequencer")

	// Wait for alt RPC sync to pick up the blocks on the sequencer chain
	receiptSync, err := waitForTransaction(tx.Hash(), l2Sync, 12*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on syncer")

	require.Equal(t, receiptSeq, receiptSync)

	// Verify that the tx was received via RPC sync (P2P is disabled)
	require.Contains(t, received, eth.BlockID{Hash: receiptSync.BlockHash, Number: receiptSync.BlockNumber.Uint64()}.String())

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received))
	require.ElementsMatch(t, received, published[:len(received)])
}

func TestSystemP2PAltSync(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)

	// remove default syncer node
	delete(cfg.Nodes, "syncer")
	// Add more syncer nodes
	cfg.Nodes["alice"] = &rollupNode.Config{
		Driver: driver.Config{
			SyncerConfDepth:    0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Nodes["bob"] = &rollupNode.Config{
		Driver: driver.Config{
			SyncerConfDepth:    0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Loggers["alice"] = testlog.Logger(t, log.LvlInfo).New("role", "alice")
	cfg.Loggers["bob"] = testlog.Logger(t, log.LvlInfo).New("role", "bob")

	// connect the nodes
	cfg.P2PTopology = map[string][]string{
		"sequencer": {"alice", "bob"},
		"alice":     {"sequencer", "bob"},
		"bob":       {"alice", "sequencer"},
	}
	// Enable the P2P req-resp based sync
	cfg.P2PReqRespSync = true

	// Disable batcher, so there will not be any L1 data to sync from
	cfg.DisableBatcher = true

	var published []string
	seqTracer := new(FnTracer)
	// The sequencer still publishes the blocks to the tracer, even if they do not reach the network due to disabled P2P
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayload) {
		published = append(published, payload.ID().String())
	}
	// Blocks are now received via the RPC based alt-sync method
	cfg.Nodes["sequencer"].Tracer = seqTracer

	sys, err := cfg.Start()
	require.NoError(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit a TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     0,
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.NoError(t, err, "Sending L2 tx to sequencer")

	// Wait for tx to be mined on the L2 sequencer chain
	receiptSeq, err := waitForTransaction(tx.Hash(), l2Seq, 6*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.NoError(t, err, "Waiting for L2 tx on sequencer")

	// Gossip is able to respond to IWANT messages for the duration of heartbeat_time * message_window = 0.5 * 12 = 6
	// Wait till we pass that, and then we'll have missed some blocks that cannot be retrieved in any way from gossip
	time.Sleep(time.Second * 10)

	// set up our syncer node, connect it to alice/bob
	cfg.Loggers["syncer"] = testlog.Logger(t, log.LvlInfo).New("role", "syncer")
	snapLog := log.New()
	snapLog.SetHandler(log.DiscardHandler())

	// Create a peer, and hook up alice and bob
	h, err := sys.Mocknet.GenPeer()
	require.NoError(t, err)
	_, err = sys.Mocknet.LinkPeers(sys.RollupNodes["alice"].P2P().Host().ID(), h.ID())
	require.NoError(t, err)
	_, err = sys.Mocknet.LinkPeers(sys.RollupNodes["bob"].P2P().Host().ID(), h.ID())
	require.NoError(t, err)

	// Configure the new rollup node that'll be syncing
	var syncedPayloads []string
	syncNodeCfg := &rollupNode.Config{
		L2Sync:    &rollupNode.PreparedL2SyncEndpoint{Client: nil},
		Driver:    driver.Config{SyncerConfDepth: 0},
		Rollup:    *sys.RollupConfig,
		P2PSigner: nil,
		RPC: rollupNode.RPCConfig{
			ListenAddr:  "127.0.0.1",
			ListenPort:  0,
			EnableAdmin: true,
		},
		P2P:                 &p2p.Prepared{HostP2P: h, EnableReqRespSync: true},
		Metrics:             rollupNode.MetricsConfig{Enabled: false}, // no metrics server
		Pprof:               kpprof.CLIConfig{},
		L1EpochPollInterval: time.Second * 10,
		Tracer: &FnTracer{
			OnUnsafeL2PayloadFn: func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
				syncedPayloads = append(syncedPayloads, payload.ID().String())
			},
		},
	}
	configureL1(syncNodeCfg, sys.Nodes["l1"])
	syncerL2Engine, _, err := initL2Geth("syncer", big.NewInt(int64(cfg.DeployConfig.L2ChainID)), sys.L2GenesisCfg, cfg.JWTFilePath)
	require.NoError(t, err)
	require.NoError(t, syncerL2Engine.Start())

	configureL2(syncNodeCfg, syncerL2Engine, cfg.JWTSecret)

	syncerNode, err := rollupNode.New(context.Background(), syncNodeCfg, cfg.Loggers["syncer"], snapLog, "", metrics.NewMetrics(""))
	require.NoError(t, err)
	err = syncerNode.Start(context.Background())
	require.NoError(t, err)

	// connect alice and bob to our new syncer node
	_, err = sys.Mocknet.ConnectPeers(sys.RollupNodes["alice"].P2P().Host().ID(), syncerNode.P2P().Host().ID())
	require.NoError(t, err)
	_, err = sys.Mocknet.ConnectPeers(sys.RollupNodes["bob"].P2P().Host().ID(), syncerNode.P2P().Host().ID())
	require.NoError(t, err)

	rpc, err := syncerL2Engine.Attach()
	require.NoError(t, err)
	l2Sync := ethclient.NewClient(rpc)

	// It may take a while to sync, but eventually we should see the sequenced data show up
	receiptSync, err := waitForTransaction(tx.Hash(), l2Sync, 100*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.NoError(t, err, "Waiting for L2 tx on syncer")

	require.Equal(t, receiptSeq, receiptSync)

	// Verify that the tx was received via P2P sync
	require.Contains(t, syncedPayloads, eth.BlockID{Hash: receiptSync.BlockHash, Number: receiptSync.BlockNumber.Uint64()}.String())

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(syncedPayloads))
	require.ElementsMatch(t, syncedPayloads, published[:len(syncedPayloads)])
}

// TestSystemDenseTopology sets up a dense p2p topology with 3 syncer nodes and 1 sequencer node.
func TestSystemDenseTopology(t *testing.T) {
	t.Skip("Skipping dense topology test to avoid flakiness.")

	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	// slow down L1 blocks, so we can see the L2 blocks arrive well before the L1 blocks do.
	// Keep the seq window small so the L2 chain is started quick
	cfg.DeployConfig.L1BlockTime = 10

	// Append additional nodes to the system to construct a dense p2p network
	cfg.Nodes["syncer2"] = &rollupNode.Config{
		Driver: driver.Config{
			SyncerConfDepth:    0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Nodes["syncer3"] = &rollupNode.Config{
		Driver: driver.Config{
			SyncerConfDepth:    0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Loggers["syncer2"] = testlog.Logger(t, log.LvlInfo).New("role", "syncer")
	cfg.Loggers["syncer3"] = testlog.Logger(t, log.LvlInfo).New("role", "syncer")

	// connect the nodes
	cfg.P2PTopology = map[string][]string{
		"syncer":  {"sequencer", "syncer2", "syncer3"},
		"syncer2": {"sequencer", "syncer", "syncer3"},
		"syncer3": {"sequencer", "syncer", "syncer2"},
	}

	// Set peer scoring for each node, but without banning
	for _, node := range cfg.Nodes {
		params, err := p2p.GetPeerScoreParams("light", 2)
		require.NoError(t, err)
		node.P2P = &p2p.Config{
			PeerScoring:    params,
			BanningEnabled: false,
		}
	}

	var published, received1, received2, received3 []common.Hash
	seqTracer, syncTracer, syncTracer2, syncTracer3 := new(FnTracer), new(FnTracer), new(FnTracer), new(FnTracer)
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayload) {
		published = append(published, payload.BlockHash)
	}
	syncTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
		received1 = append(received1, payload.BlockHash)
	}
	syncTracer2.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
		received2 = append(received2, payload.BlockHash)
	}
	syncTracer3.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
		received3 = append(received3, payload.BlockHash)
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["syncer"].Tracer = syncTracer
	cfg.Nodes["syncer2"].Tracer = syncTracer2
	cfg.Nodes["syncer3"].Tracer = syncTracer3

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]
	l2Sync2 := sys.Clients["syncer2"]
	l2Sync3 := sys.Clients["syncer3"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     0,
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.NoError(t, err, "Sending L2 tx to sequencer")

	// Wait for tx to be mined on the L2 sequencer chain
	receiptSeq, err := waitForTransaction(tx.Hash(), l2Seq, 10*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.NoError(t, err, "Waiting for L2 tx on sequencer")

	// Wait until the block it was first included in shows up in the safe chain on the syncer
	receiptSync, err := waitForTransaction(tx.Hash(), l2Sync, 10*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.NoError(t, err, "Waiting for L2 tx on syncer")
	require.Equal(t, receiptSeq, receiptSync)

	receiptSync, err = waitForTransaction(tx.Hash(), l2Sync2, 10*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.NoError(t, err, "Waiting for L2 tx on syncer2")
	require.Equal(t, receiptSeq, receiptSync)

	receiptSync, err = waitForTransaction(tx.Hash(), l2Sync3, 10*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.NoError(t, err, "Waiting for L2 tx on syncer3")
	require.Equal(t, receiptSeq, receiptSync)

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received1))
	require.GreaterOrEqual(t, len(published), len(received2))
	require.GreaterOrEqual(t, len(published), len(received3))
	require.ElementsMatch(t, published, received1[:len(published)])
	require.ElementsMatch(t, published, received2[:len(published)])
	require.ElementsMatch(t, published, received3[:len(published)])

	// Verify that the tx was received via p2p
	require.Contains(t, received1, receiptSync.BlockHash)
	require.Contains(t, received2, receiptSync.BlockHash)
	require.Contains(t, received3, receiptSync.BlockHash)
}

func TestL1InfoContract(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	endSyncBlockNumber := big.NewInt(4)
	endSeqBlockNumber := big.NewInt(6)
	endSyncBlock, err := waitForL2Block(endSyncBlockNumber, l2Sync, time.Minute)
	require.Nil(t, err)
	endSeqBlock, err := waitForL2Block(endSeqBlockNumber, l2Seq, time.Minute)
	require.Nil(t, err)

	seqL1Info, err := bindings.NewL1Block(cfg.L1InfoPredeployAddress, l2Seq)
	require.Nil(t, err)

	syncL1Info, err := bindings.NewL1Block(cfg.L1InfoPredeployAddress, l2Sync)
	require.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fillInfoLists := func(start *types.Block, contract *bindings.L1Block, client *ethclient.Client) ([]derive.L1BlockInfo, []derive.L1BlockInfo) {
		var txList, stateList []derive.L1BlockInfo
		for b := start; ; {
			var infoFromTx derive.L1BlockInfo
			require.NoError(t, infoFromTx.UnmarshalBinary(b.Transactions()[0].Data()))
			txList = append(txList, infoFromTx)

			infoFromState, err := L1InfoFromState(ctx, contract, b.Number())
			require.Nil(t, err)
			stateList = append(stateList, infoFromState)

			// Genesis L2 block contains no L1 Deposit TX
			if b.NumberU64() == 1 {
				return txList, stateList
			}
			b, err = client.BlockByHash(ctx, b.ParentHash())
			require.Nil(t, err)
		}
	}

	l1InfosFromSequencerTransactions, l1InfosFromSequencerState := fillInfoLists(endSeqBlock, seqL1Info, l2Seq)
	l1InfosFromSyncerTransactions, l1InfosFromSyncerState := fillInfoLists(endSyncBlock, syncL1Info, l2Sync)

	l1blocks := make(map[common.Hash]derive.L1BlockInfo)
	maxL1Hash := l1InfosFromSequencerTransactions[0].BlockHash
	for h := maxL1Hash; ; {
		b, err := l1Client.BlockByHash(ctx, h)
		require.Nil(t, err)

		l1blocks[h] = derive.L1BlockInfo{
			Number:                b.NumberU64(),
			Time:                  b.Time(),
			BaseFee:               b.BaseFee(),
			BlockHash:             h,
			SequenceNumber:        0, // ignored, will be overwritten
			BatcherAddr:           sys.RollupConfig.Genesis.SystemConfig.BatcherAddr,
			L1FeeOverhead:         sys.RollupConfig.Genesis.SystemConfig.Overhead,
			L1FeeScalar:           sys.RollupConfig.Genesis.SystemConfig.Scalar,
			ValidatorRewardScalar: sys.RollupConfig.Genesis.SystemConfig.ValidatorRewardScalar,
		}

		h = b.ParentHash()
		if b.NumberU64() == 0 {
			break
		}
	}

	checkInfoList := func(name string, list []derive.L1BlockInfo) {
		for _, info := range list {
			if expected, ok := l1blocks[info.BlockHash]; ok {
				expected.SequenceNumber = info.SequenceNumber // the seq nr is not part of the L1 info we know in advance, so we ignore it.
				require.Equal(t, expected, info)
			} else {
				t.Fatalf("Did not find block hash for L1 Info: %v in test %s", info, name)
			}
		}
	}

	checkInfoList("On sequencer with tx", l1InfosFromSequencerTransactions)
	checkInfoList("On sequencer with state", l1InfosFromSequencerState)
	checkInfoList("On syncer with tx", l1InfosFromSyncerTransactions)
	checkInfoList("On syncer with state", l1InfosFromSyncerState)
}

// calcGasFees determines the actual cost of the transaction given a specific basefee
// This does not include the L1 data fee charged from L2 transactions.
func calcGasFees(gasUsed uint64, gasTipCap *big.Int, gasFeeCap *big.Int, baseFee *big.Int) *big.Int {
	x := new(big.Int).Add(gasTipCap, baseFee)
	// If tip + basefee > gas fee cap, clamp it to the gas fee cap
	if x.Cmp(gasFeeCap) > 0 {
		x = gasFeeCap
	}
	return x.Mul(x, new(big.Int).SetUint64(gasUsed))
}

// calcL1GasUsed returns the gas used to include the transaction data in
// the calldata on L1
func calcL1GasUsed(data []byte, overhead *big.Int) *big.Int {
	var zeroes, ones uint64
	for _, byt := range data {
		if byt == 0 {
			zeroes++
		} else {
			ones++
		}
	}

	zeroesGas := zeroes * 4 // params.TxDataZeroGas
	onesGas := ones * 16    // params.TxDataNonZeroGasEIP2028
	l1Gas := new(big.Int).SetUint64(zeroesGas + onesGas)
	return new(big.Int).Add(l1Gas, overhead)
}

// TestWithdrawals checks that a deposit and then withdrawal execution succeeds. It verifies the
// balance changes on L1 and L2 and has to include gas fees in the balance checks.
// It does not check that the withdrawal can be executed prior to the end of the finality period.
func TestWithdrawals(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	cfg.DeployConfig.FinalizationPeriodSeconds = 2 // 2s finalization period

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice
	fromAddr := crypto.PubkeyToAddress(ethPrivKey.PublicKey)

	// Find deposit contract
	depositContract, err := bindings.NewKromaPortal(predeploys.DevKromaPortalAddr, l1Client)
	require.Nil(t, err)

	// Create L1 signer
	opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, cfg.L1ChainIDBig())
	require.Nil(t, err)

	// Start L2 balance
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Finally send TX
	mintAmount := big.NewInt(1_000_000_000_000)
	opts.Value = mintAmount
	tx, err := depositContract.DepositTransaction(opts, fromAddr, common.Big0, 1_000_000, false, nil)
	require.Nil(t, err, "with deposit tx")

	receipt, err := waitForTransaction(tx.Hash(), l1Client, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for deposit tx on L1")

	// Bind L2 Withdrawer Contract
	l2withdrawer, err := bindings.NewL2ToL1MessagePasser(predeploys.L2ToL1MessagePasserAddr, l2Seq)
	require.Nil(t, err, "binding withdrawer on L2")

	// Wait for deposit to arrive
	reconstructedDep, err := derive.UnmarshalDepositLogEvent(receipt.Logs[0])
	require.NoError(t, err, "Could not reconstruct L2 Deposit")
	tx = types.NewTx(reconstructedDep)
	receipt, err = waitForL2Transaction(tx.Hash(), l2Sync, 10*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.NoError(t, err)
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

	// Confirm L2 balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, mintAmount, diff, "Did not get expected balance change after mint")

	// Start L2 balance for withdrawal
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err = l2Seq.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Initiate Withdrawal
	withdrawAmount := big.NewInt(500_000_000_000)
	l2opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, cfg.L2ChainIDBig())
	require.Nil(t, err)
	l2opts.Value = withdrawAmount
	tx, err = l2withdrawer.InitiateWithdrawal(l2opts, fromAddr, big.NewInt(21000), nil)
	require.Nil(t, err, "sending initiate withdraw tx")

	receipt, err = waitForL2Transaction(tx.Hash(), l2Sync, 10*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "withdrawal initiated on L2 sequencer")
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction failed")

	// Verify L2 balance after withdrawal
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err := l2Sync.HeaderByNumber(ctx, receipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err = l2Sync.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Take fee into account
	diff = new(big.Int).Sub(startBalance, endBalance)
	fees := calcGasFees(receipt.GasUsed, tx.GasTipCap(), tx.GasFeeCap(), header.BaseFee)
	fees = fees.Add(fees, receipt.L1Fee)
	diff = diff.Sub(diff, fees)
	require.Equal(t, withdrawAmount, diff)

	// Take start balance on L1
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err = l1Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Get l2BlockNumber for proof generation
	ctx, cancel = context.WithTimeout(context.Background(), 40*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second*timeoutMultiplier)
	defer cancel()
	blockNumber, err := withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevKromaPortalAddr, receipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err = l2Sync.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	nextHeader, err := l2Sync.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber+1))
	require.Nil(t, err)

	rpcClient, err := rpc.Dial(sys.Nodes["syncer"].WSEndpoint())
	require.Nil(t, err)
	proofCl := gethclient.New(rpcClient)
	receiptCl := ethclient.NewClient(rpcClient)

	// Now create withdrawal
	oracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
	require.Nil(t, err)

	version := rollup.L2OutputRootVersion(sys.RollupConfig, header.Time)
	params, err := withdrawals.ProveWithdrawalParameters(context.Background(), version, proofCl, receiptCl, tx.Hash(), header, nextHeader, oracle)
	require.Nil(t, err)

	portal, err := bindings.NewKromaPortal(predeploys.DevKromaPortalAddr, l1Client)
	require.Nil(t, err)

	opts.Value = nil

	// Prove withdrawal
	tx, err = portal.ProveWithdrawalTransaction(
		opts,
		bindings.TypesWithdrawalTransaction{
			Nonce:    params.Nonce,
			Sender:   params.Sender,
			Target:   params.Target,
			Value:    params.Value,
			GasLimit: params.GasLimit,
			Data:     params.Data,
		},
		params.L2OutputIndex,
		params.OutputRootProof,
		params.WithdrawalProof,
	)
	require.Nil(t, err)

	// Ensure that our withdrawal was proved successfully
	proveReceipt, err := waitForTransaction(tx.Hash(), l1Client, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "prove withdrawal")
	require.Equal(t, types.ReceiptStatusSuccessful, proveReceipt.Status)

	// Wait for finalization and then create the Finalized Withdrawal Transaction
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	defer cancel()
	_, err = withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevKromaPortalAddr, header.Number)
	require.Nil(t, err)

	// Finalize withdrawal
	tx, err = portal.FinalizeWithdrawalTransaction(
		opts,
		bindings.TypesWithdrawalTransaction{
			Nonce:    params.Nonce,
			Sender:   params.Sender,
			Target:   params.Target,
			Value:    params.Value,
			GasLimit: params.GasLimit,
			Data:     params.Data,
		},
	)
	require.Nil(t, err)

	// Ensure that our withdrawal was finalized successfully
	finalizeReceipt, err := waitForTransaction(tx.Hash(), l1Client, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "finalize withdrawal")
	require.Equal(t, types.ReceiptStatusSuccessful, finalizeReceipt.Status)

	// Verify balance after withdrawal
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err = l1Client.HeaderByNumber(ctx, finalizeReceipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err = l1Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Ensure that withdrawal - gas fees are added to the L1 balance
	// Fun fact, the fee is greater than the withdrawal amount
	// NOTE: The gas fees include *both* the ProveWithdrawalTransaction and FinalizeWithdrawalTransaction transactions.
	diff = new(big.Int).Sub(endBalance, startBalance)
	fees = calcGasFees(proveReceipt.GasUsed+finalizeReceipt.GasUsed, tx.GasTipCap(), tx.GasFeeCap(), header.BaseFee)
	withdrawAmount = withdrawAmount.Sub(withdrawAmount, fees)
	require.Equal(t, withdrawAmount, diff)
}

// TestFees checks that L1/L2 fees are handled.
func TestFees(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	// TODO: after we have the system config contract and new kroma-geth L1 cost utils,
	// we can pull in l1 costs into every e2e test and account for it in assertions easily etc.
	cfg.DeployConfig.GasPriceOracleOverhead = 2100
	cfg.DeployConfig.GasPriceOracleScalar = 1000_000

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice
	fromAddr := crypto.PubkeyToAddress(ethPrivKey.PublicKey)

	// Find gaspriceoracle contract
	gpoContract, err := bindings.NewGasPriceOracle(predeploys.GasPriceOracleAddr, l2Seq)
	require.Nil(t, err)

	overhead, err := gpoContract.Overhead(&bind.CallOpts{})
	require.Nil(t, err, "reading gpo overhead")
	decimals, err := gpoContract.DECIMALS(&bind.CallOpts{})
	require.Nil(t, err, "reading gpo decimals")
	scalar, err := gpoContract.Scalar(&bind.CallOpts{})
	require.Nil(t, err, "reading gpo scalar")

	require.Equal(t, overhead.Uint64(), uint64(2100), "wrong gpo overhead")
	require.Equal(t, decimals.Uint64(), uint64(6), "wrong gpo decimals")
	require.Equal(t, scalar.Uint64(), uint64(1_000_000), "wrong gpo scalar")

	// Check balances of ProtocolVault
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	protocolVaultStartBalance, err := l2Seq.BalanceAt(ctx, predeploys.ProtocolVaultAddr, nil)
	require.Nil(t, err)

	// Check balances of L1FeeVault
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l1FeeVaultStartBalance, err := l2Seq.BalanceAt(ctx, predeploys.L1FeeVaultAddr, nil)
	require.Nil(t, err)

	// Simple transfer from signer to random account
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err := l2Sync.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	toAddr := common.Address{0xff, 0xff}
	transferAmount := big.NewInt(1_000_000_000)
	gasTip := big.NewInt(10)
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
		ChainID:   cfg.L2ChainIDBig(),
		Nonce:     0,
		To:        &toAddr,
		Value:     transferAmount,
		GasTipCap: gasTip,
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	sender, err := types.LatestSignerForChainID(cfg.L2ChainIDBig()).Sender(tx)
	require.NoError(t, err)
	t.Logf("waiting for tx %s from %s to %s", tx.Hash(), sender, tx.To())
	err = l2Seq.SendTransaction(context.Background(), tx)
	require.Nil(t, err, "Sending L2 tx to sequencer")

	_, err = waitForL2Transaction(tx.Hash(), l2Seq, 4*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on sequencer")

	receipt, err := waitForL2Transaction(tx.Hash(), l2Sync, 4*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on syncer")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "TX should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err := l2Seq.HeaderByNumber(ctx, receipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	validatorRewardVaultStartBalance, err := l2Seq.BalanceAt(ctx, predeploys.ValidatorRewardVaultAddr, safeAddBig(header.Number, big.NewInt(-1)))
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	validatorRewardVaultEndBalance, err := l2Seq.BalanceAt(ctx, predeploys.ValidatorRewardVaultAddr, header.Number)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := l2Seq.BalanceAt(ctx, fromAddr, header.Number)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	protocolVaultEndBalance, err := l2Seq.BalanceAt(ctx, predeploys.ProtocolVaultAddr, header.Number)
	require.Nil(t, err)

	l1Header, err := sys.Clients["l1"].HeaderByNumber(ctx, nil)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l1FeeVaultEndBalance, err := l2Seq.BalanceAt(ctx, predeploys.L1FeeVaultAddr, header.Number)
	require.Nil(t, err)

	// Diff fee recipients balances
	protocolVaultDiff := new(big.Int).Sub(protocolVaultEndBalance, protocolVaultStartBalance)
	l1FeeVaultDiff := new(big.Int).Sub(l1FeeVaultEndBalance, l1FeeVaultStartBalance)
	validatorRewardVaultDiff := new(big.Int).Sub(validatorRewardVaultEndBalance, validatorRewardVaultStartBalance)

	// get a validator reward scalar from L1Block contract
	l1BlockContract, err := bindings.NewL1Block(predeploys.L1BlockAddr, l2Seq)
	require.Nil(t, err)

	validatorRewardScalar, err := l1BlockContract.ValidatorRewardScalar(&bind.CallOpts{})
	require.Nil(t, err, "reading validatorRewardScalar")

	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	fee := new(big.Int)
	fee.Mul(gasUsed, header.BaseFee)
	fee.Add(fee, new(big.Int).Mul(gasTip, gasUsed))

	R := big.NewRat(validatorRewardScalar.Int64(), 10000)
	reward := new(big.Int).Mul(fee, R.Num())
	reward.Div(reward, R.Denom())

	// Tally Validator reward
	require.Equal(t, reward, validatorRewardVaultDiff, "validator reward mismatch")

	// Tally Protocol fund
	protocolFee := new(big.Int).Sub(fee, reward)
	require.Equal(t, protocolFee.Cmp(protocolVaultDiff), 0, "protocol fund mismatch")

	// Tally sequencer reward
	bytes, err := tx.MarshalBinary()
	require.Nil(t, err)
	l1GasUsed := calcL1GasUsed(bytes, overhead)
	divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)
	l1Fee := new(big.Int).Mul(l1GasUsed, l1Header.BaseFee)
	l1Fee = l1Fee.Mul(l1Fee, scalar)
	l1Fee = l1Fee.Div(l1Fee, divisor)

	require.Equal(t, l1Fee, l1FeeVaultDiff, "sequencer reward mismatch")

	// Tally Sequencer reward against GasPriceOracle
	gpoL1Fee, err := gpoContract.GetL1Fee(&bind.CallOpts{}, bytes)
	require.Nil(t, err)
	require.Equal(t, l1Fee, gpoL1Fee, "sequencer reward mismatch")

	// Calculate total fee
	protocolVaultDiff.Add(protocolVaultDiff, validatorRewardVaultDiff)
	totalFee := new(big.Int).Add(protocolVaultDiff, l1FeeVaultDiff)
	balanceDiff := new(big.Int).Sub(startBalance, endBalance)
	balanceDiff.Sub(balanceDiff, transferAmount)
	require.Equal(t, balanceDiff, totalFee, "balances should add up")
}

func TestStopStartSequencer(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	rollupNode := sys.RollupNodes["sequencer"]

	nodeRPC, err := rpc.DialContext(context.Background(), rollupNode.HTTPEndpoint())
	require.Nil(t, err, "Error dialing node")

	blockBefore := latestBlock(t, l2Seq)
	time.Sleep(time.Duration(cfg.DeployConfig.L2BlockTime+1) * time.Second)
	blockAfter := latestBlock(t, l2Seq)
	require.Greaterf(t, blockAfter, blockBefore, "Chain did not advance")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	blockHash := common.Hash{}
	err = nodeRPC.CallContext(ctx, &blockHash, "admin_stopSequencer")
	require.Nil(t, err, "Error stopping sequencer")

	blockBefore = latestBlock(t, l2Seq)
	time.Sleep(time.Duration(cfg.DeployConfig.L2BlockTime+1) * time.Second)
	blockAfter = latestBlock(t, l2Seq)
	require.Equal(t, blockAfter, blockBefore, "Chain advanced after stopping sequencer")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = nodeRPC.CallContext(ctx, nil, "admin_startSequencer", blockHash)
	require.Nil(t, err, "Error starting sequencer")

	blockBefore = latestBlock(t, l2Seq)
	time.Sleep(time.Duration(cfg.DeployConfig.L2BlockTime+1) * time.Second)
	blockAfter = latestBlock(t, l2Seq)
	require.Greater(t, blockAfter, blockBefore, "Chain did not advance after starting sequencer")
}

func TestStopStartBatcher(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["syncer"].HTTPEndpoint())
	require.Nil(t, err)
	rollupClient := sources.NewRollupClient(client.NewBaseRPCClient(rollupRPCClient))

	l2Seq := sys.Clients["sequencer"]
	l2Sync := sys.Clients["syncer"]

	// retrieve the initial sync status
	seqStatus, err := rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)

	nonce := uint64(0)
	sendTx := func() *types.Receipt {
		// Submit TX to L2 sequencer node
		tx := types.MustSignNewTx(cfg.Secrets.Alice, types.LatestSignerForChainID(cfg.L2ChainIDBig()), &types.DynamicFeeTx{
			ChainID:   cfg.L2ChainIDBig(),
			Nonce:     nonce,
			To:        &common.Address{0xff, 0xff},
			Value:     big.NewInt(1_000_000_000),
			GasTipCap: big.NewInt(10),
			GasFeeCap: big.NewInt(200),
			Gas:       21000,
		})
		nonce++
		err = l2Seq.SendTransaction(context.Background(), tx)
		require.Nil(t, err, "Sending L2 tx to sequencer")

		// Let it show up on the unsafe chain
		receipt, err := waitForTransaction(tx.Hash(), l2Seq, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
		require.Nil(t, err, "Waiting for L2 tx on sequencer")

		return receipt
	}
	// send a transaction
	receipt := sendTx()

	// wait until the block the tx was first included in shows up in the safe chain on the syncer
	safeBlockInclusionDuration := time.Duration(3*cfg.DeployConfig.L1BlockTime) * time.Second
	_, err = waitForL2Block(receipt.BlockNumber, l2Sync, safeBlockInclusionDuration)
	require.Nil(t, err, "Waiting for block on syncer")

	// ensure the safe chain advances
	newSeqStatus, err := rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain did not advance")

	// stop the batch submission
	err = sys.Batcher.Stop(context.Background())
	require.NoError(t, err)

	// wait for any old safe blocks being submitted / derived
	time.Sleep(safeBlockInclusionDuration)

	// get the initial sync status
	seqStatus, err = rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)

	// send another tx
	sendTx()
	time.Sleep(safeBlockInclusionDuration)

	// ensure that the safe chain does not advance while the batcher is stopped
	newSeqStatus, err = rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.Equal(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain advanced while batcher was stopped")

	// start the batch submission
	err = sys.Batcher.Start()
	require.NoError(t, err)
	time.Sleep(safeBlockInclusionDuration)

	// send a third tx
	receipt = sendTx()

	// wait until the block the tx was first included in shows up in the safe chain on the syncer
	_, err = waitForL2Block(receipt.BlockNumber, l2Sync, safeBlockInclusionDuration)
	require.Nil(t, err, "Waiting for block on syncer")

	// ensure that the safe chain advances after restarting the batcher
	newSeqStatus, err = rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain did not advance after batcher was restarted")
}

func TestChallenge(t *testing.T) {
	parallel(t)
	if !verboseGethNodes {
		log.Root().SetHandler(log.DiscardHandler())
	}

	cfg := DefaultSystemConfig(t)
	cfg.EnableMaliciousValidator = true
	cfg.EnableGuardian = true

	sys, err := cfg.Start()
	require.NoError(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]

	// deposit to ValidatorPool to be a challenger
	err = cfg.DepositValidatorPool(l1Client, cfg.Secrets.Challenger1, big.NewInt(1_000_000_000))
	require.NoError(t, err, "Error challenger deposit to ValidatorPool")

	// OutputOracle is already deployed
	l2OutputOracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
	require.NoError(t, err)

	// Colosseum is already deployed
	colosseum, err := bindings.NewColosseumCaller(predeploys.DevColosseumAddr, l1Client)
	require.NoError(t, err)

	// SecurityCouncil is already deployed
	securityCouncil, err := bindings.NewSecurityCouncilCaller(predeploys.DevSecurityCouncilAddr, l1Client)
	require.NoError(t, err)

	targetOutputOracleIndex := uint64(math.Ceil(float64(testdata.TargetBlockNumber) / float64(cfg.DeployConfig.L2OutputOracleSubmissionInterval)))

	// set a timeout for one cycle of challenge
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		select {
		case <-ctx.Done():
			t.Fatalf("Timed out for challenge test")
		default:
			challengeStatus, err := colosseum.GetStatus(&bind.CallOpts{}, new(big.Int).SetUint64(targetOutputOracleIndex), cfg.Secrets.Addresses().Challenger1)
			require.NoError(t, err)

			// wait until status becomes READY_TO_PROVE
			if challengeStatus != chal.StatusReadyToProve {
				continue
			}
		}
		break
	}
	cancel()

	// set a timeout for security council to validate output
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for ; ; <-ticker.C {
		select {
		case <-ctx.Done():
			// after security council timed out, the challenge is regarded to be correct
			return
		default:
			challengeStatus, err := colosseum.GetStatus(&bind.CallOpts{}, new(big.Int).SetUint64(targetOutputOracleIndex), cfg.Secrets.Addresses().Challenger1)
			require.NoError(t, err)

			// after challenge is proven, status is NONE
			if challengeStatus == chal.StatusNone {
				// check tx not executed
				tx, err := securityCouncil.Transactions(&bind.CallOpts{}, new(big.Int).SetUint64(0))
				require.NoError(t, err)
				require.False(t, tx.Executed)

				// check output is deleted by challenger
				output, err := l2OutputOracle.GetL2Output(&bind.CallOpts{}, new(big.Int).SetUint64(targetOutputOracleIndex))
				require.NoError(t, err)
				require.Equal(t, output.Submitter, cfg.Secrets.Addresses().Challenger1)
				outputDeleted := val.IsOutputDeleted(output.OutputRoot)
				require.True(t, outputDeleted)
			}
		}
	}
}

func safeAddBig(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

func latestBlock(t *testing.T, client *ethclient.Client) uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	blockAfter, err := client.BlockNumber(ctx)
	require.Nil(t, err, "Error getting latest block")
	return blockAfter
}
