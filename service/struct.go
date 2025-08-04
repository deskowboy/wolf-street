package service

type Candle struct {
	Date  string
	Close float64
}

type Trade struct {
	BuyDate   string
	BuyPrice  float64
	SellDate  string
	SellPrice float64
}

type Result struct {
	Trades        []Trade
	TotalTrades   int
	WinningTrades int
	TotalReturn   float64
}
