package graph

import (
	"github.com/agmanchon/blockchainmonitor/config"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/kafkaconnection"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/postgresdb"
	"sync"
)

type FeatureInstallOptions struct {
	Configuration      config.InstallConfig
	PostgresConnection postgresdb.PostgresConnection
	MainWaitGroup      *sync.WaitGroup
	KafkaConnection    kafkaconnection.KafkaConnection
}
