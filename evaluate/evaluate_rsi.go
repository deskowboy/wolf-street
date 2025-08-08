package evaluate

// =======================================================
// 附加：RSI 评分（可选）
// =======================================================

func EvaluateRSISignals(rsiIndex []float64, idx int) (EvalResult, error) {

	rsi := rsiIndex[idx]

	res := EvalResult{
		Score:      0,
		Signals:    []string{},
		Components: map[string]float64{},
	}
	switch {
	case rsi < 20:
		acc(&res, "rsi_severe_os", 2)
		res.Signals = append(res.Signals, "RSI严重超卖")
	case rsi >= 20 && rsi < 30:
		acc(&res, "rsi_os", 1)
		res.Signals = append(res.Signals, "RSI轻度超卖")
	case rsi > 70 && rsi <= 80:
		acc(&res, "rsi_ob", -1)
		res.Signals = append(res.Signals, "RSI轻度超买")
	case rsi > 80:
		acc(&res, "rsi_severe_ob", -2)
		res.Signals = append(res.Signals, "RSI严重超买")
	}
	res.Score = sumComponents(res.Components)
	return res, nil
}

func EvaluateRSISignalsBak(rsi float64) (int, []string) {
	score := 0
	var signals []string

	switch {
	case rsi < 20:
		score += 2
		signals = append(signals, "RSI严重超卖")
	case rsi >= 20 && rsi < 30:
		score += 1
		signals = append(signals, "RSI轻度超卖")
	case rsi > 70 && rsi <= 80:
		score -= 1
		signals = append(signals, "RSI轻度超买")
	case rsi > 80:
		score -= 2
		signals = append(signals, "RSI严重超买")
	}

	return score, signals
}
