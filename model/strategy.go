package model

type Strategy struct {
	ID          int
	Name        string
	Description string
	Category    string // e.g., "Trend Following", "Momentum", "Arbitrage"
}

/*
	Add On:
	- 趋势跟随策略（Trend Following）
		常用指标：
		均线交叉（如 50 日均线上穿 200 日均线）
		动量指标（如 RSI、MACD）
		优缺点：
		优势：牛市中表现突出，持股周期较长
		劣势：震荡市容易频繁止损

	- 事件驱动策略（Event-Driven）(very short & news)
		核心理念：围绕公司事件博弈收益
		典型事件：并购重组 / 财报超预期 / 股东增减持 / 股权激励解禁
		特点：高频交易者、对冲基金常用

*/

func GetAllStrategy() ([]Strategy, error) {
	var Strategies = []Strategy{

		{ID: 1, Name: "StrategyScoringEngine", Category: "StrategyScoringEngine", Description: "StrategyScoringEngine"},
		{ID: 2, Name: "RSI + Bollinger Band", Category: "Mean Reversion", Description: "Combine RSI oversold/overbought with Bollinger Bands."},
		//{ID: 3, Name: "MA Cross (Golden/Death Cross)", Category: "Trend Following", Description: "50-day MA crossing 200-day MA."},
		//{ID: 4, Name: "MACD Cross Strategy", Category: "Momentum", Description: "Trade on MACD line crossovers."},
		//{ID: 5, Name: "Breakout Momentum", Category: "Momentum", Description: "Trade breakout patterns with volume confirmation."},
		//{ID: 6, Name: "Mean Reversion (Donchian Channel)", Category: "Mean Reversion", Description: "Price reverting to Donchian channel median."},
		//{ID: 7, Name: "Multi-Factor Scoring Model", Category: "Factor Investing", Description: "Composite score based on multiple factors."},
		//{ID: 8, Name: "Pairs Trading Arbitrage", Category: "Arbitrage", Description: "Statistical arbitrage with correlated pairs."},
		//{ID: 9, Name: "ML Predictive Strategy", Category: "AI/ML", Description: "Machine Learning based predictive models."},
		//{ID: 10, Name: "Composite Strategy Fusion", Category: "Hybrid", Description: "Fusion of multiple strategies dynamically."},
		//{ID: 11, Name: "Trend Following", Category: "Trend Following", Description: "Ride long-term up/down trends."},
		//{ID: 12, Name: "Momentum Strategy", Category: "Momentum", Description: "Ride short-term price momentum."},
	}

	return Strategies, nil
}
