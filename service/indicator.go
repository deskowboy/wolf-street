package service

import "math"

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
