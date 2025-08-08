package evaluate

import (
	"errors"
	"go.uber.org/zap"
	"math"
	"sort"
	"wolf_street/pkginit"
)

// =======================================================
// 数据结构定义
// =======================================================

// Engine: 扩展用的引擎数据结构
//   - 当前 StochRSI 单 series 版不依赖它
//   - 如果未来要加 K/D 金叉判断、多周期 HTF 共振，可直接用它
type Engine struct {
	StochRSI    []float64 // 标准化到 0~1 的 StochRSI 数值序列
	StochK      []float64 // StochRSI 对应的 K 值（0~100 或 0~1）
	StochD      []float64 // StochRSI 对应的 D 值（0~100 或 0~1）
	StochRSIHTF []float64 // 高周期 StochRSI（如 4 小时），用于多周期共振判断
}

// EvalResult: 单次评估的返回结果
type EvalResult struct {
	Score      float64            // 总得分
	Signals    []string           // 命中的信号列表（中文描述）
	Components map[string]float64 // 每个子项得分构成，方便调试 & 调参
}

// =======================================================
// 配置结构体（高度可调）
// =======================================================

type StochRSIConfig struct {
	// ---- 基本分层阈值 ----
	SevereOversold   float64 // 严重超卖阈值（默认 0.10）
	Oversold         float64 // 轻度超卖阈值（默认 0.20）
	Overbought       float64 // 轻度超买阈值（默认 0.80）
	SevereOverbought float64 // 严重超买阈值（默认 0.90）

	// ---- 分层权重 ----
	WSevereOS float64 // 严重超卖加分（默认 +2）
	WOS       float64 // 轻度超卖加分（默认 +1）
	WOB       float64 // 轻度超买扣分（默认 -1）
	WSevereOB float64 // 严重超买扣分（默认 -2）

	// ---- 趋势判断 ----
	SlopeLookback int // 计算斜率的回看周期数
	MinRiseBars   int // 底部区域要求连续抬升的 bar 数

	// ---- 金叉/死叉（本单 series 版不启用，但保留配置）----
	EnableCrossBoost    bool    // 是否开启金叉/死叉加分
	CrossoverHysteresis int     // 金叉/死叉信号滞后期
	WCrossUpInOS        float64 // 超卖区金叉加分
	WCrossDownInOB      float64 // 超买区死叉扣分

	// ---- 极值持续与冷却 ----
	PersistBars  int     // 在极值区域持续 N 根后附加分
	WPersistOS   float64 // 超卖持续附加分
	WPersistOB   float64 // 超买持续附加分
	CooldownBars int     // 信号冷却期（调用方外部实现）

	// ---- 多周期共振（本单 series 版不启用）----
	EnableMTF bool
	WMTFOS    float64
	WMTFOB    float64

	// ---- 百分位动态阈值 ----
	UsePercentile    bool                                      // 是否启用动态分位阈值
	PercentileFunc   func(series []float64, p float64) float64 // 分位数计算函数
	PercentileWindow int                                       // 分位数计算窗口
	POversold        float64                                   // 低分位（默认 0.2）
	POverbought      float64                                   // 高分位（默认 0.8）

	// ---- 归一化控制 ----
	NormalizeKD bool // 如果 K/D 是 0~100 则自动归一化到 0~1（series 版未用）
}

// 默认配置生成函数
func DefaultStochRSIConfig() StochRSIConfig {
	return StochRSIConfig{
		SevereOversold:      0.10,
		Oversold:            0.20,
		Overbought:          0.80,
		SevereOverbought:    0.90,
		WSevereOS:           2.0,
		WOS:                 1.0,
		WOB:                 -1.0,
		WSevereOB:           -2.0,
		SlopeLookback:       3,
		MinRiseBars:         2,
		EnableCrossBoost:    false, // 本版默认关闭
		CrossoverHysteresis: 1,
		WCrossUpInOS:        2.0,
		WCrossDownInOB:      -2.0,
		PersistBars:         3,
		WPersistOS:          0.5,
		WPersistOB:          -0.5,
		CooldownBars:        2,
		EnableMTF:           false,
		WMTFOS:              0.5,
		WMTFOB:              -0.5,
		UsePercentile:       false,
		PercentileFunc:      nil,
		PercentileWindow:    100,
		POversold:           0.20,
		POverbought:         0.80,
		NormalizeKD:         true,
	}
}

// =======================================================
// 核心函数（单 series 版 StochRSI 评估）
// =======================================================

func EvaluateStochRSISignals(series []float64, index int, cfg StochRSIConfig) (EvalResult, error) {
	res := EvalResult{
		Score:      0,
		Signals:    []string{},
		Components: map[string]float64{},
	}

	// ---------- 基本检查 ----------
	if len(series) == 0 || index < 0 || index >= len(series) {
		return res, errors.New("invalid series or index")
	}
	// 需要的最少历史长度（取 slopeLookback / minRiseBars / crossoverHysteresis 最大值）
	need := max(1+cfg.SlopeLookback, max(cfg.MinRiseBars, cfg.CrossoverHysteresis))
	if index-need < 0 {
		pkginit.Logger.Debug("EvaluateStochRSISignals: ",
			zap.Any("index", index),
			zap.Any("need", need),
		)
		return res, errors.New("not enough history for evaluation")
	}

	// 当前值
	s := series[index]

	// ---------- 百分位动态阈值 ----------
	os, ob := cfg.Oversold, cfg.Overbought
	if cfg.UsePercentile && cfg.PercentileFunc != nil && cfg.PercentileWindow > 10 {
		start := max(0, index-cfg.PercentileWindow+1)
		window := series[start : index+1]
		os = cfg.PercentileFunc(window, cfg.POversold)
		ob = cfg.PercentileFunc(window, cfg.POverbought)

		// 动态调整严重超卖/超买
		cfg.SevereOversold = clamp01(os * 0.5)
		cfg.SevereOverbought = clamp01((1.0 + ob) / 2)
	}

	// ---------- 1) 分层打分 ----------
	layerScore := 0.0
	switch {
	case s < cfg.SevereOversold:
		layerScore += cfg.WSevereOS
		res.Signals = append(res.Signals, "StochRSI严重超卖")
	case s >= cfg.SevereOversold && s < os:
		layerScore += cfg.WOS
		res.Signals = append(res.Signals, "StochRSI轻度超卖")
	case s > ob && s <= cfg.SevereOverbought:
		layerScore += cfg.WOB
		res.Signals = append(res.Signals, "StochRSI轻度超买")
	case s > cfg.SevereOverbought:
		layerScore += cfg.WSevereOB
		res.Signals = append(res.Signals, "StochRSI严重超买")
	}
	acc(&res, "layer", layerScore)

	// ---------- 2) 底部抬升 ----------
	if s < os {
		rising := isRising(series, index, cfg.MinRiseBars)
		if rising && slope(series, index, cfg.SlopeLookback) > 0 {
			acc(&res, "bottom_rise", 1.0)
			res.Signals = append(res.Signals, "StochRSI底部回升")
		}
	}

	// ---------- 3) 金叉/死叉 ----------
	// series 版无 K/D 数据，此处直接跳过

	// ---------- 4) 极值持续 ----------
	if cfg.PersistBars > 0 {
		if stayedBelow(series, index, os, cfg.PersistBars) {
			acc(&res, "persist_os", cfg.WPersistOS)
			res.Signals = append(res.Signals, "StochRSI超卖持续")
		}
		if stayedAbove(series, index, ob, cfg.PersistBars) {
			acc(&res, "persist_ob", cfg.WPersistOB)
			res.Signals = append(res.Signals, "StochRSI超买持续")
		}
	}

	// 汇总总分
	res.Score = sumComponents(res.Components)
	return res, nil
}

// =======================================================
// 工具函数
// =======================================================

// 判断连续 bars 是否上涨
func isRising(series []float64, idx int, bars int) bool {
	if bars <= 0 {
		return false
	}
	start := idx - bars
	if start < 0 {
		return false
	}
	for i := start; i < idx; i++ {
		if series[i+1] <= series[i] {
			return false
		}
	}
	return true
}

// 简单斜率计算
func slope(series []float64, idx int, lookback int) float64 {
	if lookback <= 0 || idx-lookback < 0 {
		return 0
	}
	return series[idx] - series[idx-lookback]
}

// 判断是否在指定阈值下方持续 bars 根
func stayedBelow(series []float64, idx int, th float64, bars int) bool {
	if bars <= 0 || idx-bars+1 < 0 {
		return false
	}
	for i := idx - bars + 1; i <= idx; i++ {
		if series[i] >= th {
			return false
		}
	}
	return true
}

// 判断是否在指定阈值上方持续 bars 根
func stayedAbove(series []float64, idx int, th float64, bars int) bool {
	if bars <= 0 || idx-bars+1 < 0 {
		return false
	}
	for i := idx - bars + 1; i <= idx; i++ {
		if series[i] <= th {
			return false
		}
	}
	return true
}

// 累加子项分数
func acc(res *EvalResult, key string, v float64) {
	if math.Abs(v) < 1e-12 {
		return
	}
	if res.Components == nil {
		res.Components = map[string]float64{}
	}
	res.Components[key] += v
}

// 汇总总分
func sumComponents(m map[string]float64) float64 {
	sum := 0.0
	for _, v := range m {
		sum += v
	}
	return sum
}

// 限制值到 [0,1] 范围
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// max/min 工具
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =======================================================
// 百分位计算工具（可选）
// =======================================================

// PercentileLinear: 线性插值百分位计算函数
//   - series: 数列
//   - p: 百分位（0~1），如 0.2 表示 20% 分位
func PercentileLinear(series []float64, p float64) float64 {
	if len(series) == 0 {
		return 0
	}
	if p <= 0 {
		return minFloat(series)
	}
	if p >= 1 {
		return maxFloat(series)
	}

	cp := make([]float64, len(series))
	copy(cp, series)
	sort.Float64s(cp)

	pos := (float64(len(cp) - 1)) * p
	l := int(math.Floor(pos))
	u := int(math.Ceil(pos))
	if l == u {
		return cp[l]
	}
	frac := pos - float64(l)
	return cp[l]*(1-frac) + cp[u]*frac
}

func minFloat(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	m := a[0]
	for i := 1; i < len(a); i++ {
		if a[i] < m {
			m = a[i]
		}
	}
	return m
}

func maxFloat(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	m := a[0]
	for i := 1; i < len(a); i++ {
		if a[i] > m {
			m = a[i]
		}
	}
	return m
}
