package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"math"
	"os"
	"wolf_street/model"
	"wolf_street/pkginit"
	"wolf_street/service"
	"wolf_street/util"
)

type DataPoint struct {
	Date       string
	Close      float64
	Gain       float64
	Loss       float64
	AvgGain    float64
	AvgLoss    float64
	RSI        float64
	Signal     string
	RSIComment string
	SMA5       float64
	SMA10      float64
	MA_Cross   string
	Peak       bool
	TopDiv     bool
	BottomDiv  bool
}

func calculateRSI(data []DataPoint, period int) []DataPoint {
	if len(data) <= period {
		return data
	}

	for i := 1; i < len(data); i++ {
		change := data[i].Close - data[i-1].Close
		if change > 0 {
			data[i].Gain = change
			data[i].Loss = 0
		} else {
			data[i].Gain = 0
			data[i].Loss = -change
		}
	}

	var sumGain, sumLoss float64
	for i := 1; i <= period; i++ {
		sumGain += data[i].Gain
		sumLoss += data[i].Loss
	}
	dat := &data[period]
	dat.AvgGain = sumGain / float64(period)
	dat.AvgLoss = sumLoss / float64(period)
	dat.RSI = calcRSI(dat.AvgGain, dat.AvgLoss)
	dat.Signal = interpretSignal(dat.RSI)
	dat.RSIComment = interpretRSIZone(dat.RSI)

	for i := period + 1; i < len(data); i++ {
		data[i].AvgGain = (data[i-1].AvgGain*(float64(period-1)) + data[i].Gain) / float64(period)
		data[i].AvgLoss = (data[i-1].AvgLoss*(float64(period-1)) + data[i].Loss) / float64(period)
		data[i].RSI = calcRSI(data[i].AvgGain, data[i].AvgLoss)
		data[i].Signal = interpretSignal(data[i].RSI)
		data[i].RSIComment = interpretRSIZone(data[i].RSI)
	}

	return data
}

func calcRSI(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

func interpretSignal(rsi float64) string {
	if rsi > 70 {
		return "📉 卖出"
	} else if rsi < 30 {
		return "📈 买入"
	}
	return "⏳ 观望"
}

func interpretRSIZone(rsi float64) string {
	switch {
	case rsi < 20:
		return "极度超卖，抄底观察 ✅"
	case rsi < 30:
		return "超卖区，可考虑买入 ✅"
	case rsi < 50:
		return "偏弱震荡，观望 ⏳"
	case rsi < 70:
		return "偏强震荡，继续持有 🟢"
	case rsi <= 100:
		return "超买区，考虑减仓 📉"
	default:
		return "⚠️ 无效"
	}
}

func calculateSMA(data []DataPoint, period int) []float64 {
	sma := make([]float64, len(data))
	for i := range data {
		if i+1 < period {
			sma[i] = math.NaN()
			continue
		}
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += data[j].Close
		}
		sma[i] = sum / float64(period)
	}
	return sma
}

func detectPeaks(data []DataPoint) []DataPoint {
	for i := 1; i < len(data)-1; i++ {
		if data[i].Close > data[i-1].Close && data[i].Close > data[i+1].Close {
			data[i].Peak = true
		}
	}
	return data
}

func detectDivergence(data []DataPoint) []DataPoint {
	var peaks []int
	var troughs []int

	for i, dp := range data {
		if i == 0 || i == len(data)-1 {
			continue
		}
		if dp.Close > data[i-1].Close && dp.Close > data[i+1].Close {
			peaks = append(peaks, i)
		} else if dp.Close < data[i-1].Close && dp.Close < data[i+1].Close {
			troughs = append(troughs, i)
		}
	}

	for i := 1; i < len(peaks); i++ {
		prev, curr := peaks[i-1], peaks[i]
		if data[curr].Close > data[prev].Close && data[curr].RSI < data[prev].RSI {
			data[curr].TopDiv = true
		}
	}

	for i := 1; i < len(troughs); i++ {
		prev, curr := troughs[i-1], troughs[i]
		if data[curr].Close < data[prev].Close && data[curr].RSI > data[prev].RSI {
			data[curr].BottomDiv = true
		}
	}

	return data
}

func detectMACross(data []DataPoint) []DataPoint {
	for i := 1; i < len(data); i++ {
		if math.IsNaN(data[i-1].SMA5) || math.IsNaN(data[i-1].SMA10) || math.IsNaN(data[i].SMA5) || math.IsNaN(data[i].SMA10) {
			continue
		}
		if data[i-1].SMA5 < data[i-1].SMA10 && data[i].SMA5 >= data[i].SMA10 {
			data[i].MA_Cross = "金叉 📈"
		} else if data[i-1].SMA5 > data[i-1].SMA10 && data[i].SMA5 <= data[i].SMA10 {
			data[i].MA_Cross = "死叉 📉"
		}
	}
	return data
}

func StrategyRSIAndSMA() {
	prices := []float64{
		44.34, 44.09, 44.15, 43.61, 44.33,
		44.83, 45.10, 45.42, 45.84, 46.08,
		45.89, 46.03, 45.61, 46.28, 46.28,
		46.00, 46.03, 46.41, 46.22, 45.64,
		46.21, 46.25, 45.71, 46.45, 45.78,
		45.35, 45.45, 45.01, 44.50, 44.25,
		43.80, 43.95, 44.40, 44.85, 45.30,
		44.95, 44.55, 44.15, 43.90,
	}

	data := make([]DataPoint, len(prices))
	for i := range prices {
		data[i].Close = prices[i]
		data[i].Date = fmt.Sprintf("Day %d", i+1)
	}

	data = calculateRSI(data, 14)
	sma5 := calculateSMA(data, 5)
	sma10 := calculateSMA(data, 10)
	for i := range data {
		data[i].SMA5 = sma5[i]
		data[i].SMA10 = sma10[i]
	}

	data = detectMACross(data)
	data = detectPeaks(data)
	data = detectDivergence(data)

	for _, dp := range data {
		fmt.Printf("%s: Close=%.2f RSI=%.2f [%s]\nExplanation: %s\nSMA5=%.2f SMA10=%.2f %s\nPeak=%v 顶背离=%v 底背离=%v\n\n",
			dp.Date, dp.Close, dp.RSI, dp.Signal, dp.RSIComment, dp.SMA5, dp.SMA10, dp.MA_Cross, dp.Peak, dp.TopDiv, dp.BottomDiv)
	}
}

func main() {
	pkginit.InitLogger() // Init Logger

	app := &cli.App{
		Name:  "Stock Strategy CLI",
		Usage: "Choose strategy and stock to execute backtest",
		Action: func(c *cli.Context) error {
			/* Step 1: Strategy Selection */
			selectedStrategy, err := util.CliMenuSelectStrategy(15)
			if err != nil {
				pkginit.Logger.Error("Strategy selection failed", zap.Error(err))
				return err
			}

			var stock model.Stock
			var candles []service.Candle

			// Step 2 & 3: Stock Selection + Load Candle Data Loop
			for {
				stock, err = util.CliMenuSelectStock(10)
				if err != nil {
					pkginit.Logger.Error("Stock selection failed", zap.Error(err))
					return err
				}

				candles, err = util.LoadCandleData(stock.Code, stock.Number)
				if err != nil {

					pkginit.Logger.Error("LoadCandleData", zap.Error(err))

					// 文件不存在：提示用户重新选股，不退出
					if errors.Is(err, os.ErrNotExist) {
						fmt.Printf("Data file for %s (%s) not found. Please select another stock.\n\n", stock.Name, stock.Code)
						continue // 重新回到股票选择
					}

					// 其他文件异常则退出
					pkginit.Logger.Error("Failed to load candle data", zap.Error(err))
					return err
				}

				// 成功加载Candle数据，退出循环
				break
			}

			// Step 4: Execute Strategy (Placeholder Logic)
			switch selectedStrategy.ID {
			case 1:
				err := service.StrategyScoringEngine(candles)
				//result = service.StrategyRSIBollinger(candles)
				if err != nil {
					pkginit.Logger.Error("Strategy failed:", zap.Any("Strategy", selectedStrategy.Name), zap.Error(err))
				}
			case 2:
				// result = service.StrategyMACross(candles)
			default:
				pkginit.Logger.Error("Strategy not implemented yet")
				return nil
			}

			// Placeholder result output

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		pkginit.Logger.Error("Main error", zap.Error(err))
	}
}
