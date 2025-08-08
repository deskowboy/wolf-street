// package evaluate
package evaluate

import (
	"errors"
	"fmt"
	"math"
)

type KDJConfig struct {
	JOverbought, JExtremeOverbought float64
	JSold, JExtremeSold             float64
	JMomentumStep                   float64
	KDWideGap                       float64
	PersistenceN                    int

	ScoreOverbought, ScoreExtremeOverbought int
	ScoreOversold, ScoreExtremeOversold     int
	ScoreGoldenLow, ScoreGolden             int
	ScoreDeathHigh, ScoreDeath              int
	ScoreJUp, ScoreJDown                    int
	ScoreKDom, ScoreDDom                    int
	ScoreHighPersist, ScoreLowPersist       int
	ScoreBearDiv, ScoreBullDiv              int
}

func DefaultKDJConfig() KDJConfig {
	return KDJConfig{
		JOverbought: 80, JExtremeOverbought: 90,
		JSold: 20, JExtremeSold: 10,
		JMomentumStep: 10,
		KDWideGap:     20,
		PersistenceN:  3,

		ScoreOverbought: -1, ScoreExtremeOverbought: -2,
		ScoreOversold: +1, ScoreExtremeOversold: +2,
		ScoreGoldenLow: +2, ScoreGolden: +1,
		ScoreDeathHigh: -2, ScoreDeath: -1,
		ScoreJUp: +1, ScoreJDown: -1,
		ScoreKDom: +1, ScoreDDom: -1,
		ScoreHighPersist: -1, ScoreLowPersist: +1,
		ScoreBearDiv: -2, ScoreBullDiv: +2,
	}
}

var (
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrNotEnoughData   = errors.New("not enough data for KDJ evaluation")
)

// ---- Interfaces to avoid importing service ----
type KDJSeries interface {
	Len() int
	K(i int) float64
	D(i int) float64
	J(i int) float64
}
type PriceSeries interface {
	Len() int
	At(i int) float64
}

// Public API (no service.ScoringEngine here)
func EvaluateKDJSignals(kdj KDJSeries, prices PriceSeries, index int) (int, []string, error) {
	return EvaluateKDJSignalsWithConfig(kdj, prices, index, DefaultKDJConfig())
}

func EvaluateKDJSignalsWithConfig(kdj KDJSeries, prices PriceSeries, index int, cfg KDJConfig) (int, []string, error) {
	if kdj == nil || kdj.Len() == 0 {
		return 0, nil, ErrNotEnoughData
	}
	if index < 0 || index >= kdj.Len() {
		return 0, nil, ErrIndexOutOfRange
	}
	if index == 0 {
		return 0, nil, nil
	}
	score, signals := kdjScore(kdj, prices, index, cfg)
	return score, signals, nil
}

func kdjScore(s KDJSeries, prices PriceSeries, i int, cfg KDJConfig) (int, []string) {
	k, d, j := s.K(i), s.D(i), s.J(i)
	kp, dp, jp := s.K(i-1), s.D(i-1), s.J(i-1)

	score := 0
	var signals []string

	// 1) J bands
	switch {
	case j >= cfg.JExtremeOverbought:
		score += cfg.ScoreExtremeOverbought
		signals = append(signals, fmt.Sprintf("KDJ 极度超买(J≥%.0f)", cfg.JExtremeOverbought))
	case j >= cfg.JOverbought:
		score += cfg.ScoreOverbought
		signals = append(signals, fmt.Sprintf("KDJ 超买(%.0f≤J<%.0f)", cfg.JOverbought, cfg.JExtremeOverbought))
	case j <= cfg.JExtremeSold:
		score += cfg.ScoreExtremeOversold
		signals = append(signals, fmt.Sprintf("KDJ 极度超卖(J≤%.0f)", cfg.JExtremeSold))
	case j <= cfg.JSold:
		score += cfg.ScoreOversold
		signals = append(signals, fmt.Sprintf("KDJ 超卖(%.0f<J≤%.0f)", cfg.JExtremeSold, cfg.JSold))
	}

	// 2) Crosses
	golden := (kp <= dp) && (k > d)
	death := (kp >= dp) && (k < d)
	if golden {
		w := cfg.ScoreGolden
		if k < 50 && d < 50 {
			w = cfg.ScoreGoldenLow
		}
		score += w
		signals = append(signals, fmt.Sprintf("KDJ 金叉(权重 %+d)", w))
	}
	if death {
		w := cfg.ScoreDeath
		if k > 50 && d > 50 {
			w = cfg.ScoreDeathHigh
		}
		score += w
		signals = append(signals, fmt.Sprintf("KDJ 死叉(权重 %+d)", w))
	}

	// 3) Momentum ΔJ
	if dJ := j - jp; dJ >= cfg.JMomentumStep {
		score += cfg.ScoreJUp
		signals = append(signals, fmt.Sprintf("KDJ 动能上行(ΔJ≥%.0f)", cfg.JMomentumStep))
	} else if dJ <= -cfg.JMomentumStep {
		score += cfg.ScoreJDown
		signals = append(signals, fmt.Sprintf("KDJ 动能下行(ΔJ≤-%.0f)", cfg.JMomentumStep))
	}

	// 4) K-D spread
	if diffKD := math.Abs(k - d); diffKD >= cfg.KDWideGap {
		if k > d {
			score += cfg.ScoreKDom
			signals = append(signals, fmt.Sprintf("K>D 强势(乖离≥%.0f)", cfg.KDWideGap))
		} else {
			score += cfg.ScoreDDom
			signals = append(signals, fmt.Sprintf("K<D 弱势(乖离≥%.0f)", cfg.KDWideGap))
		}
	}

	// 5) Persistence
	if cfg.PersistenceN > 1 {
		need := cfg.PersistenceN
		begin := i - (need - 1)
		if begin < 0 {
			begin = 0
			need = i + 1
		}
		highCnt, lowCnt := 0, 0
		for t := begin; t <= i; t++ {
			if s.J(t) >= cfg.JOverbought {
				highCnt++
			}
			if s.J(t) <= cfg.JSold {
				lowCnt++
			}
		}
		if highCnt == need && need == cfg.PersistenceN {
			score += cfg.ScoreHighPersist
			signals = append(signals, fmt.Sprintf("KDJ 高位钝化(%d根)", cfg.PersistenceN))
		}
		if lowCnt == need && need == cfg.PersistenceN {
			score += cfg.ScoreLowPersist
			signals = append(signals, fmt.Sprintf("KDJ 低位钝化(%d根)", cfg.PersistenceN))
		}
	}

	// 6) Divergence (simple 3-bar)
	pricesOK := (prices != nil) && (prices.Len() == s.Len())
	if i >= 2 && pricesOK {
		priceHH := prices.At(i) > prices.At(i-1) && prices.At(i-1) > prices.At(i-2)
		jHH := s.J(i) > s.J(i-1) && s.J(i-1) > s.J(i-2)
		priceLL := prices.At(i) < prices.At(i-1) && prices.At(i-1) < prices.At(i-2)
		jLL := s.J(i) < s.J(i-1) && s.J(i-1) < s.J(i-2)

		if priceHH && !jHH {
			score += cfg.ScoreBearDiv
			signals = append(signals, "KDJ 看跌背离(价新高/J未新高)")
		}
		if priceLL && !jLL {
			score += cfg.ScoreBullDiv
			signals = append(signals, "KDJ 看涨背离(价新低/J未新低)")
		}
	}

	return score, signals
}
