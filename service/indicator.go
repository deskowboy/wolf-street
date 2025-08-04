package service

import "math"

func CalculateRSI(prices []float64, period int) []float64 {
	rsi := make([]float64, len(prices))
	var gainSum, lossSum float64

	for i := 1; i < period; i++ {
		change := prices[i] - prices[i-1]
		if change >= 0 {
			gainSum += change
		} else {
			lossSum -= change
		}
	}

	for i := period; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change >= 0 {
			gainSum = (gainSum*(float64(period-1)) + change) / float64(period)
			lossSum = (lossSum * float64(period-1)) / float64(period)
		} else {
			gainSum = (gainSum * float64(period-1)) / float64(period)
			lossSum = (lossSum*(float64(period-1)) - change) / float64(period)
		}

		rs := gainSum / (lossSum + 1e-6)
		rsi[i] = 100 - (100 / (1 + rs))
	}
	return rsi
}

func CalculateBollinger(prices []float64, period int) (lowerBand, upperBand []float64) {
	lowerBand = make([]float64, len(prices))
	upperBand = make([]float64, len(prices))

	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += prices[j]
		}
		mean := sum / float64(period)

		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			variance += (prices[j] - mean) * (prices[j] - mean)
		}
		stddev := math.Sqrt(variance / float64(period))

		upperBand[i] = mean + 2*stddev
		lowerBand[i] = mean - 2*stddev
	}
	return
}
