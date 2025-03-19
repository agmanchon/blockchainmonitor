# BlockchainMonitor Application

## Overview
The **BlockchainMonitor** application allows you to monitor and filter transactions from different configured blockchain networks and send them to a Kafka queue.To run the application follow the steps below

## Setup and Running the Application

### 1. Provision the Environment
- The application is developed using Go version 1.23.7. Make sure you have the correct version installed on your machine. You can verify your Go version by running:
   ```bash
  go version
- To start the required services (Kafka, Zookeeper, PostgreSQL), run the following command:
   ```bash
   make provision

This will execute docker-compose up using a pre-configured docker-compose.yml file that includes Kafka, Zookeeper, and PostgreSQL services.

### 2.Configure Your Blockdaemon API Key
 
- In the setup directory, you will find a config file. Add your Blockdaemon API key and your repository path to this file as shown below:
   ```bash
  database:
  postgres:
    host: "localhost"
    port: 5432
    scheme: "public"
    username: "postgres"
    password: "postgres"
    sslMode: "disable"
    database: "blockchain_monitor"
    schemaFilePath: "{repository-path}/blockchainmonitor/setup/schema.sql"

  queue:
  kafka:
  brokers: "localhost:9092"
  port: 9092
  topic: "transaction-event"
  
  blockchainNetwork:
  blockdaemon:
  apiKey: "{api-key}"
  eth:
  host: "https://svc.blockdaemon.com/ethereum/mainnet/native"
  btc:
  host: "https://svc.blockdaemon.com/bitcoin/mainnet/native"
  sol:
  host: "https://svc.blockdaemon.com/solana/mainnet/native"


- Replace your_api_key_here with your actual Blockdaemon API key.

### 3.Run the Application:

- After configuring the API key, you can start the application by running the following command:
   ```bash
  make run
