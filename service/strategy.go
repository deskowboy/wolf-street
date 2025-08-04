package service

func StrategyRSIBollinger(candles []Candle) Result {
	prices := make([]float64, len(candles))
	for i, c := range candles {
		prices[i] = c.Close
	}

	rsi := CalculateRSI(prices, 14)
	lowerBand, _ := CalculateBollinger(prices, 20)

	var trades []Trade
	holding := false
	buyPrice := 0.0
	buyDate := ""
	holdDays := 0

	for i := 20; i < len(candles); i++ {
		if !holding {
			if rsi[i] < 30 && prices[i] < lowerBand[i] {
				holding = true
				buyPrice = prices[i]
				buyDate = candles[i].Date
				holdDays = 0
			}
		} else {
			holdDays++
			if rsi[i] > 50 || holdDays >= 3 {
				trades = append(trades, Trade{
					BuyDate:   buyDate,
					BuyPrice:  buyPrice,
					SellDate:  candles[i].Date,
					SellPrice: prices[i],
				})
				holding = false
			}
		}
	}

	result := Result{Trades: trades, TotalTrades: len(trades)}
	for _, trade := range trades {
		if trade.SellPrice > trade.BuyPrice {
			result.WinningTrades++
		}
		result.TotalReturn += (trade.SellPrice - trade.BuyPrice) / trade.BuyPrice
	}

	return result
}
