package usecases

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type TransactionManagerUseCase interface {
	Start(ctx context.Context, parentWaitGroup *sync.WaitGroup)
}

var _ TransactionManagerUseCase = (*DefaultTransactionManager)(nil)

type DefaultTransactionManager struct {
	chainID                              string
	blockchainNetworkInfraID             string
	blockchainNetworkInfra               BlockchainNetworkInfraAdapterManager
	numberOfTransactionProcessorRoutines int
	transactionFetchedQueue              TransactionQueue
	transactionStorage                   TransactionStorage
	kafkaProducer                        KafkaProducer
	userAddressesStorage                 UserAddressesStorage
	userStorage                          UserStorage
}

type DefaultTransactionManagerOptions struct {
	ChainID                              string
	BlockchainNetworkInfraID             string
	BlockchainNetworkInfra               BlockchainNetworkInfraAdapterManager
	NumberOfTransactionProcessorRoutines *int
	TransactionFetchedQueue              TransactionQueue
	TransactionStorage                   TransactionStorage
	KafkaProducer                        KafkaProducer
	UserAddressesStorage                 UserAddressesStorage
	UserStorage                          UserStorage
}

func ProvideDefaultTransactionManagerUseCase(options DefaultTransactionManagerOptions) (*DefaultTransactionManager, error) {
	if options.BlockchainNetworkInfraID == "" {
		return nil, errors.New("blockchain network infra id is required")
	}
	if options.ChainID == "" {
		return nil, errors.New("chain id is required")
	}
	if options.BlockchainNetworkInfra == nil {
		return nil, errors.New("blockchain network infra is required")
	}

	var numberOfTransactionProcessorRoutines int
	if options.NumberOfTransactionProcessorRoutines == nil {
		numberOfTransactionProcessorRoutines = 1
	} else {
		numberOfTransactionProcessorRoutines = *options.NumberOfTransactionProcessorRoutines
	}
	if options.TransactionFetchedQueue == nil {
		return nil, errors.New("transaction queue is required")
	}
	if options.TransactionStorage == nil {
		return nil, errors.New("transaction storage is required")
	}
	if options.KafkaProducer == nil {
		return nil, errors.New("kafka producer is required")
	}
	if options.UserAddressesStorage == nil {
		return nil, errors.New("user addresses storage is required")
	}
	if options.UserStorage == nil {
		return nil, errors.New("user storage is required")
	}
	return &DefaultTransactionManager{
		chainID:                              options.ChainID,
		blockchainNetworkInfraID:             options.BlockchainNetworkInfraID,
		blockchainNetworkInfra:               options.BlockchainNetworkInfra,
		numberOfTransactionProcessorRoutines: numberOfTransactionProcessorRoutines,
		transactionFetchedQueue:              options.TransactionFetchedQueue,
		transactionStorage:                   options.TransactionStorage,
		kafkaProducer:                        options.KafkaProducer,
		userAddressesStorage:                 options.UserAddressesStorage,
		userStorage:                          options.UserStorage,
	}, nil
}
func (d DefaultTransactionManager) Start(ctx context.Context, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go d.startTransactionProcess(ctx, waitGroup)
	waitGroup.Add(1)
	go d.startTransactionOrphanCollectorProcess(ctx, waitGroup)

	<-ctx.Done()
	waitGroup.Wait()
}

func (d DefaultTransactionManager) startTransactionProcess(ctx context.Context, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(d.numberOfTransactionProcessorRoutines)
	for currentRoutine := 0; currentRoutine < d.numberOfTransactionProcessorRoutines; currentRoutine++ {
		go d.spawnTransactionProcessor(ctx, waitGroup)
	}
	<-ctx.Done()
	d.transactionFetchedQueue.shutdown()
	waitGroup.Wait()
}
func (d DefaultTransactionManager) spawnTransactionProcessor(ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("transaction processor shutdown" + " " + d.chainID)
			return
		default:
			_ = d.processNextTransaction(ctx)
			t := time.Second
			sleep(ctx, t)
		}
	}
}

func (d DefaultTransactionManager) processNextTransaction(ctx context.Context) error {
	currentTransaction := d.next()
	if currentTransaction != nil {
		failure := d.processTransaction(ctx, currentTransaction)
		if failure != nil {
			return failure
		}
	} else {
		log.Println("no more transactions to process")
	}
	return nil
}

func (d DefaultTransactionManager) next() *Transaction {
	return d.transactionFetchedQueue.poll()
}
func (d DefaultTransactionManager) processTransaction(ctx context.Context, transaction *Transaction) error {

	storedTransaction, err := d.transactionStorage.Get(ctx, transaction.ID)
	if err != nil {
		return err
	}
	if storedTransaction.State != TransactionFetched {
		return errors.New("transaction state is not fetched")
	}

	blockchainNetworlInfraAdapter, err := d.blockchainNetworkInfra.Use(ctx, d.blockchainNetworkInfraID)
	if err != nil {
		return err
	}
	//log.Println("llamamos get transaction")
	getTransactionInput := GetTransactionInput{}
	getTransactionInput.BlockchainNetworkType = d.chainID
	getTransactionInput.ID = transaction.ID
	getTransactionInput.BlockNumber = transaction.BlockNumber
	getTransactionOuput, err := blockchainNetworlInfraAdapter.GetTransaction(ctx, getTransactionInput)
	if err != nil {
		log.Println(transaction.ID + ":" + err.Error())
		return err
	}

	userCollection, err := d.userStorage.All(ctx, UserListOptions{})
	if err != nil {
		return err
	}
	addressesToFilter := make(map[string]struct{})
	for _, userId := range userCollection.Items {
		addresses, err := d.userAddressesStorage.All(ctx, UserAddresessListOptions{UserID: userId})
		if err != nil {
			return err
		}
		for _, address := range addresses.Items {
			addressesToFilter[address] = struct{}{} // Using an empty struct for minimal memory usage
		}
	}

	// TODO, add valid addresses and user to mocks, just now all the transactions are sent to kafka
	eventToEmit := KafkaEvent{}
	eventToEmit.Fees = getTransactionOuput.Fees
	eventToEmit.Amount = getTransactionOuput.Amount
	eventToEmit.Source = getTransactionOuput.Source
	eventToEmit.Destination = getTransactionOuput.Destination
	err = d.kafkaProducer.SendEvent(ctx, eventToEmit)
	if err != nil {
		return err
	}
	storedTransaction.State = TransactionProcessed
	storedTransaction.LastUpdated = time.Now()
	_, err = d.transactionStorage.Edit(ctx, *storedTransaction)
	if err != nil {
		return err
	}
	return nil

}

func (d DefaultTransactionManager) startTransactionOrphanCollectorProcess(ctx context.Context, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for {
			t := 100 * time.Second
			sleep(ctx, t)
			err := d.collectAndProcessOrphanTxs(ctx)
			if err != nil {
				return
			}

		}
	}()
	<-ctx.Done()
	log.Println("transaction orphan collector shutdown")
	waitGroup.Wait()
}
func (d DefaultTransactionManager) collectAndProcessOrphanTxs(ctx context.Context) error {
	moreItems := true
	for moreItems {
		//add pagination
		lastUpdateBefore := time.Now().Add(-5 * time.Second)
		listOptions := TransactionListOptions{}
		listOptions.State = TransactionFetched
		listOptions.LastUpdate = lastUpdateBefore
		transactionCollection, err := d.transactionStorage.All(ctx, listOptions)
		if err != nil {
			return err
		}
		for _, transaction := range transactionCollection.Items {
			err := d.processTransaction(ctx, &transaction)
			if err != nil {
				return err
			}
			sleep(ctx, 2*time.Second)
		}
	}
	return nil
}
