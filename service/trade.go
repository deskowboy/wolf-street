package service

func BacktestTrades(se *ScoringEngine, threshold int) []Trade {
	var trades []Trade
	var position string
	var entryPrice float64

	for i := 0; i < len(se.Prices); i++ {
		score, _ := se.Score(i)
		price := se.Prices[i]

		if position == "" {
			if score >= threshold {
				position = "LONG"
				entryPrice = price
				trades = append(trades, Trade{Date: se.Candles[i].Date, Signal: "BUY", Price: price})
			} else if score <= -threshold {
				position = "SHORT"
				entryPrice = price
				trades = append(trades, Trade{Date: se.Candles[i].Date, Signal: "SELL", Price: price})
			}
		} else if position == "LONG" && score <= -threshold {
			pnl := price - entryPrice
			trades = append(trades, Trade{Date: se.Candles[i].Date, Signal: "SELL", Price: price, PnL: pnl})
			position = ""
		} else if position == "SHORT" && score >= threshold {
			pnl := entryPrice - price
			trades = append(trades, Trade{Date: se.Candles[i].Date, Signal: "BUY", Price: price, PnL: pnl})
			position = ""
		}
	}
	return trades
}
