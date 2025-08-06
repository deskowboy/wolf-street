package service

import "fmt"

type ScoringEngine struct {
	KCUpper      []float64
	KCMiddle     []float64
	KCLower      []float64
	TDSequential []int
	RSI          []float64
	StochRSI     []float64
	CCI          []float64
	KDJ          []KDJValue
	SAR          []float64
	Bollinger    BollingerBand
	//UpperBand    []float64
	//LowerBand    []float64
	//MACDLine   []float64
	//SignalLine []float64
	MACD     MACD
	EMAShort []float64
	EMALong  []float64
	ATR      []float64
	VWAP     []float64
	ArBr     ARBR
	CR       []float64
	Ichimoku []float64
	Prices   []float64
	Candles  []Candle
}

func (se *ScoringEngine) Score(index int) (score int, signals []string) {
	score = 0
	signals = []string{}
	price := se.Prices[index]

	// RSI
	if se.RSI[index] < 30 {
		score += 1
		signals = append(signals, "RSI超卖")
	} else if se.RSI[index] > 70 {
		score -= 1
		signals = append(signals, "RSI超买")
	}

	// StochRSI
	if se.StochRSI[index] < 0.2 {
		score += 1
		signals = append(signals, "StochRSI超卖")
	} else if se.StochRSI[index] > 0.8 {
		score -= 1
		signals = append(signals, "StochRSI超买")
	}

	// CCI
	if se.CCI[index] > 100 {
		score += 1
		signals = append(signals, "CCI强多头")
	} else if se.CCI[index] < -100 {
		score -= 1
		signals = append(signals, "CCI强空头")
	}

	// KDJ
	if se.KDJ[index].J > 80 {
		score -= 1
		signals = append(signals, "KDJ超买")
	} else if se.KDJ[index].J < 20 {
		score += 1
		signals = append(signals, "KDJ超卖")
	}

	// Bollinger Bands
	if price < se.Bollinger.LowerBand[index] {
		score += 1
		signals = append(signals, "布林带下轨突破")
	} else if price > se.Bollinger.UpperBand[index] {
		score -= 1
		signals = append(signals, "布林带上轨突破")
	}

	// EMA
	if se.EMAShort[index] > se.EMALong[index] {
		score += 1
		signals = append(signals, "EMA金叉")
	} else if se.EMAShort[index] < se.EMALong[index] {
		score -= 1
		signals = append(signals, "EMA死叉")
	}

	// MACD
	if index > 0 {
		if se.MACD.MACDLine[index-1] < se.MACD.SignalLine[index-1] && se.MACD.MACDLine[index] > se.MACD.SignalLine[index] {
			score += 1
			signals = append(signals, "MACD金叉")
		} else if se.MACD.MACDLine[index-1] > se.MACD.SignalLine[index-1] && se.MACD.MACDLine[index] < se.MACD.SignalLine[index] {
			score -= 1
			signals = append(signals, "MACD死叉")
		}
	}

	// SAR 反转信号
	if price > se.SAR[index] {
		score += 1
		signals = append(signals, "SAR支撑")
	} else if price < se.SAR[index] {
		score -= 1
		signals = append(signals, "SAR压制")
	}

	// ATR 辅助
	if index > 0 && se.ATR[index] > se.ATR[index-1] {
		signals = append(signals, "ATR上升")
	} else if index > 0 && se.ATR[index] < se.ATR[index-1] {
		signals = append(signals, "ATR下降")
	}

	// VWAP 信号
	if price > se.VWAP[index] {
		score += 1
		signals = append(signals, "价格上穿VWAP（强势）")
	} else if price < se.VWAP[index] {
		score -= 1
		signals = append(signals, "价格下穿VWAP（弱势）")
	}

	// ARBR 信号
	if se.ArBr.AR[index] > 120 && se.ArBr.BR[index] > 120 {
		score += 1
		signals = append(signals, "ARBR极强多头")
	} else if se.ArBr.AR[index] < 80 && se.ArBr.BR[index] < 80 {
		score -= 1
		signals = append(signals, "ARBR极弱空头")
	}

	// CR 信号
	if se.CR[index] > 150 {
		score += 1
		signals = append(signals, "CR强多头确认")
	} else if se.CR[index] < 100 {
		score -= 1
		signals = append(signals, "CR偏空头确认")
	}

	// Ichimoku 基准线信号
	if price > se.Ichimoku[index] {
		score += 1
		signals = append(signals, "价格上穿一目均衡表基准线（偏多）")
	} else if price < se.Ichimoku[index] {
		score -= 1
		signals = append(signals, "价格下穿一目均衡表基准线（偏空）")
	}

	// Keltner Channel (KC)
	//if price > se.KCUpper[index] {
	//	score += 1
	//	signals = append(signals, "价格突破Keltner上轨（趋势强势）")
	//} else if price < se.KCLower[index] {
	//	score -= 1
	//	signals = append(signals, "价格跌破Keltner下轨（弱势）")
	//}

	// TD Sequential
	if se.TDSequential[index] == 9 {
		score -= 1
		signals = append(signals, "TD9顶部反转警告")
	} else if se.TDSequential[index] == -9 {
		score += 1
		signals = append(signals, "TD9底部反转警告")
	}

	return
}

func GenerateTradeSignal(score int) string {
	if score >= 2 {
		return "BUY"
	} else if score <= -2 {
		return "SELL"
	} else {
		return "HOLD"
	}
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
	//kcUpper, kcMiddle, kcLower := CalculateKeltnerChannel(highs, lows, closes, 20)
	tdSeq := CalculateTDSequential(closes)

	se := ScoringEngine{
		RSI:       rsi,
		StochRSI:  stochRsi,
		CCI:       cci,
		KDJ:       kdj,
		SAR:       sar,
		Bollinger: bollinger,
		MACD:      macd,
		EMAShort:  emaShort,
		EMALong:   emaLong,
		ATR:       atr,
		VWAP:      vwap,
		Prices:    prices, // 必须加这个
		Candles:   candles,
		ArBr:      arbr,
		CR:        cr,
		Ichimoku:  ichimoku,
		//KCUpper:      kcUpper,
		//KCMiddle:     kcMiddle,
		//KCLower:      kcLower,
		TDSequential: tdSeq,
		// 省略其他指标初始化
	}

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
