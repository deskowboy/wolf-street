package service

import "fmt"

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

func PrintTradeStats(trades []Trade) {
	totalPnL := 0.0
	winCount := 0
	lossCount := 0

	fmt.Printf("\n ===== PrintTradeStats ===== \n\n")

	for _, trade := range trades {
		fmt.Printf("%s | %s @ %.2f | PnL: %.2f\n", trade.Date, trade.Signal, trade.Price, trade.PnL)
		totalPnL += trade.PnL
		if trade.PnL > 0 {
			winCount++
		} else if trade.PnL < 0 {
			lossCount++
		}
	}

	totalTrades := winCount + lossCount
	fmt.Printf("\n总盈亏: %.2f\n", totalPnL)
	fmt.Printf("胜率: %.2f%%\n", float64(winCount)/float64(totalTrades)*100)
	fmt.Printf("总交易次数: %d\n", totalTrades)
}
