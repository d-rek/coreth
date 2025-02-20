// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package coreth

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/coreth/consensus/dummy"
	"github.com/ava-labs/coreth/core"
	"github.com/ava-labs/coreth/core/state"
	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/eth"
	"github.com/ava-labs/coreth/miner"
	"github.com/ava-labs/coreth/node"
	"github.com/ava-labs/coreth/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
)

var (
	BlackholeAddr = common.Address{
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)

type Tx = types.Transaction
type Block = types.Block
type Hash = common.Hash

type ETHChain struct {
	backend *eth.Ethereum
	cb      *dummy.ConsensusCallbacks
	mcb     *miner.MinerCallbacks
}

// NewETHChain creates an Ethereum blockchain with the given configs.
func NewETHChain(config *eth.Config, nodecfg *node.Config, chainDB ethdb.Database, settings eth.Settings, initGenesis bool) *ETHChain {
	if config == nil {
		config = &eth.DefaultConfig
	}
	if nodecfg == nil {
		nodecfg = &node.Config{}
	}
	//mux := new(event.TypeMux)
	node, err := node.New(nodecfg)
	if err != nil {
		panic(err)
	}
	//if ep != "" {
	//	log.Info(fmt.Sprintf("temporary keystore = %s", ep))
	//}
	cb := new(dummy.ConsensusCallbacks)
	mcb := new(miner.MinerCallbacks)
	backend, err := eth.New(node, config, cb, mcb, chainDB, settings, initGenesis)
	if err != nil {
		panic(fmt.Sprintf("failed to create new eth backend due to %s", err))
	}
	chain := &ETHChain{backend: backend, cb: cb, mcb: mcb}
	backend.SetEtherbase(BlackholeAddr)
	return chain
}

func (self *ETHChain) Start() {
	self.backend.StartMining()
	self.backend.Start()
}

func (self *ETHChain) Stop() {
	self.backend.Stop()
}

func (self *ETHChain) GenBlock() {
	self.backend.Miner().GenBlock()
}

func (self *ETHChain) SubscribeNewMinedBlockEvent() *event.TypeMuxSubscription {
	return self.backend.Miner().GetWorkerMux().Subscribe(core.NewMinedBlockEvent{})
}

func (self *ETHChain) BlockChain() *core.BlockChain {
	return self.backend.BlockChain()
}

func (self *ETHChain) APIBackend() *eth.EthAPIBackend {
	return self.backend.APIBackend
}

func (self *ETHChain) UnlockIndexing() {
	self.backend.BlockChain().UnlockIndexing()
}

func (self *ETHChain) PendingSize() (int, error) {
	pending, err := self.backend.TxPool().Pending()
	count := 0
	for _, txs := range pending {
		count += len(txs)
	}
	return count, err
}

func (self *ETHChain) AddRemoteTxs(txs []*types.Transaction) []error {
	return self.backend.TxPool().AddRemotes(txs)
}

func (self *ETHChain) AddLocalTxs(txs []*types.Transaction) []error {
	return self.backend.TxPool().AddLocals(txs)
}

func (self *ETHChain) SetOnSeal(cb func(*types.Block) error) {
	self.cb.OnSeal = cb
}

func (self *ETHChain) SetOnSealHash(cb func(*types.Header)) {
	self.cb.OnSealHash = cb
}

func (self *ETHChain) SetOnSealFinish(cb func(*types.Block) error) {
	self.mcb.OnSealFinish = cb
}

func (self *ETHChain) SetOnSealDrop(cb func(*types.Block)) {
	self.mcb.OnSealDrop = cb
}

func (self *ETHChain) SetOnAPIs(cb dummy.OnAPIsCallbackType) {
	self.cb.OnAPIs = cb
}

func (self *ETHChain) SetOnFinalize(cb dummy.OnFinalizeCallbackType) {
	self.cb.OnFinalize = cb
}

func (self *ETHChain) SetOnFinalizeAndAssemble(cb dummy.OnFinalizeAndAssembleCallbackType) {
	self.cb.OnFinalizeAndAssemble = cb
}

func (self *ETHChain) SetOnExtraStateChange(cb dummy.OnExtraStateChangeType) {
	self.cb.OnExtraStateChange = cb
}

// Returns a new mutable state based on the current HEAD block.
func (self *ETHChain) CurrentState() (*state.StateDB, error) {
	return self.backend.BlockChain().State()
}

// Returns a new mutable state based on the given block.
func (self *ETHChain) BlockState(block *types.Block) (*state.StateDB, error) {
	return self.backend.BlockChain().StateAt(block.Root())
}

// Retrives a block from the database by hash.
func (self *ETHChain) GetBlockByHash(hash common.Hash) *types.Block {
	return self.backend.BlockChain().GetBlockByHash(hash)
}

// Retrives a block from the database by number.
func (self *ETHChain) GetBlockByNumber(num uint64) *types.Block {
	return self.backend.BlockChain().GetBlockByNumber(num)
}

// Validate the canonical chain from current block to the genesis.
// This should only be called as a convenience method in tests, not
// in production as it traverses the entire chain.
func (self *ETHChain) ValidateCanonicalChain() error {
	return self.backend.BlockChain().ValidateCanonicalChain()
}

// WriteCanonicalFromCurrentBlock writes the canonical chain from the
// current block to the genesis.
func (self *ETHChain) WriteCanonicalFromCurrentBlock(toBlock *types.Block) error {
	return self.backend.BlockChain().WriteCanonicalFromCurrentBlock(toBlock)
}

// SetPreference sets the current head block to the one provided as an argument
// regardless of what the chain contents were prior.
func (self *ETHChain) SetPreference(block *types.Block) error {
	return self.BlockChain().SetPreference(block)
}

// Accept sets a minimum height at which no reorg can pass. Additionally,
// this function may trigger a reorg if the block being accepted is not in the
// canonical chain.
func (self *ETHChain) Accept(block *types.Block) error {
	return self.BlockChain().Accept(block)
}

// LastAcceptedBlock returns the last block to be marked as accepted.
func (self *ETHChain) LastAcceptedBlock() *types.Block {
	return self.BlockChain().LastAcceptedBlock()
}

func (self *ETHChain) GetReceiptsByHash(hash common.Hash) types.Receipts {
	return self.backend.BlockChain().GetReceiptsByHash(hash)
}

func (self *ETHChain) GetGenesisBlock() *types.Block {
	return self.backend.BlockChain().Genesis()
}

func (self *ETHChain) InsertChain(chain []*types.Block) (int, error) {
	return self.backend.BlockChain().InsertChain(chain)
}

func (self *ETHChain) NewRPCHandler(maximumDuration time.Duration) *rpc.Server {
	return rpc.NewServer(maximumDuration)
}

func (self *ETHChain) AttachEthService(handler *rpc.Server, namespaces []string) {
	nsmap := make(map[string]bool)
	for _, ns := range namespaces {
		nsmap[ns] = true
	}
	for _, api := range self.backend.APIs() {
		if nsmap[api.Namespace] {
			handler.RegisterName(api.Namespace, api.Service)
		}
	}
}

func (self *ETHChain) GetTxSubmitCh() <-chan core.NewTxsEvent {
	newTxsChan := make(chan core.NewTxsEvent)
	self.backend.TxPool().SubscribeNewTxsEvent(newTxsChan)
	return newTxsChan
}

func (self *ETHChain) GetTxPool() *core.TxPool {
	return self.backend.TxPool()
}

// SetGasPrice sets the gas price on the backend
func (self *ETHChain) SetGasPrice(newGasPrice *big.Int) {
	self.backend.SetGasPrice(newGasPrice)
}
