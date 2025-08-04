package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"go.uber.org/zap"
	"os"
	"strconv"
	"wolf_street/model"
	"wolf_street/pkginit"
	"wolf_street/service"
)

func CliMenuSelectStrategy(size int) (model.Strategy, error) {
	strategies, err := model.GetAllStrategy()
	if err != nil {
		return model.Strategy{}, err
	}

	displayList := make([]string, len(strategies))
	for i, strategy := range strategies {
		displayList[i] = fmt.Sprintf("%d. %s [%s]", strategy.ID, strategy.Name, strategy.Category)
	}

	prompt := promptui.Select{
		Label: "Select Strategy",
		Items: displayList,
		Size:  size, // per page 10
	}

	index, _, err := prompt.Run()
	if err != nil {
		return model.Strategy{}, err
	}

	return strategies[index], nil
}

func CliMenuSelectStock(size int) (model.Stock, error) {
	stocks, err := model.GetAllStock()
	if err != nil {
		return model.Stock{}, err
	}

	displayList := make([]string, len(stocks))
	for i, stock := range stocks {
		displayList[i] = fmt.Sprintf("%s [ %s ]  (%s)", stock.Code, stock.Number, stock.Name)
	}

	prompt := promptui.Select{
		Label: "Select Stock",
		Items: displayList,
		Size:  size, // per page 10
	}

	index, _, err := prompt.Run()
	if err != nil {
		return model.Stock{}, err
	}

	return stocks[index], nil
}

// LoadCandleData loads CSV file for given stock code and returns candle slice
func LoadCandleData(stockCode, stockNumber string) ([]service.Candle, error) {

	if stockCode == "" {
		return nil, errors.New("stock code is empty")
	}

	filePath := "./data_set/" + stockCode + "_" + stockNumber + "_data.csv"
	file, err := os.Open(filePath)
	if err != nil {
		// 判断是否为文件不存在错误
		if errors.Is(err, os.ErrNotExist) {
			pkginit.Logger.Warn("Candle data file not found", zap.String("filePath", filePath))
			return nil, os.ErrNotExist
		}

		// 其他打开文件的错误
		pkginit.Logger.Error("Failed to open candle data file", zap.String("filePath", filePath), zap.Error(err))
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		pkginit.Logger.Error("Failed to read CSV records", zap.Error(err))
		return nil, err
	}

	var candles []service.Candle
	for _, record := range records[1:] { // Skip header row
		closeDate := record[0]
		closePrice, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			pkginit.Logger.Warn("Invalid close price", zap.String("date", closeDate), zap.String("value", record[1]))
			continue
		}

		//pkginit.Logger.Debug("Parsing Candle", zap.String("date", closeDate), zap.Float64("closePrice", closePrice))

		candles = append(candles, service.Candle{
			Date:  closeDate,
			Close: closePrice,
		})
	}

	return candles, nil
}
