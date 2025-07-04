package fapi

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/util"
)

// handleMessage 处理接收到的消息
func (ws *WebSocketBase) handleMessage(msg []byte) error {
	// 解析消息
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		logger.Errorf("[Binance Futures] Failed to unmarshal message: %v", err)
		return err
	}

	// 检查是否为错误消息
	if code, ok := data["code"].(float64); ok {
		msg, _ := data["msg"].(string)
		err := fmt.Errorf("[Binance Futures] Error: code=%v, msg=%s", code, msg)
		logger.Error(err)
		return err
	}

	// 根据消息类型分发处理
	if stream, ok := data["stream"].(string); ok {
		// 公共频道消息
		parts := strings.Split(stream, "@")
		if len(parts) < 2 {
			return fmt.Errorf("[Binance Futures] Invalid stream format: %s", stream)
		}

		symbol := parts[0]
		streamType := parts[1]

		data, _ := data["data"].(map[string]interface{})

		// 根据流类型处理消息
		if strings.HasPrefix(streamType, "depth") {
			// 深度数据
			return ws.handleDepthUpdateMessage(data)
		} else if streamType == "ticker" {
			// 行情数据
			ws.handleTickerMessage(symbol, data)
		} else if strings.HasPrefix(streamType, "kline") {
			// K线数据
			ws.handleKlineMessage(symbol, streamType, data)
		} else if streamType == "trade" {
			// 交易数据
			ws.handleTradeMessage(symbol, data)
		} else if streamType == "markPrice" {
			// 资金费率
			ws.handleFundingRateMessage(symbol, data)
		}
	} else if e, ok := data["e"].(string); ok {
		// 私有频道消息
		switch e {
		// 订单更新
		case "ORDER_TRADE_UPDATE":
			ws.handleOrderUpdateMessage(data)
		// 账户更新
		case "ACCOUNT_UPDATE":
			ws.handleAccountUpdateMessage(data)
		// 深度更新
		case "depthUpdate":
			ws.handleDepthUpdateMessage(data)
		// 行情更新
		case "24hrTicker":
			ws.handleTickerUpdateMessage(data)
		// K线更新
		case "kline":
			ws.handleKlineUpdateMessage(data)
		// 交易更新
		case "aggTrade":
			ws.handleTradeUpdateMessage(data)
		// 资金费率更新
		case "markPriceUpdate":
			ws.handleFundingRateUpdateMessage(data)
		}
	}

	return nil
}

// handleDepthMessage 处理深度数据消息
func (ws *WebSocketBase) handleDepthMessage(symbol string, data interface{}) error {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance Futures] Failed to marshal depth data: %v", err)
		return err
	}

	// 解析深度数据
	var depthData struct {
		LastUpdateID int64           `json:"lastUpdateId"`
		Bids         [][]interface{} `json:"bids"`
		Asks         [][]interface{} `json:"asks"`
	}

	if err := json.Unmarshal(dataBytes, &depthData); err != nil {
		logger.Errorf("[Binance Futures] Failed to unmarshal depth data: %v", err)
		return err
	}

	// 创建深度对象
	depth := &model.Depth{
		UTime: time.Now(),
		Asks:  make(model.DepthItems, 0, len(depthData.Asks)),
		Bids:  make(model.DepthItems, 0, len(depthData.Bids)),
	}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		err := fmt.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		logger.Error(err)
		return err
	}
	depth.Pair = pair

	// 解析买单
	for _, item := range depthData.Bids {
		if len(item) < 2 {
			continue
		}
		price, _ := util.ToFloat64(item[0])
		amount, _ := util.ToFloat64(item[1])
		depth.Bids = append(depth.Bids, model.DepthItem{Price: price, Amount: amount})
	}

	// 解析卖单
	for _, item := range depthData.Asks {
		if len(item) < 2 {
			continue
		}
		price, _ := util.ToFloat64(item[0])
		amount, _ := util.ToFloat64(item[1])
		depth.Asks = append(depth.Asks, model.DepthItem{Price: price, Amount: amount})
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.depthHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(depth)
	}

	return nil
}

// handleDepthUpdateMessage 处理深度更新消息
func (ws *WebSocketBase) handleDepthUpdateMessage(msg map[string]interface{}) error {
	// 解析深度数据
	symbol, _ := msg["s"].(string)
	eventTime, _ := msg["E"].(float64)

	// 解析买单和卖单
	bidsData, _ := msg["b"].([]interface{})
	asksData, _ := msg["a"].([]interface{})

	// 创建深度对象
	depth := &model.Depth{
		UTime: time.Unix(0, int64(eventTime)*int64(time.Millisecond)),
		Asks:  make(model.DepthItems, 0, len(asksData)),
		Bids:  make(model.DepthItems, 0, len(bidsData)),
	}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		err := fmt.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		logger.Error(err)
		return err
	}
	depth.Pair = pair

	// 解析买单
	for _, item := range bidsData {
		bid, ok := item.([]interface{})
		if !ok || len(bid) < 2 {
			continue
		}
		price, _ := util.ToFloat64(bid[0])
		amount, _ := util.ToFloat64(bid[1])
		depth.Bids = append(depth.Bids, model.DepthItem{Price: price, Amount: amount})
	}

	// 解析卖单
	for _, item := range asksData {
		ask, ok := item.([]interface{})
		if !ok || len(ask) < 2 {
			continue
		}
		price, _ := util.ToFloat64(ask[0])
		amount, _ := util.ToFloat64(ask[1])
		depth.Asks = append(depth.Asks, model.DepthItem{Price: price, Amount: amount})
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.depthHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(depth)
	}

	return nil
}

// handleOrderUpdateMessage 处理订单更新消息
func (ws *WebSocketBase) handleOrderUpdateMessage(msg map[string]interface{}) {
	// 解析订单数据
	data, ok := msg["o"].(map[string]interface{})
	if !ok {
		logger.Errorf("[Binance Futures] Invalid order update data format")
		return
	}

	// 创建订单对象
	order := &model.Order{}

	// 设置交易对
	symbol, _ := data["s"].(string)
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		return
	}
	order.Pair = pair

	// 设置订单ID
	order.CId, _ = data["c"].(string)
	orderId, _ := util.ToInt64(data["i"])
	order.Id = fmt.Sprintf("%d", orderId)

	// 设置价格和数量
	order.Price, _ = util.ToFloat64(data["p"])
	order.Qty, _ = util.ToFloat64(data["q"])
	order.ExecutedQty, _ = util.ToFloat64(data["z"])
	order.PriceAvg, _ = util.ToFloat64(data["ap"])

	// 设置订单类型
	orderType, _ := data["o"].(string)
	switch orderType {
	case "LIMIT":
		order.OrderTy = model.LIMIT
	case "MARKET":
		order.OrderTy = model.MARKET
	case "STOP":
		order.OrderTy = model.STOP
	case "STOP_MARKET":
		order.OrderTy = model.STOP_MARKET
	default:
		order.OrderTy = model.LIMIT
	}

	// 设置订单方向
	side, _ := data["S"].(string)
	if side == "BUY" {
		order.Side = model.BUY
	} else {
		order.Side = model.SELL
	}

	// 设置订单状态
	status, _ := data["X"].(string)
	switch status {
	case "NEW":
		order.Status = model.ORDER_UNFINISH
	case "PARTIALLY_FILLED":
		order.Status = model.ORDER_PART_FINISH
	case "FILLED":
		order.Status = model.ORDER_FINISH
	case "CANCELED":
		order.Status = model.ORDER_CANCEL
	case "REJECTED":
		order.Status = model.ORDER_REJECT
	case "EXPIRED":
		order.Status = model.ORDER_CANCEL
	default:
		order.Status = model.ORDER_UNFINISH
	}

	// 设置订单时间
	order.CreatedAt, _ = util.ToInt64(data["T"])

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.orderHandlers["order"]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(order)
	}
}

// handleAccountUpdateMessage 处理账户更新消息
func (ws *WebSocketBase) handleAccountUpdateMessage(msg map[string]interface{}) {
	// 解析账户数据
	data, ok := msg["a"].(map[string]interface{})
	if !ok {
		logger.Errorf("[Binance Futures] Invalid account update data format")
		return
	}

	// 解析持仓更新
	positions, ok := data["P"].([]interface{})
	if ok && len(positions) > 0 {
		// 创建持仓列表
		positionList := make([]model.FuturesPosition, 0, len(positions))

		// 解析每个持仓
		for _, p := range positions {
			pos, ok := p.(map[string]interface{})
			if !ok {
				continue
			}

			// 创建持仓对象
			position := model.FuturesPosition{}

			// 设置交易对
			symbol, _ := pos["s"].(string)
			ws.mutex.RLock()
			pair, ok := ws.currencyPairM[symbol]
			ws.mutex.RUnlock()
			if !ok {
				logger.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
				continue
			}
			position.Pair = pair

			// 设置持仓数据
			position.Qty, _ = util.ToFloat64(pos["pa"])
			position.AvgPx, _ = util.ToFloat64(pos["ep"])
			lever, _ := util.ToInt(pos["pa"])
			position.Lever = float64(lever)
			position.UplRatio, _ = util.ToFloat64(pos["up"])
			position.Upl, _ = util.ToFloat64(pos["up"])

			// 设置持仓方向
			side, _ := pos["ps"].(string)
			if side == "LONG" {
				position.PosSide = model.BUY
			} else {
				position.PosSide = model.SELL
			}

			// 添加到持仓列表
			positionList = append(positionList, position)
		}

		// 调用持仓处理器
		ws.mutex.RLock()
		handler, ok := ws.positionHandlers["position"]
		ws.mutex.RUnlock()

		if ok && handler != nil && len(positionList) > 0 {
			handler(positionList)
		}
	}

	// 解析账户余额更新
	balances, ok := data["B"].([]interface{})
	if ok && len(balances) > 0 {
		// 创建账户映射
		accountMap := make(map[string]model.FuturesAccount)

		// 解析每个余额
		for _, b := range balances {
			bal, ok := b.(map[string]interface{})
			if !ok {
				continue
			}

			// 获取币种
			asset, _ := bal["a"].(string)

			// 创建账户对象
			account := model.FuturesAccount{}

			// 设置账户数据
			account.Coin = asset
			account.Eq, _ = util.ToFloat64(bal["wb"])
			account.AvailEq, _ = util.ToFloat64(bal["cw"])
			account.FrozenBal = account.Eq - account.AvailEq

			// 添加到账户映射
			accountMap[asset] = account
		}

		// 调用账户处理器
		ws.mutex.RLock()
		handler, ok := ws.accountHandlers["account"]
		ws.mutex.RUnlock()

		if ok && handler != nil && len(accountMap) > 0 {
			handler(accountMap)
		}
	}
}

// handleTradeMessage 处理交易数据消息
func (ws *WebSocketBase) handleTradeMessage(symbol string, data interface{}) error {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance Futures] Failed to marshal trade data: %v", err)
		return err
	}

	// 解析交易数据
	var tradeData struct {
		EventType        string `json:"e"`
		EventTime        int64  `json:"E"`
		Symbol           string `json:"s"`
		AggregateTradeID int64  `json:"a"`
		Price            string `json:"p"`
		Quantity         string `json:"q"`
		FirstTradeID     int64  `json:"f"`
		LastTradeID      int64  `json:"l"`
		TradeTime        int64  `json:"T"`
		IsBuyerMaker     bool   `json:"m"`
	}

	if err := json.Unmarshal(dataBytes, &tradeData); err != nil {
		logger.Errorf("[Binance Futures] Failed to unmarshal trade data: %v", err)
		return err
	}

	// 创建交易对象
	trade := model.Trade{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[tradeData.Symbol]
	ws.mutex.RUnlock()
	if !ok {
		err := fmt.Errorf("[Binance Futures] Unknown symbol: %s", tradeData.Symbol)
		logger.Error(err)
		return err
	}
	trade.Pair = pair

	// 设置交易数据
	trade.Tid = fmt.Sprintf("%d", tradeData.AggregateTradeID)
	trade.Price, _ = util.ToFloat64(tradeData.Price)
	trade.Amount, _ = util.ToFloat64(tradeData.Quantity)
	trade.Timestamp = tradeData.TradeTime
	if tradeData.IsBuyerMaker {
		trade.Direction = string(model.SELL)
	} else {
		trade.Direction = string(model.BUY)
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.tradeHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler([]model.Trade{trade})
	}

	return nil
}

// handleTradeUpdateMessage 处理交易更新消息
func (ws *WebSocketBase) handleTradeUpdateMessage(msg map[string]interface{}) {
	// 解析交易数据
	symbol, _ := msg["s"].(string)
	aggTradeID, _ := util.ToInt64(msg["a"])
	price, _ := util.ToFloat64(msg["p"])
	quantity, _ := util.ToFloat64(msg["q"])
	tradeTime, _ := util.ToInt64(msg["T"])
	isBuyerMaker, _ := msg["m"].(bool)

	// 创建交易对象
	trade := model.Trade{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		return
	}
	trade.Pair = pair

	// 设置交易数据
	trade.Tid = fmt.Sprintf("%d", aggTradeID)
	trade.Price = price
	trade.Amount = quantity
	trade.Timestamp = tradeTime
	if isBuyerMaker {
		trade.Direction = string(model.SELL)
	} else {
		trade.Direction = string(model.BUY)
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.tradeHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler([]model.Trade{trade})
	}
}

// handleFundingRateMessage 处理资金费率消息
func (ws *WebSocketBase) handleFundingRateMessage(symbol string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance Futures] Failed to marshal funding rate data: %v", err)
		return
	}

	// 解析资金费率数据
	var fundingRateData struct {
		EventType            string `json:"e"`
		EventTime            int64  `json:"E"`
		Symbol               string `json:"s"`
		MarkPrice            string `json:"p"`
		IndexPrice           string `json:"i"`
		EstimatedSettlePrice string `json:"P"`
		FundingRate          string `json:"r"`
		NextFundingTime      int64  `json:"T"`
	}

	if err := json.Unmarshal(dataBytes, &fundingRateData); err != nil {
		logger.Errorf("[Binance Futures] Failed to unmarshal funding rate data: %v", err)
		return
	}

	// 创建资金费率对象
	fundingRate := &model.FundingRate{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[fundingRateData.Symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", fundingRateData.Symbol)
		return
	}
	fundingRate.Symbol = pair.Symbol

	// 设置资金费率数据
	fundingRate.Rate, _ = util.ToFloat64(fundingRateData.FundingRate)
	fundingRate.Tm = fundingRateData.EventTime

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.fundingRateHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(fundingRate)
	}
}

// handleFundingRateUpdateMessage 处理资金费率更新消息
func (ws *WebSocketBase) handleFundingRateUpdateMessage(msg map[string]interface{}) {
	// 解析资金费率数据
	symbol, _ := msg["s"].(string)
	eventTime, _ := util.ToInt64(msg["E"])
	fundingRate, _ := util.ToFloat64(msg["r"])
	// nextFundingTime可以在需要时使用
	// nextFundingTime, _ := util.ToInt64(msg["T"])

	// 创建资金费率对象
	fr := &model.FundingRate{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		return
	}
	fr.Symbol = pair.Symbol

	// 设置资金费率数据
	fr.Rate = fundingRate
	fr.Tm = eventTime

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.fundingRateHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(fr)
	}
}

// handleTickerMessage 处理行情数据消息
func (ws *WebSocketBase) handleTickerMessage(symbol string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance Futures] Failed to marshal ticker data: %v", err)
		return
	}

	// 解析行情数据
	var tickerData struct {
		Symbol             string `json:"s"`
		PriceChange        string `json:"p"`
		PriceChangePercent string `json:"P"`
		WeightedAvgPrice   string `json:"w"`
		LastPrice          string `json:"c"`
		LastQty            string `json:"Q"`
		BidPrice           string `json:"b"`
		BidQty             string `json:"B"`
		AskPrice           string `json:"a"`
		AskQty             string `json:"A"`
		OpenPrice          string `json:"o"`
		HighPrice          string `json:"h"`
		LowPrice           string `json:"l"`
		Volume             string `json:"v"`
		QuoteVolume        string `json:"q"`
		OpenTime           int64  `json:"O"`
		CloseTime          int64  `json:"C"`
		FirstId            int64  `json:"F"`
		LastId             int64  `json:"L"`
		Count              int64  `json:"n"`
	}

	if err := json.Unmarshal(dataBytes, &tickerData); err != nil {
		logger.Errorf("[Binance Futures] Failed to unmarshal ticker data: %v", err)
		return
	}

	// 创建行情对象
	ticker := &model.Ticker{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[tickerData.Symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", tickerData.Symbol)
		return
	}
	ticker.Pair = pair

	// 设置行情数据
	ticker.Last, _ = util.ToFloat64(tickerData.LastPrice)
	ticker.Buy, _ = util.ToFloat64(tickerData.BidPrice)
	ticker.Sell, _ = util.ToFloat64(tickerData.AskPrice)
	ticker.High, _ = util.ToFloat64(tickerData.HighPrice)
	ticker.Low, _ = util.ToFloat64(tickerData.LowPrice)
	ticker.Vol, _ = util.ToFloat64(tickerData.Volume)
	ticker.Percent, _ = util.ToFloat64(tickerData.PriceChangePercent)
	ticker.Timestamp = tickerData.CloseTime

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.tickerHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(ticker)
	}
}

// handleTickerUpdateMessage 处理行情更新消息
func (ws *WebSocketBase) handleTickerUpdateMessage(msg map[string]interface{}) {
	// 解析行情数据
	symbol, _ := msg["s"].(string)

	// 创建行情对象
	ticker := &model.Ticker{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		return
	}
	ticker.Pair = pair

	// 设置行情数据
	ticker.Last, _ = util.ToFloat64(msg["c"])
	ticker.Buy, _ = util.ToFloat64(msg["b"])
	ticker.Sell, _ = util.ToFloat64(msg["a"])
	ticker.High, _ = util.ToFloat64(msg["h"])
	ticker.Low, _ = util.ToFloat64(msg["l"])
	ticker.Vol, _ = util.ToFloat64(msg["v"])
	ticker.Percent, _ = util.ToFloat64(msg["P"])
	ticker.Timestamp, _ = util.ToInt64(msg["C"])

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.tickerHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(ticker)
	}
}

// handleKlineMessage 处理K线数据消息
func (ws *WebSocketBase) handleKlineMessage(symbol string, streamType string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance Futures] Failed to marshal kline data: %v", err)
		return
	}

	// 解析K线数据
	var klineData struct {
		Symbol string `json:"s"`
		Kline  struct {
			StartTime             int64  `json:"t"`
			EndTime               int64  `json:"T"`
			Symbol                string `json:"s"`
			Interval              string `json:"i"`
			FirstTradeID          int64  `json:"f"`
			LastTradeID           int64  `json:"L"`
			OpenPrice             string `json:"o"`
			ClosePrice            string `json:"c"`
			HighPrice             string `json:"h"`
			LowPrice              string `json:"l"`
			BaseAssetVolume       string `json:"v"`
			NumberOfTrades        int64  `json:"n"`
			IsClosed              bool   `json:"x"`
			QuoteAssetVolume      string `json:"q"`
			TakerBuyBaseAssetVol  string `json:"V"`
			TakerBuyQuoteAssetVol string `json:"Q"`
		} `json:"k"`
		EventTime int64 `json:"E"`
	}

	if err := json.Unmarshal(dataBytes, &klineData); err != nil {
		logger.Errorf("[Binance Futures] Failed to unmarshal kline data: %v", err)
		return
	}

	// 提取K线周期
	parts := strings.Split(streamType, "_")
	if len(parts) < 2 {
		logger.Errorf("[Binance Futures] Invalid kline stream format: %s", streamType)
		return
	}
	interval := parts[1]

	// 转换为KlinePeriod
	period := reverseAdaptKlinePeriod(interval)

	// 创建K线对象
	kline := model.Kline{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[klineData.Symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", klineData.Symbol)
		return
	}
	kline.Pair = pair

	// 设置K线数据
	kline.Timestamp = klineData.Kline.StartTime
	kline.Open, _ = util.ToFloat64(klineData.Kline.OpenPrice)
	kline.Close, _ = util.ToFloat64(klineData.Kline.ClosePrice)
	kline.High, _ = util.ToFloat64(klineData.Kline.HighPrice)
	kline.Low, _ = util.ToFloat64(klineData.Kline.LowPrice)
	kline.Vol, _ = util.ToFloat64(klineData.Kline.BaseAssetVolume)

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.klineHandlers[strings.ToLower(symbol)+"_"+string(period)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler([]model.Kline{kline})
	}
}

// handleKlineUpdateMessage 处理K线更新消息
func (ws *WebSocketBase) handleKlineUpdateMessage(msg map[string]interface{}) {
	// 解析K线数据
	symbol, _ := msg["s"].(string)
	kData, ok := msg["k"].(map[string]interface{})
	if !ok {
		logger.Errorf("[Binance Futures] Invalid kline data format")
		return
	}

	// 提取K线周期
	interval, _ := kData["i"].(string)

	// 转换为KlinePeriod
	period := reverseAdaptKlinePeriod(interval)

	// 创建K线对象
	kline := model.Kline{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance Futures] Unknown symbol: %s", symbol)
		return
	}
	kline.Pair = pair

	// 设置K线数据
	kline.Timestamp, _ = util.ToInt64(kData["t"])
	kline.Open, _ = util.ToFloat64(kData["o"])
	kline.Close, _ = util.ToFloat64(kData["c"])
	kline.High, _ = util.ToFloat64(kData["h"])
	kline.Low, _ = util.ToFloat64(kData["l"])
	kline.Vol, _ = util.ToFloat64(kData["v"])

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.klineHandlers[strings.ToLower(symbol)+"_"+string(period)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler([]model.Kline{kline})
	}
}

// reverseAdaptKlinePeriod 函数已在 websocket.go 中定义，不需要在此重复定义
