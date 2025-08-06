package service

import "math"

/*  */
func CalculateKDJ(highs, lows, closes []float64, period int) []KDJValue {
	n := len(closes)
	kdj := make([]KDJValue, n)

	var k, d float64 = 50, 50

	for i := period - 1; i < n; i++ {
		low := lows[i-period+1]
		high := highs[i-period+1]
		for j := i - period + 1; j <= i; j++ {
			if lows[j] < low {
				low = lows[j]
			}
			if highs[j] > high {
				high = highs[j]
			}
		}

		if high != low {
			rsv := (closes[i] - low) / (high - low) * 100
			k = k*2/3 + rsv/3
			d = d*2/3 + k/3
			jValue := 3*k - 2*d
			kdj[i] = KDJValue{K: k, D: d, J: jValue}
		}
	}
	return kdj
}

/*  */
func CalculateSAR(highs, lows []float64, accelerationFactor float64, maxAccelerationFactor float64) []float64 {
	n := len(highs)
	sar := make([]float64, n)

	isUptrend := true
	af := accelerationFactor
	highest := highs[0]
	lowest := lows[0]

	for i := 1; i < n; i++ {
		if isUptrend {
			sar[i] = sar[i-1] + af*(highest-sar[i-1])
			if lows[i] < sar[i] {
				isUptrend = false
				sar[i] = highest
				af = accelerationFactor
				lowest = lows[i]
			} else {
				if highs[i] > highest {
					highest = highs[i]
					af = math.Min(af+accelerationFactor, maxAccelerationFactor)
				}
			}
		} else {
			sar[i] = sar[i-1] + af*(lowest-sar[i-1])
			if highs[i] > sar[i] {
				isUptrend = true
				sar[i] = lowest
				af = accelerationFactor
				highest = highs[i]
			} else {
				if lows[i] < lowest {
					lowest = lows[i]
					af = math.Min(af+accelerationFactor, maxAccelerationFactor)
				}
			}
		}
	}
	return sar
}

/*
RSI < 30 → 超卖区，考虑买入
RSI > 70 → 超买区，考虑卖出
*/
func CalculateRSI(prices []float64, period int) []float64 {
	rsi := make([]float64, len(prices))
	var gainSum, lossSum float64

	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change >= 0 {
			gainSum += change
		} else {
			lossSum -= change
		}
	}

	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - (100 / (1 + rs))
	}

	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change >= 0 {
			avgGain = (avgGain*(float64(period-1)) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*(float64(period-1)) - change) / float64(period)
		}

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	// period前的rsi值设为0或者NaN
	for i := 0; i < period; i++ {
		rsi[i] = 0
	}
	return rsi
}

/*
价格下穿布林带下轨，考虑买入
价格上穿布林带上轨，考虑卖出
*/
func CalculateBollinger(prices []float64, period int) (lowerBand, upperBand []float64) {
	n := len(prices)
	lowerBand = make([]float64, n)
	upperBand = make([]float64, n)

	var sum, sumSquares float64

	for i := 0; i < period; i++ {
		sum += prices[i]
		sumSquares += prices[i] * prices[i]
	}

	for i := period - 1; i < n; i++ {
		if i >= period {
			sum -= prices[i-period]
			sumSquares -= prices[i-period] * prices[i-period]
			sum += prices[i]
			sumSquares += prices[i] * prices[i]
		}

		mean := sum / float64(period)
		variance := (sumSquares / float64(period)) - (mean * mean)
		stddev := math.Sqrt(variance)

		upperBand[i] = mean + 2*stddev
		lowerBand[i] = mean - 2*stddev
	}

	return
}

/*
短期EMA上穿长期EMA → 黄金交叉，买入信号
短期EMA下穿长期EMA → 死亡交叉，卖出信号
*/
func CalculateEMA(prices []float64, period int) []float64 {
	ema := make([]float64, len(prices))
	k := 2.0 / (float64(period) + 1.0)

	// 初始值用SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema[period-1] = sum / float64(period)

	for i := period; i < len(prices); i++ {
		ema[i] = prices[i]*k + ema[i-1]*(1-k)
	}

	return ema
}

/*
MACD线上穿Signal线 → 金叉，买入信号
MACD线下穿Signal线 → 死叉，卖出信号
*/
func CalculateMACD(prices []float64) (macdLine, signalLine, histogram []float64) {
	ema12 := CalculateEMA(prices, 12)
	ema26 := CalculateEMA(prices, 26)

	n := len(prices)
	macdLine = make([]float64, n)
	for i := 0; i < n; i++ {
		macdLine[i] = ema12[i] - ema26[i]
	}

	// Signal Line是MACD Line的9周期EMA
	signalLine = CalculateEMA(macdLine, 9)

	histogram = make([]float64, n)
	for i := 0; i < n; i++ {
		histogram[i] = macdLine[i] - signalLine[i]
	}

	return
}

/*
ATR上升，信号可靠性增强 (不直接作为买卖信号)
ATR下降，信号减弱 (辅助判断信号有效性)
*/
func CalculateATR(highs, lows, closes []float64, period int) []float64 {
	atr := make([]float64, len(closes))
	trs := make([]float64, len(closes))

	for i := 1; i < len(closes); i++ {
		highLow := highs[i] - lows[i]
		highClose := math.Abs(highs[i] - closes[i-1])
		lowClose := math.Abs(lows[i] - closes[i-1])

		trs[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}

	// 初始ATR用SMA
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += trs[i]
	}
	atr[period] = sum / float64(period)

	for i := period + 1; i < len(closes); i++ {
		atr[i] = (atr[i-1]*(float64(period-1)) + trs[i]) / float64(period)
	}

	return atr
}

/*
StochRSI < 0.2 → 超卖 → 可能买入
StochRSI > 0.8 → 超买 → 可能卖出
*/
func CalculateStochRSI(prices []float64, period int) []float64 {
	rsi := CalculateRSI(prices, period)
	stochRsi := make([]float64, len(rsi))

	for i := period; i < len(rsi); i++ {
		lowest := rsi[i-period]
		highest := rsi[i-period]
		for j := i - period + 1; j <= i; j++ {
			if rsi[j] < lowest {
				lowest = rsi[j]
			}
			if rsi[j] > highest {
				highest = rsi[j]
			}
		}
		if highest-lowest == 0 {
			stochRsi[i] = 0
		} else {
			stochRsi[i] = (rsi[i] - lowest) / (highest - lowest)
		}
	}
	return stochRsi
}

/*
CCI > +100 → 多头强势（买入）
CCI < -100 → 空头强势（卖出）
*/
func CalculateCCI(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	cci := make([]float64, n)

	for i := period - 1; i < n; i++ {
		typicalPrices := []float64{}
		for j := i - period + 1; j <= i; j++ {
			typicalPrice := (highs[j] + lows[j] + closes[j]) / 3
			typicalPrices = append(typicalPrices, typicalPrice)
		}

		sum := 0.0
		for _, tp := range typicalPrices {
			sum += tp
		}
		meanTP := sum / float64(period)

		meanDeviation := 0.0
		for _, tp := range typicalPrices {
			meanDeviation += math.Abs(tp - meanTP)
		}
		meanDeviation /= float64(period)

		currentTP := (highs[i] + lows[i] + closes[i]) / 3
		if meanDeviation != 0 {
			cci[i] = (currentTP - meanTP) / (0.015 * meanDeviation)
		} else {
			cci[i] = 0
		}
	}

	return cci
}
