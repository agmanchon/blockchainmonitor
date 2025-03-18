package cmd

import (
	"context"
	"github.com/agmanchon/blockchainmonitor/config"
	transactiongraph "github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/graph"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/kafkaconnection"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/postgresdb"
	"github.com/spf13/cobra"
	"log
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the application " + appName,
	Run:   startRun,
}

func startRun(_ *cobra.Command, _ []string) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.ReadInstallConfig()
	checkError(err)

	dbConnection, connectDbErr := postgresdb.ConnectDB(cfg.PostgresConfig)
	checkError(connectDbErr)
	defer dbConnection.GetDB().Close()

	executeSchemaErr := dbConnection.ExecuteSchema()
	checkError(executeSchemaErr)

	kafkaProducerConnection, queueConnectionErr := kafkaconnection.NewKafkaConnectionFromConfig(cfg.KafkaConfig.Kafka.Brokers, cfg.KafkaConfig.Kafka.Topic)
	if queueConnectionErr != nil {
		checkError(queueConnectionErr)
	}

	var mainWaitGroup = &sync.WaitGroup{}

	var installInput installFeaturesInput
	installInput.configuration = *cfg
	installInput.WaitGroup = mainWaitGroup
	installInput.postgresConnection = dbConnection
	installInput.kafkaProductorConnection = *kafkaProducerConnection

	go startTransactionsFeature(ctx, installInput)

	for {
		select {
		case <-signalChan:
			cancel()
			mainWaitGroup.Wait()
			log.Println("Shutdown complete.")
			os.Exit(0)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type installFeaturesInput struct {
	configuration            config.InstallConfig
	postgresConnection       postgresdb.PostgresConnection
	kafkaProductorConnection kafkaconnection.KafkaConnection
	*sync.WaitGroup
}

func startTransactionsFeature(ctx context.Context, input installFeaturesInput) {
	defer input.WaitGroup.Done()
	options := transactiongraph.FeatureInstallOptions{}
	options.PostgresConnection = input.postgresConnection
	options.MainWaitGroup = input.WaitGroup
	options.Configuration = input.configuration
	options.KafkaConnection = input.kafkaProductorConnection
	input.WaitGroup.Add(1)
	go transactiongraph.Install(ctx, options)

	log.Println("transaction feature finished")

}
