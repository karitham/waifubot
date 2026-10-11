package collection

// Source identifies how a character entered a collection.
type Source string

const (
	SourceRoll       Source = "ROLL"        // free roll
	SourceClaim      Source = "CLAIM"       // channel drop claim
	SourceSeriesRoll Source = "SERIES_ROLL" // token-paid series roll
	SourceTrade      Source = "TRADE"       // received from another user
	SourceExchange   Source = "EXCHANGE"    // sale result; never stored
)
