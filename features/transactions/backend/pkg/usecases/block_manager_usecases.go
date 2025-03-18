package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"
)

type BlockManagerUseCase interface {
	Start(ctx context.Context, parentWaitGroup *sync.WaitGroup)
}

var _ BlockManagerUseCase = (*DefaultBlockManager)(nil)

type DefaultBlockManager struct {
	chainID                        string
	blockchainNetworkInfraID       string
	blockchainNetworkInfra         BlockchainNetworkInfraAdapterManager
	blockPointerStorage            BlockPointerStorage
	blockStorage                   BlockStorage
	transactionStorage             TransactionStorage
	numberOfBlockProcessorRoutines int
	transactionFetchedQueue        TransactionQueue
}

type DefaultBlockManagerOptions struct {
	ChainID                        string
	BlockchainNetworkInfraID       string
	BlockchainNetworkInfra         BlockchainNetworkInfraAdapterManager
	BlockPointerStorage            BlockPointerStorage
	BlockStorage                   BlockStorage
	TransactionStorage             TransactionStorage
	NumberOfBlockProcessorRoutines *int
	TransactionFetchedQueue        TransactionQueue
}

func ProvideDefaultBlockManager(options DefaultBlockManagerOptions) (*DefaultBlockManager, error) {
	if options.ChainID == "" {
		return nil, errors.New("chainID is required")
	}
	if options.BlockchainNetworkInfraID == "" {
		return nil, errors.New("BlockchainNetworkInfraID is required")
	}
	if options.BlockchainNetworkInfra == nil {
		return nil, errors.New("blockchainNetworkInfra is required")
	}
	if options.BlockPointerStorage == nil {
		return nil, errors.New("blockPointerStorage is required")
	}
	if options.BlockStorage == nil {
		return nil, errors.New("blockStorage is required")
	}
	if options.TransactionStorage == nil {
		return nil, errors.New("transactionStorage is required")
	}
	var numberOfBlockProcessorRoutines int
	if options.NumberOfBlockProcessorRoutines == nil {
		numberOfBlockProcessorRoutines = 1
	} else {
		numberOfBlockProcessorRoutines = *options.NumberOfBlockProcessorRoutines
	}
	if options.TransactionFetchedQueue == nil {
		return nil, errors.New("transactionFetchedQueue is required")
	}
	//transactionFetchedQueue := NewTransactionsBlockingSyncQueue(TransactionFetched)
	return &DefaultBlockManager{
		chainID:                        options.ChainID,
		blockchainNetworkInfraID:       options.BlockchainNetworkInfraID,
		blockchainNetworkInfra:         options.BlockchainNetworkInfra,
		blockPointerStorage:            options.BlockPointerStorage,
		blockStorage:                   options.BlockStorage,
		transactionStorage:             options.TransactionStorage,
		numberOfBlockProcessorRoutines: numberOfBlockProcessorRoutines,
		transactionFetchedQueue:        options.TransactionFetchedQueue,
	}, nil
}

func (d DefaultBlockManager) Start(ctx context.Context, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()
	err := d.initializazeBlockManagerUsecases(ctx)
	if err != nil {
		return
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go d.startBlockOrchestrator(ctx, waitGroup)
	<-ctx.Done()
	waitGroup.Wait()

}

func (d DefaultBlockManager) initializazeBlockManagerUsecases(ctx context.Context) error {
	// 1. Initialize block pointer
	err := d.initializeBlockPointer(ctx)
	maxNumberOfRetries := 9
	secondsBetweenRetries := 1
	for i := 0; i < maxNumberOfRetries; i++ {
		if err == nil {
			break
		}
		t := time.Duration(secondsBetweenRetries) * time.Second
		err := sleep(ctx, t)
		if err != nil {
			return errors.New("could not initialize block pointer due to an error during sleep")
		}
		err = d.initializeBlockPointer(ctx)
	}
	if err != nil {
		return errors.New("could not initialize block pointer")
	}
	return nil
}

func (d DefaultBlockManager) initializeBlockPointer(ctx context.Context) error {
	_, err := d.blockPointerStorage.Get(ctx, d.chainID)
	if err == nil {
		return nil
		// blockpointer not configured
	} else {

		options := BlockStorageListOptions{}
		options.ChainID = d.chainID
		options.State = BlockProcessed
		blockCollection, err := d.blockStorage.AllByDescOrder(ctx, options)
		if err != nil {
			return err
		}
		items := blockCollection.Items
		var blockNumber big.Int
		if len(items) == 0 {
			blockchainNetworlInfraAdapter, err := d.blockchainNetworkInfra.Use(ctx, d.blockchainNetworkInfraID)
			if err != nil {
				return err
			}
			getMostRecentBlockNumberInput := GetMostRecentBlockNumberInput{}
			getMostRecentBlockNumberInput.BlockchainNetworkType = d.chainID
			getMostRecentBlockNumber, err := blockchainNetworlInfraAdapter.GetMostRecentBlockNumber(ctx, getMostRecentBlockNumberInput)
			if err != nil {
				log.Println("MostRecentBlockNumber" + err.Error())
				return err
			}
			blockNumber = getMostRecentBlockNumber.BlockNumber
		} else {
			blockNumber = items[0].BlockNumber
		}
		blockPointer := BlockPointer{
			ID:          d.chainID,
			ChainID:     d.chainID,
			BlockNumber: blockNumber,
		}

		_, err = d.blockPointerStorage.Add(ctx, blockPointer)
		if err != nil {
			return err
		}
		// add metric with it
		//blocksBehindNetwork := blockNumber.Int64() - added.BlockNumber.Int64()
		return nil
	}
}

func (d DefaultBlockManager) startBlockOrchestrator(ctx context.Context, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(d.numberOfBlockProcessorRoutines)
	for currentRoutine := 0; currentRoutine < d.numberOfBlockProcessorRoutines; currentRoutine++ {
		go d.spawnBlockProcessor(ctx, waitGroup)
	}
	<-ctx.Done()
	waitGroup.Wait()
}

func (d DefaultBlockManager) spawnBlockProcessor(ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("blockchain manager shutdown")
			return
		default:
			_ = d.processNextBlock(ctx)
			_ = d.blocksUpdate(ctx)
			sleep(ctx, time.Second)
		}
	}
}

func (d DefaultBlockManager) processNextBlock(ctx context.Context) error {
	nextBlock, err := d.next(ctx)
	if err != nil {
		return err
	}
	if nextBlock != nil {
		log.Printf("Processing next block %s chain:%s\n", nextBlock.BlockNumber.String(), d.chainID)
	}
	if nextBlock != nil {
		err = d.processBlock(ctx, *nextBlock)
		if err != nil {
			return err
		}
	} else {
		//not found new blocks it would have to sleep block time
		//t := time.Duration(int64(eventManager.blockTime)) * time.Second
		t := time.Second
		sleep(ctx, t)
	}
	return nil
}

func (d DefaultBlockManager) next(ctx context.Context) (*Block, error) {
	p, err := d.blockPointerStorage.Get(ctx, d.chainID)
	if err != nil {
		return nil, err
	}
	next := &p.BlockNumber
	assignedBlockWasProcessed := true
	for assignedBlockWasProcessed {
		next = increment(next)
		id := buildBlockDbID(*next, d.chainID)
		nextBlock, err := d.blockStorage.Get(ctx, id)
		//Not found
		if err != nil {
			newBlock, err := d.createBlock(ctx, next)
			if err != nil {
				return nil, err
			}
			return newBlock, nil
		}
		assignedBlockWasProcessed = blockIsProcessed(nextBlock)
		if !assignedBlockWasProcessed {
			return nextBlock, nil
		}
	}
	return nil, nil

}
func (d DefaultBlockManager) createBlock(ctx context.Context, next *big.Int) (*Block, error) {
	blockchainNetworlInfraAdapter, err := d.blockchainNetworkInfra.Use(ctx, d.blockchainNetworkInfraID)
	if err != nil {
		return nil, err
	}
	getMostRecentBlockNumberInput := GetMostRecentBlockNumberInput{}
	getMostRecentBlockNumberInput.BlockchainNetworkType = d.chainID
	getMostRecentBlockNumber, err := blockchainNetworlInfraAdapter.GetMostRecentBlockNumber(ctx, getMostRecentBlockNumberInput)
	if err != nil {
		return nil, err
	}

	if next.Int64() > getMostRecentBlockNumber.BlockNumber.Int64() {
		return nil, nil
	}

	nextBlock := &Block{
		ID:          buildBlockDbID(*next, d.chainID),
		BlockNumber: *next,
		State:       BlockFetched,
		ChainID:     d.chainID,
	}

	addedBlock, err := d.blockStorage.Add(ctx, *nextBlock)
	if err != nil {
		return nil, err
	}
	return addedBlock, nil
}
func (d DefaultBlockManager) processBlock(ctx context.Context, block Block) error {
	storedBlock, err := d.blockStorage.Get(ctx, block.ID)
	if err != nil {
		return err
	}
	if storedBlock.State != BlockFetched {
		return nil
	}
	blockchainNetworlInfraAdapter, err := d.blockchainNetworkInfra.Use(ctx, d.blockchainNetworkInfraID)
	if err != nil {
		return err
	}
	getBlockByNumberInput := GetBlockByNumberInput{}
	getBlockByNumberInput.BlockchainNetworkType = d.chainID
	getBlockByNumberInput.BlockNumber = storedBlock.BlockNumber
	getBlockNumberOutput, err := blockchainNetworlInfraAdapter.GetBlockByNumber(ctx, getBlockByNumberInput)
	if err != nil {
		log.Println(storedBlock.BlockNumber.String() + ":" + err.Error())

		return err
	}
	if len(getBlockNumberOutput.TransactionHashes) > 0 {
		transactionInsertBatch := d.transactionStorage.CreateInsertBatch()
		for _, transactionId := range getBlockNumberOutput.TransactionHashes {
			transaction := Transaction{}
			transaction.ID = transactionId
			transaction.ChainID = d.chainID
			transaction.State = TransactionFetched
			transaction.BlockNumber = block.BlockNumber
			transaction.CreationDate = time.Now()
			transaction.LastUpdated = time.Now()
			_, err := transactionInsertBatch.Add(ctx, transaction)
			if err != nil {
				return err
			}

		}
		flushFailure := transactionInsertBatch.Flush(ctx)
		if flushFailure != nil {
			return flushFailure
		}
		for _, transactionId := range getBlockNumberOutput.TransactionHashes {
			transaction := Transaction{}
			transaction.ID = transactionId
			transaction.ChainID = d.chainID
			transaction.State = TransactionFetched
			transaction.BlockNumber = block.BlockNumber
			transaction.CreationDate = time.Now()
			transaction.LastUpdated = time.Now()
			_, err := transactionInsertBatch.Add(ctx, transaction)
			if err != nil {
				return err
			}
			transactionSent := d.transactionFetchedQueue.add(transaction)
			if !transactionSent {
				return errors.New("transaction not sent")
			}
		}
	}
	block.State = BlockProcessed
	_, err = d.blockStorage.Edit(ctx, block)
	if err != nil {
		return err
	}
	return nil
}
func (d DefaultBlockManager) blocksUpdate(ctx context.Context) error {
	from, getBlockPointerFailure := d.blockPointerStorage.Get(ctx, d.chainID)
	if getBlockPointerFailure != nil {
		return getBlockPointerFailure
	}

	to := from.BlockNumber

	options := BlockStorageListByAscFromBlockNumberOptions{}
	options.ChainID = d.chainID
	options.BlockNumber = to

	blockCollection, allBlocksFailure := d.blockStorage.AllByAscFromBlockNumber(ctx, options)
	if allBlocksFailure != nil {
		return allBlocksFailure
	}

	for i := 0; i < len(blockCollection.Items); i++ {
		if !blockIsProcessed(&blockCollection.Items[i]) || to.Int64()+1 < blockCollection.Items[i].BlockNumber.Int64() {
			break
		}
		to = blockCollection.Items[i].BlockNumber
	}
	toPosition := BlockPointer{
		ID:          d.chainID,
		ChainID:     d.chainID,
		BlockNumber: to,
	}

	_, editBlockPointerFailure := d.blockPointerStorage.Edit(ctx, toPosition)
	if editBlockPointerFailure != nil {
		return editBlockPointerFailure
	}

	return nil
}

func blockIsProcessed(block *Block) bool {
	return block != nil && (block.State == BlockProcessed)
}
func sleep(ctx context.Context, t time.Duration) error {
	timeoutchan := make(chan bool)
	go func() {
		<-time.After(t)
		timeoutchan <- true
	}()

	select {
	case <-timeoutchan:
		break
	case <-ctx.Done():
		return errors.New("terminated sleep due to a context cancellation")
	}
	return nil
}

func increment(value *big.Int) *big.Int {
	return value.Add(value, big.NewInt(int64(1)))
}

func buildBlockDbID(blockNumber big.Int, chainID string) string {
	return fmt.Sprintf("%s-%s", blockNumber.String(), chainID)
}
