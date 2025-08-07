package service

import (
	"fmt"
	"time"
)

func StrategyScoringEngine(candles []Candle) error {
	bar := NewTaggedProgressBar(len(candles), len(candles))

	var open, highs, lows, closes []float64
	for _, candle := range candles {
		open = append(open, candle.Open)
		highs = append(highs, candle.High)
		lows = append(lows, candle.Low)
		closes = append(closes, candle.Close)

		bar.Add(1)
		time.Sleep(10 * time.Millisecond)
		bar.Finish()
	}
	prices := closes

	rsi := CalculateRSI(prices, 14)
	stochRsi := CalculateStochRSI(prices, 14)
	cci := CalculateCCI(highs, lows, closes, 20)
	kdj := CalculateKDJ(highs, lows, closes, 9)
	sar := CalculateSAR(highs, lows, 0.02, 0.2)
	bollinger := CalculateBollinger(prices, 20)
	emaShort := CalculateEMA(prices, 12)
	emaLong := CalculateEMA(prices, 26)
	macd := CalculateMACD(prices)
	atr := CalculateATR(highs, lows, closes, 14)
	vwap := CalculateVWAP(candles)
	arbr := CalculateARBR(candles)
	cr := CalculateCR(candles, 26)
	ichimoku := CalculateIchimokuBaseLine(highs, lows, 26)
	kcband := CalculateKeltnerChannel(highs, lows, closes, 20)
	tdSeq := CalculateTDSequential(closes)

	se := ScoringEngine{
		RSI:          rsi,
		StochRSI:     stochRsi,
		CCI:          cci,
		KDJ:          kdj,
		SAR:          sar,
		Bollinger:    bollinger,
		MACD:         macd,
		EMAShort:     emaShort,
		EMALong:      emaLong,
		ATR:          atr,
		VWAP:         vwap,
		Prices:       prices, // 必须加这个
		Candles:      candles,
		ArBr:         arbr,
		CR:           cr,
		Ichimoku:     ichimoku,
		KC:           kcband,
		TDSequential: tdSeq,
		// 省略其他指标初始化
	}

	fmt.Println(" \n\n Scoring Engine Result: \n ")
	for i := 0; i < len(prices); i++ {
		score, signals := se.Score(i)
		if i < 10 {
			tradeSignal := GenerateTradeSignal(score)
			fmt.Printf(" (%d.) %s = $ %f | Score: %d, tradeSignal: %s , Signals: %v\n", i+1, candles[i].Date, candles[i].Close, score, tradeSignal, signals)
		}
	}

	/* Trading */
	trades := BacktestTrades(&se, 2)
	PrintTradeStats(trades)

	return nil
}
