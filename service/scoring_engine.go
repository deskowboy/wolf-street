package service

import "fmt"

// --- 打分引擎 (Scoring Engine) ---
type ScoringEngine struct {
	RSI        []float64
	UpperBand  []float64
	LowerBand  []float64
	MACDLine   []float64
	SignalLine []float64
	EMAShort   []float64
	EMALong    []float64
	ATR        []float64
	Prices     []float64
}

func (se *ScoringEngine) Score(index int) (score int, signals []string) {
	score = 0
	signals = []string{}

	// RSI 信号
	if se.RSI[index] < 30 {
		score += 1
		signals = append(signals, "RSI超卖")
	} else if se.RSI[index] > 70 {
		score -= 1
		signals = append(signals, "RSI超买")
	}

	// 布林带信号
	price := se.Prices[index]
	if price < se.LowerBand[index] {
		score += 1
		signals = append(signals, "价格穿破布林下轨")
	} else if price > se.UpperBand[index] {
		score -= 1
		signals = append(signals, "价格穿破布林上轨")
	}

	// MACD 信号
	if index > 0 {
		if se.MACDLine[index-1] < se.SignalLine[index-1] && se.MACDLine[index] > se.SignalLine[index] {
			score += 1
			signals = append(signals, "MACD金叉")
		} else if se.MACDLine[index-1] > se.SignalLine[index-1] && se.MACDLine[index] < se.SignalLine[index] {
			score -= 1
			signals = append(signals, "MACD死叉")
		}
	}

	// EMA排列信号
	if se.EMAShort[index] > se.EMALong[index] {
		score += 1
		signals = append(signals, "短期EMA上穿长期EMA")
	} else {
		score -= 1
		signals = append(signals, "短期EMA下穿长期EMA")
	}

	// ATR (波动辅助判断)
	if index > 0 && se.ATR[index] > se.ATR[index-1] {
		signals = append(signals, "ATR上升")
	} else if index > 0 && se.ATR[index] < se.ATR[index-1] {
		signals = append(signals, "ATR下降")
	}

	return
}

func StrategyScoringEngine(candles []Candle) error {
	var open, highs, lows, closes []float64
	for _, candle := range candles {
		open = append(open, candle.Open)
		highs = append(highs, candle.High)
		lows = append(lows, candle.Low)
		closes = append(closes, candle.Close)
	}
	prices := closes

	// 计算各类指标
	rsi := CalculateRSI(prices, 14)
	lowerBand, upperBand := CalculateBollinger(prices, 20)
	emaShort := CalculateEMA(prices, 12)
	emaLong := CalculateEMA(prices, 26)
	macdLine, signalLine, _ := CalculateMACD(prices)
	atr := CalculateATR(highs, lows, closes, 14)

	// 初始化 Scoring Engine
	se := ScoringEngine{
		RSI:        rsi,
		UpperBand:  upperBand,
		LowerBand:  lowerBand,
		MACDLine:   macdLine,
		SignalLine: signalLine,
		EMAShort:   emaShort,
		EMALong:    emaLong,
		ATR:        atr,
		Prices:     closes,
	}

	// 逐个时间点评分
	var scores []int
	var allSignals [][]string

	for i := 0; i < len(prices); i++ {
		score, signals := se.Score(i)
		scores = append(scores, score)
		allSignals = append(allSignals, signals)
		if len(signals) > 0 {
			fmt.Printf(" (%d.) %s : %f : Score = %d, Signals = %v\n", i, candles[i].Date, candles[i].Close, score, signals)
		}
	}

	return nil
}
