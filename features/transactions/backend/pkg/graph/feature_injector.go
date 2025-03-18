package graph

import (
	"context"
	"log"

	"github.com/agmanchon/blockchainmonitor/config"
	blockdaemonbtcadapter "github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/blockdaemon/btc"
	blockdaemonethadapter "github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/blockdaemon/eth"
	blockdaemonsoladapter "github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/blockdaemon/sol"
	kafkaadapter "github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/kafka"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/postgres"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/kafkaconnection"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/postgresdb"
	"sync"
)

type FeatureInstall struct {
	configuration      config.InstallConfig
	postgresConnection postgresdb.PostgresConnection
	mainWaitGroup      *sync.WaitGroup
	kafkaConnection    kafkaconnection.KafkaConnection
}

func Install(ctx context.Context, options FeatureInstallOptions) *FeatureInstall {
	feature := &FeatureInstall{
		configuration:      options.Configuration,
		postgresConnection: options.PostgresConnection,
		mainWaitGroup:      options.MainWaitGroup,
		kafkaConnection:    options.KafkaConnection,
	}
	return feature.install(ctx)
}

func (feature *FeatureInstall) install(ctx context.Context) *FeatureInstall {

	err := feature.initializeFeature(ctx)
	if err != nil {
		checkError(err)
	}
	return nil
}
func (feature *FeatureInstall) initializeFeature(ctx context.Context) error {
	//defer feature.mainWaitGroup.Done()
	// Infra adapters
	blockchainNetworkInfraAdapterManager, err := usecases.ProvideDefaultBlockchainNetworkInfraAdapterManager()
	if err != nil {
		return err
	}
	blockDaemonNetworkFactoryOptions := blockchainnetworkinfra.BlockDaemonNetworkFactoryOptions{}
	blockDaemonNetworkFactoryOptions.APIKey = feature.configuration.BlockchainNetwork.Blockdaemon.APIKey
	blockDaemonNetworkFactory, err := blockchainnetworkinfra.NewBlockDaemonNetworkFactory(blockDaemonNetworkFactoryOptions)
	if err != nil {
		return err
	}

	err = blockchainNetworkInfraAdapterManager.RegisterBlockchainNetworkFactory(blockDaemonNetworkFactory)
	if err != nil {
		return err
	}
	defaultBtcAdapterOptions := blockdaemonbtcadapter.DefaultBtcAdapterOptions{}
	defaultBtcAdapterOptions.Host = feature.configuration.BlockchainNetwork.Blockdaemon.Btc.Host
	defaultBtcAdapter, err := blockdaemonbtcadapter.ProvideDefaultBtcAdapter(defaultBtcAdapterOptions)
	if err != nil {
		return err
	}
	blockDaemonNetworkFactory.RegisterBlockDaemonNetwork(blockchainnetworkinfra.BlockdaemonBtcNetwork, defaultBtcAdapter)

	defaultEthAdapterOptions := blockdaemonethadapter.DefaultEthAdapterOptions{}
	defaultEthAdapterOptions.Host = feature.configuration.BlockchainNetwork.Blockdaemon.Eth.Host
	defaultEthAdapter, err := blockdaemonethadapter.ProvideDefaultEthAdapter(defaultEthAdapterOptions)
	if err != nil {
		return err
	}
	blockDaemonNetworkFactory.RegisterBlockDaemonNetwork(blockchainnetworkinfra.BlockdaemonEthNetwork, defaultEthAdapter)

	defaultSolAdapterOptions := blockdaemonsoladapter.DefaultSolAdapterOptions{}
	defaultSolAdapterOptions.Host = feature.configuration.BlockchainNetwork.Blockdaemon.Sol.Host
	defaultSolAdapter, err := blockdaemonsoladapter.ProvideDefaultSolAdapter(defaultSolAdapterOptions)
	if err != nil {
		return err
	}
	blockDaemonNetworkFactory.RegisterBlockDaemonNetwork(blockchainnetworkinfra.BlockdaemonSolNetwork, defaultSolAdapter)

	//Postgres adapters
	postgresBlockPointerStorageAdapterOptions := postgres.PostgresBlockPointerStorageAdapterOptions{}
	postgresBlockPointerStorageAdapterOptions.DB = feature.postgresConnection.GetDB()
	postgresBlockPointerStorageAdapter, err := postgres.ProvidePostgresBlockPointerStorageAdapter(postgresBlockPointerStorageAdapterOptions)
	if err != nil {
		return err
	}

	postgresBlockStorageAdapterOptions := postgres.PostgresBlockStorageAdapterOptions{}
	postgresBlockStorageAdapterOptions.DB = feature.postgresConnection.GetDB()
	postgresBlockStorageAdapter, err := postgres.ProvidePostgresBlockStorageAdapter(postgresBlockStorageAdapterOptions)
	if err != nil {
		return err
	}

	postgresTransactionStorageAdapterOptions := postgres.PostgresTransactionStorageAdapterOptions{}
	postgresTransactionStorageAdapterOptions.DB = feature.postgresConnection.GetDB()
	postgresTransactionStorageAdapter, err := postgres.ProvidePostgresTransactionStorageAdapter(postgresTransactionStorageAdapterOptions)
	if err != nil {
		return err
	}
	postgresUserAddressesMockStorageAdapterOptions := postgres.PostgresUserAddressesMockStorageAdapterOptions{}
	postgresUserAddressesMockStorageAdapter, err := postgres.ProvidePostgresUserAddressesStorageAdapter(postgresUserAddressesMockStorageAdapterOptions)
	if err != nil {
		return err
	}

	postgresUserMockStorageAdapterOptions := postgres.PostgresUserMockStorageAdapterOptions{}
	postgresUserMockStorageAdapter, err := postgres.ProvidePostgresUserMockStorageAdapter(postgresUserMockStorageAdapterOptions)
	if err != nil {
		return err
	}

	// kafka
	kafkaProducerAdapterOptions := kafkaadapter.KafkaProducerAdapterOptions{}
	kafkaProducerAdapterOptions.Topic = feature.configuration.KafkaConfig.Kafka.Topic
	kafkaProducerAdapterOptions.Producer = feature.kafkaConnection.Producer
	kafkaProducerAdapter, err := kafkaadapter.NewKafkaProducerAdapter(kafkaProducerAdapterOptions)
	if err != nil {
		return err
	}
	defer kafkaProducerAdapter.Close()

	kafkaConsumerAdapterOptions := kafkaadapter.KafkaConsumerAdapterOptions{}
	kafkaConsumerAdapterOptions.Topic = feature.configuration.KafkaConfig.Kafka.Topic
	kafkaConsumerAdapterOptions.Consumer = feature.kafkaConnection.Consumer
	kafkaConsumerAdapter, err := kafkaadapter.NewKafkaConsumerAdapter(kafkaConsumerAdapterOptions)
	if err != nil {
		return err
	}
	defer kafkaConsumerAdapter.Close()

	// Sync Queue

	ethTransactionsBlockingSyncQueueUsecases := usecases.NewTransactionsBlockingSyncQueue(blockchainnetworkinfra.BlockdaemonEthNetwork + usecases.TransactionFetched)
	btcTransactionsBlockingSyncQueueUsecases := usecases.NewTransactionsBlockingSyncQueue(blockchainnetworkinfra.BlockdaemonBtcNetwork + usecases.TransactionFetched)
	solTransactionsBlockingSyncQueueUsecases := usecases.NewTransactionsBlockingSyncQueue(blockchainnetworkinfra.BlockdaemonSolNetwork + usecases.TransactionFetched)

	//Usecases

	workerManager := usecases.NewDefaultWorkerManager()
	blockmanagerEthUsecasesOptions := usecases.DefaultBlockManagerOptions{}
	blockmanagerEthUsecasesOptions.BlockStorage = *postgresBlockStorageAdapter
	blockmanagerEthUsecasesOptions.TransactionStorage = *postgresTransactionStorageAdapter
	blockmanagerEthUsecasesOptions.BlockPointerStorage = *postgresBlockPointerStorageAdapter
	blockmanagerEthUsecasesOptions.BlockchainNetworkInfra = blockchainNetworkInfraAdapterManager
	blockmanagerEthUsecasesOptions.ChainID = blockchainnetworkinfra.BlockdaemonEthNetwork
	blockmanagerEthUsecasesOptions.TransactionFetchedQueue = ethTransactionsBlockingSyncQueueUsecases
	blockmanagerEthUsecasesOptions.BlockchainNetworkInfraID = blockchainnetworkinfra.BlockDaemonInfraType
	blockManagerEthUsecases, err := usecases.ProvideDefaultBlockManager(blockmanagerEthUsecasesOptions)
	if err != nil {
		return err
	}

	transactionsEthUsecasesOptions := usecases.DefaultTransactionManagerOptions{}
	transactionsEthUsecasesOptions.TransactionStorage = *postgresTransactionStorageAdapter
	transactionsEthUsecasesOptions.TransactionFetchedQueue = ethTransactionsBlockingSyncQueueUsecases
	transactionsEthUsecasesOptions.BlockchainNetworkInfra = blockchainNetworkInfraAdapterManager
	transactionsEthUsecasesOptions.ChainID = blockchainnetworkinfra.BlockdaemonEthNetwork
	transactionsEthUsecasesOptions.BlockchainNetworkInfraID = blockchainnetworkinfra.BlockDaemonInfraType
	transactionsEthUsecasesOptions.KafkaProducer = kafkaProducerAdapter
	transactionsEthUsecasesOptions.UserAddressesStorage = postgresUserAddressesMockStorageAdapter
	transactionsEthUsecasesOptions.UserStorage = postgresUserMockStorageAdapter
	transactionEthManagerUsecases, err := usecases.ProvideDefaultTransactionManagerUseCase(transactionsEthUsecasesOptions)
	if err != nil {
		return err
	}

	blockmanagerSolUsecasesOptions := usecases.DefaultBlockManagerOptions{}
	blockmanagerSolUsecasesOptions.BlockStorage = *postgresBlockStorageAdapter
	blockmanagerSolUsecasesOptions.TransactionStorage = *postgresTransactionStorageAdapter
	blockmanagerSolUsecasesOptions.BlockPointerStorage = *postgresBlockPointerStorageAdapter
	blockmanagerSolUsecasesOptions.BlockchainNetworkInfra = blockchainNetworkInfraAdapterManager
	blockmanagerSolUsecasesOptions.ChainID = blockchainnetworkinfra.BlockdaemonSolNetwork
	blockmanagerSolUsecasesOptions.TransactionFetchedQueue = solTransactionsBlockingSyncQueueUsecases
	blockmanagerSolUsecasesOptions.BlockchainNetworkInfraID = blockchainnetworkinfra.BlockDaemonInfraType
	blockManagerSolUsecases, err := usecases.ProvideDefaultBlockManager(blockmanagerSolUsecasesOptions)
	if err != nil {
		return err
	}

	transactionsSolUsecasesOptions := usecases.DefaultTransactionManagerOptions{}
	transactionsSolUsecasesOptions.TransactionStorage = *postgresTransactionStorageAdapter
	transactionsSolUsecasesOptions.TransactionFetchedQueue = solTransactionsBlockingSyncQueueUsecases
	transactionsSolUsecasesOptions.BlockchainNetworkInfra = blockchainNetworkInfraAdapterManager
	transactionsSolUsecasesOptions.ChainID = blockchainnetworkinfra.BlockdaemonSolNetwork
	transactionsSolUsecasesOptions.BlockchainNetworkInfraID = blockchainnetworkinfra.BlockDaemonInfraType
	transactionsSolUsecasesOptions.KafkaProducer = kafkaProducerAdapter
	transactionsSolUsecasesOptions.UserAddressesStorage = postgresUserAddressesMockStorageAdapter
	transactionsSolUsecasesOptions.UserStorage = postgresUserMockStorageAdapter
	transactionSolManagerUsecases, err := usecases.ProvideDefaultTransactionManagerUseCase(transactionsSolUsecasesOptions)
	if err != nil {
		return err
	}

	blockmanagerBtcUsecasesOptions := usecases.DefaultBlockManagerOptions{}
	blockmanagerBtcUsecasesOptions.BlockStorage = *postgresBlockStorageAdapter
	blockmanagerBtcUsecasesOptions.TransactionStorage = *postgresTransactionStorageAdapter
	blockmanagerBtcUsecasesOptions.BlockPointerStorage = *postgresBlockPointerStorageAdapter
	blockmanagerBtcUsecasesOptions.BlockchainNetworkInfra = blockchainNetworkInfraAdapterManager
	blockmanagerBtcUsecasesOptions.ChainID = blockchainnetworkinfra.BlockdaemonBtcNetwork
	blockmanagerBtcUsecasesOptions.TransactionFetchedQueue = btcTransactionsBlockingSyncQueueUsecases
	blockmanagerBtcUsecasesOptions.BlockchainNetworkInfraID = blockchainnetworkinfra.BlockDaemonInfraType
	blockManagerBtcUsecases, err := usecases.ProvideDefaultBlockManager(blockmanagerBtcUsecasesOptions)
	if err != nil {
		return err
	}

	transactionsBtcUsecasesOptions := usecases.DefaultTransactionManagerOptions{}
	transactionsBtcUsecasesOptions.TransactionStorage = *postgresTransactionStorageAdapter
	transactionsBtcUsecasesOptions.TransactionFetchedQueue = btcTransactionsBlockingSyncQueueUsecases
	transactionsBtcUsecasesOptions.BlockchainNetworkInfra = blockchainNetworkInfraAdapterManager
	transactionsBtcUsecasesOptions.ChainID = blockchainnetworkinfra.BlockdaemonBtcNetwork
	transactionsBtcUsecasesOptions.BlockchainNetworkInfraID = blockchainnetworkinfra.BlockDaemonInfraType
	transactionsBtcUsecasesOptions.KafkaProducer = kafkaProducerAdapter
	transactionsBtcUsecasesOptions.UserAddressesStorage = postgresUserAddressesMockStorageAdapter
	transactionsBtcUsecasesOptions.UserStorage = postgresUserMockStorageAdapter
	transactionBtcManagerUsecases, err := usecases.ProvideDefaultTransactionManagerUseCase(transactionsBtcUsecasesOptions)
	if err != nil {
		return err
	}
	waitGroup := &sync.WaitGroup{}
	go kafkaConsumerAdapter.StartConsumer(ctx, handleMessage)
	workerManager.Add(blockManagerEthUsecases)
	workerManager.Add(transactionEthManagerUsecases)
	workerManager.Add(blockManagerSolUsecases)
	workerManager.Add(transactionSolManagerUsecases)
	workerManager.Add(blockManagerBtcUsecases)
	workerManager.Add(transactionBtcManagerUsecases)
	waitGroup.Add(1)
	workerManager.Start(ctx, waitGroup)

	<-ctx.Done()
	waitGroup.Wait()
	log.Println("transaction thread finished")

	return nil
}

func handleMessage(event usecases.KafkaEvent) error {
	// Process the event (you can add your custom logic here)
	//log.Printf("Received event: %v\n", event)
	return nil
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
