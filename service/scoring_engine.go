package service

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	_const "wolf_street/const"
	evaluate2 "wolf_street/evaluate"
	"wolf_street/pkginit"
)

type ScoringEngine struct {
	KC           KC
	TDSequential []int
	RSI          []float64
	StochRSI     []float64
	CCI          []float64
	KDJ          []KDJValue
	SAR          []float64
	Bollinger    BollingerBand
	MACD         MACD
	EMAShort     []float64
	EMALong      []float64
	ATR          []float64
	VWAP         []float64
	ArBr         ARBR
	CR           []float64
	Ichimoku     []float64
	Prices       []float64
	Candles      []Candle
}

func (se *ScoringEngine) Score(index int) (score int, signals []string) {
	score = 0
	signals = []string{}
	price := se.Prices[index]

	/* RSI */
	res, err := evaluate2.EvaluateRSISignals(se.RSI, index)
	if err != nil {
		pkginit.Logger.Error("EvaluateRSISignals failed", zap.Error(err))
		return
	}
	signals = append(signals, res.Signals...)
	score += score

	/* StochRSI */
	cfg := evaluate2.DefaultStochRSIConfig()
	// 可按需微调
	cfg.SlopeLookback = 4
	cfg.MinRiseBars = 3
	cfg.CrossoverHysteresis = 2
	cfg.EnableMTF = true
	if index > cfg.SlopeLookback {
		res, err = evaluate2.EvaluateStochRSISignals(se.StochRSI, index, cfg)
		if err != nil {
			pkginit.Logger.Error("EvaluateStochRSISignals failed", zap.Error(err))
			//return
		}
		/*	fmt.Println("Score:", res.Score)
			fmt.Println("Signals:", res.Signals)
			fmt.Println("Components:", res.Components) // 看到每一项的贡献，方便调参
		*/
		signals = append(signals, res.Signals...)
		score += int(res.Score)
	}

	// CCI
	if se.CCI[index] > 100 {
		score += 1
		signals = append(signals, "CCI强多头")
	} else if se.CCI[index] < -100 {
		score -= 1
		signals = append(signals, "CCI强空头")
	}

	/* KDJ */
	//if se.KDJ[index].J > 80 {
	//	score -= 1
	//	signals = append(signals, "KDJ超买")
	//} else if se.KDJ[index].J < 20 {
	//	score += 1
	//	signals = append(signals, "KDJ超卖")
	//}
	kdj := kdjAdapter{ref: se.KDJ}
	pr := priceAdapter{ref: se.Prices}
	evaluateScore, evaluateSignals, err := evaluate2.EvaluateKDJSignals(kdj, pr, index)
	if err != nil {
		log.Println("KDJ evaluation failed:", err)
	} else {
		score += evaluateScore
		signals = append(signals, evaluateSignals...)
		fmt.Println("KDJ score:", score)
		fmt.Println("Signals:", signals)
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
	if price > se.KC.UpperBand[index] {
		score += 1
		signals = append(signals, "价格突破Keltner上轨（趋势强势）")
	} else if price < se.KC.LowerBand[index] {
		score -= 1
		signals = append(signals, "价格跌破Keltner下轨（弱势）")
	}

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

func (se *ScoringEngine) ScoreBak(index int) (score int, signals []string) {
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
	if price > se.KC.UpperBand[index] {
		score += 1
		signals = append(signals, "价格突破Keltner上轨（趋势强势）")
	} else if price < se.KC.LowerBand[index] {
		score -= 1
		signals = append(signals, "价格跌破Keltner下轨（弱势）")
	}

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
	if score >= _const.TradeSignalBuyThreshold {
		return "BUY"
	} else if score <= _const.TradeSignalSellThreshold {
		return "SELL"
	} else {
		return "HOLD"
	}
}

type kdjAdapter struct{ ref []KDJValue }

func (a kdjAdapter) Len() int        { return len(a.ref) }
func (a kdjAdapter) K(i int) float64 { return a.ref[i].K }
func (a kdjAdapter) D(i int) float64 { return a.ref[i].D }
func (a kdjAdapter) J(i int) float64 { return a.ref[i].J }

type priceAdapter struct{ ref []float64 }

func (a priceAdapter) Len() int         { return len(a.ref) }
func (a priceAdapter) At(i int) float64 { return a.ref[i] }

// Example call from your engine
func (se *ScoringEngine) EvalKDJAt(index int) (int, []string, error) {
	kdj := kdjAdapter{ref: se.KDJ}
	pr := priceAdapter{ref: se.Prices}
	return evaluate2.EvaluateKDJSignals(kdj, pr, index)
}
