package usecases

const (
	// BlockFetched state implies that a block was fetched from the blockchain
	BlockFetched string = "FETCHED"
	// BlockProcessed state implies that a block's events were extracted and transactions were processed
	BlockProcessed string = "PROCESSED"
)
