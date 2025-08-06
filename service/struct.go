package service

type Candle struct {
	Date  string
	Open  float64
	High  float64
	Low   float64
	Close float64
}

type Trading struct {
	BuyDate   string
	BuyPrice  float64
	SellDate  string
	SellPrice float64
}

type Result struct {
	Trades        []Trading
	TotalTrades   int
	WinningTrades int
	TotalReturn   float64
}

/* Trading */

type Trade struct {
	Date   string
	Signal string
	Price  float64
	PnL    float64
}

/* Indicator */

type BollingerBand struct {
	LowerBand []float64
	MidBand   []float64
	UpperBand []float64
}

type KDJValue struct {
	K float64
	D float64
	J float64
}

type ARBR struct {
	AR []float64
	BR []float64
}

type MACD struct {
	MACDLine   []float64
	SignalLine []float64
	Histogram  []float64
}
