package spot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/util"
	"github.com/nntaoli-project/goex/v2/websocket"
	"strings"
	"sync"
	"time"
)

// WebSocket 币安现货WebSocket API实现
type WebSocket struct {
	name                string
	ws                  websocket.IWebSocketClient
	baseURL             string
	connected           bool
	mutex               sync.RWMutex
	depthHandlers       map[string]func(*model.Depth)
	tickerHandlers      map[string]func(*model.Ticker)
	klineHandlers       map[string]func([]model.Kline)
	tradeHandlers       map[string]func([]model.Trade)
	errorHandler        func(error)
	connectedHandler    func()
	disconnectedHandler func(error)
	currencyPairM       map[string]model.CurrencyPair
}

// NewWebSocket 创建币安现货WebSocket API
func NewWebSocket() *WebSocket {
	ws := &WebSocket{
		name:           "binance.com",
		baseURL:        "wss://stream.binance.com:9443/ws",
		depthHandlers:  make(map[string]func(*model.Depth)),
		tickerHandlers: make(map[string]func(*model.Ticker)),
		klineHandlers:  make(map[string]func([]model.Kline)),
		tradeHandlers:  make(map[string]func([]model.Trade)),
		currencyPairM:  make(map[string]model.CurrencyPair),
	}

	// 设置WebSocket客户端
	ws.ws = websocket.WsCli

	// 设置消息处理器
	ws.ws.SetHandler("message", ws.handleMessage)

	// 设置错误处理器
	ws.ws.SetErrorHandler(func(err error) {
		logger.Errorf("[Binance] WebSocket error: %v", err)
		if ws.errorHandler != nil {
			ws.errorHandler(err)
		}
	})

	// 设置连接成功处理器
	ws.ws.SetConnectedHandler(func() {
		ws.mutex.Lock()
		ws.connected = true
		ws.mutex.Unlock()

		logger.Info("[Binance] WebSocket connected")
		if ws.connectedHandler != nil {
			ws.connectedHandler()
		}
	})

	// 设置连接断开处理器
	ws.ws.SetDisconnectedHandler(func(err error) {
		ws.mutex.Lock()
		ws.connected = false
		ws.mutex.Unlock()

		logger.Errorf("[Binance] WebSocket disconnected: %v", err)
		if ws.disconnectedHandler != nil {
			ws.disconnectedHandler(err)
		}
	})

	return ws
}

// GetName 获取交易所名称
func (ws *WebSocket) GetName() string {
	return ws.name
}

// Connect 连接到WebSocket服务器
func (ws *WebSocket) Connect() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if ws.connected {
		return errors.New("already connected")
	}

	// 获取交易对信息
	spot := New()
	currencyPairM, _, err := spot.GetExchangeInfo()
	if err != nil {
		return fmt.Errorf("failed to get exchange info: %w", err)
	}
	ws.currencyPairM = currencyPairM

	// 连接到WebSocket服务器
	return ws.ws.Connect(ws.baseURL)
}

// Close 关闭WebSocket连接
func (ws *WebSocket) Close() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !ws.connected {
		return errors.New("not connected")
	}

	return ws.ws.Close()
}

// IsConnected 检查是否已连接
func (ws *WebSocket) IsConnected() bool {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	return ws.connected
}

// SubscribeDepth 订阅深度数据
func (ws *WebSocket) SubscribeDepth(pair model.CurrencyPair, size int, handler func(*model.Depth), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 设置处理器
	ws.mutex.Lock()
	ws.depthHandlers[symbol] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	speed := "100ms" // 默认使用100ms更新速度
	for _, opt := range opts {
		if opt.Key == "speed" {
			speed = opt.Value
		}
	}

	// 根据size选择深度级别
	depthLevel := "20"
	if size <= 5 {
		depthLevel = "5"
	} else if size <= 10 {
		depthLevel = "10"
	} else if size <= 20 {
		depthLevel = "20"
	}

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@depth%s@%s", symbol, depthLevel, speed),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// SubscribeTicker 订阅行情数据
func (ws *WebSocket) SubscribeTicker(pair model.CurrencyPair, handler func(*model.Ticker), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 设置处理器
	ws.mutex.Lock()
	ws.tickerHandlers[symbol] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@ticker", symbol),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// SubscribeKline 订阅K线数据
func (ws *WebSocket) SubscribeKline(pair model.CurrencyPair, period model.KlinePeriod, handler func([]model.Kline), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 设置处理器
	ws.mutex.Lock()
	ws.klineHandlers[symbol+"_"+string(period)] = handler
	ws.mutex.Unlock()

	// 转换K线周期
	interval := adaptKlinePeriod(period)

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@kline_%s", symbol, interval),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// SubscribeTrade 订阅交易数据
func (ws *WebSocket) SubscribeTrade(pair model.CurrencyPair, handler func([]model.Trade), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 设置处理器
	ws.mutex.Lock()
	ws.tradeHandlers[symbol] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@trade", symbol),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// UnsubscribeDepth 取消订阅深度数据
func (ws *WebSocket) UnsubscribeDepth(pair model.CurrencyPair, opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.depthHandlers, symbol)
	ws.mutex.Unlock()

	// 构造取消订阅消息
	speed := "100ms"   // 默认使用100ms更新速度
	depthLevel := "20" // 默认深度级别

	for _, opt := range opts {
		if opt.Key == "speed" {
			speed = opt.Value
		} else if opt.Key == "depth_level" {
			depthLevel = opt.Value
		}
	}

	// 构造取消订阅消息
	unsubMsg := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@depth%s@%s", symbol, depthLevel, speed),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送取消订阅消息
	msgBytes, err := json.Marshal(unsubMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// UnsubscribeTicker 取消订阅行情数据
func (ws *WebSocket) UnsubscribeTicker(pair model.CurrencyPair, opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.tickerHandlers, symbol)
	ws.mutex.Unlock()

	// 构造取消订阅消息
	unsubMsg := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@ticker", symbol),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送取消订阅消息
	msgBytes, err := json.Marshal(unsubMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// UnsubscribeKline 取消订阅K线数据
func (ws *WebSocket) UnsubscribeKline(pair model.CurrencyPair, period model.KlinePeriod, opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.klineHandlers, symbol+"_"+string(period))
	ws.mutex.Unlock()

	// 转换K线周期
	interval := adaptKlinePeriod(period)

	// 构造取消订阅消息
	unsubMsg := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@kline_%s", symbol, interval),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送取消订阅消息
	msgBytes, err := json.Marshal(unsubMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// UnsubscribeTrade 取消订阅交易数据
func (ws *WebSocket) UnsubscribeTrade(pair model.CurrencyPair, opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.tradeHandlers, symbol)
	ws.mutex.Unlock()

	// 构造取消订阅消息
	unsubMsg := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@trade", symbol),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送取消订阅消息
	msgBytes, err := json.Marshal(unsubMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription message: %w", err)
	}

	return ws.ws.SendMessage(msgBytes)
}

// handleMessage 处理接收到的消息
func (ws *WebSocket) handleMessage(message []byte) {
	// 尝试解析消息类型
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		logger.Errorf("[Binance] Failed to unmarshal message: %v", err)
		return
	}

	// 处理订阅确认消息
	if _, ok := msg["result"]; ok {
		logger.Debugf("[Binance] Subscription confirmed: %s", string(message))
		return
	}

	// 处理错误消息
	if errMsg, ok := msg["error"]; ok {
		logger.Errorf("[Binance] Error message: %v", errMsg)
		return
	}

	// 处理数据消息
	if stream, ok := msg["stream"]; ok {
		streamStr := stream.(string)
		data := msg["data"]

		// 解析流类型和交易对
		parts := strings.Split(streamStr, "@")
		if len(parts) < 2 {
			logger.Errorf("[Binance] Invalid stream format: %s", streamStr)
			return
		}

		symbol := parts[0]
		streamType := parts[1]

		// 根据流类型处理数据
		switch {
		case strings.HasPrefix(streamType, "depth"):
			ws.handleDepthMessage(symbol, data)
		case streamType == "ticker":
			ws.handleTickerMessage(symbol, data)
		case strings.HasPrefix(streamType, "kline"):
			ws.handleKlineMessage(symbol, streamType, data)
		case streamType == "trade":
			ws.handleTradeMessage(symbol, data)
		default:
			logger.Debugf("[Binance] Unhandled stream type: %s", streamType)
		}
	} else if e, ok := msg["e"]; ok {
		// 直接推送的数据格式
		eventType := e.(string)
		switch eventType {
		case "depthUpdate":
			ws.handleDepthUpdateMessage(msg)
		case "24hrTicker":
			ws.handleTickerUpdateMessage(msg)
		case "kline":
			ws.handleKlineUpdateMessage(msg)
		case "trade":
			ws.handleTradeUpdateMessage(msg)
		default:
			logger.Debugf("[Binance] Unhandled event type: %s", eventType)
		}
	}
}

// handleDepthMessage 处理深度数据消息
func (ws *WebSocket) handleDepthMessage(symbol string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance] Failed to marshal depth data: %v", err)
		return
	}

	// 解析深度数据
	var depthData struct {
		LastUpdateID int64      `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
		Symbol       string     `json:"s"`
		EventTime    int64      `json:"E"`
	}

	if err := json.Unmarshal(dataBytes, &depthData); err != nil {
		logger.Errorf("[Binance] Failed to unmarshal depth data: %v", err)
		return
	}

	// 创建深度对象
	depth := &model.Depth{
		UTime: time.Unix(0, depthData.EventTime*int64(time.Millisecond)),
		Asks:  make(model.DepthItems, 0, len(depthData.Asks)),
		Bids:  make(model.DepthItems, 0, len(depthData.Bids)),
	}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[strings.ToUpper(symbol)]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance] Unknown symbol: %s", symbol)
		return
	}
	depth.Pair = pair

	// 解析买单
	for _, bid := range depthData.Bids {
		price, err := util.ToFloat64(bid[0])
		if err != nil {
			continue
		}
		amount, err := util.ToFloat64(bid[1])
		if err != nil {
			continue
		}
		depth.Bids = append(depth.Bids, model.DepthItem{Price: price, Amount: amount})
	}

	// 解析卖单
	for _, ask := range depthData.Asks {
		price, err := util.ToFloat64(ask[0])
		if err != nil {
			continue
		}
		amount, err := util.ToFloat64(ask[1])
		if err != nil {
			continue
		}
		depth.Asks = append(depth.Asks, model.DepthItem{Price: price, Amount: amount})
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.depthHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(depth)
	}
}

// handleDepthUpdateMessage 处理深度更新消息
func (ws *WebSocket) handleDepthUpdateMessage(msg map[string]interface{}) {
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
		logger.Errorf("[Binance] Unknown symbol: %s", symbol)
		return
	}
	depth.Pair = pair

	// 解析买单
	for _, item := range bidsData {
		bid, ok := item.([]interface{})
		if !ok || len(bid) < 2 {
			continue
		}
		price, err := util.ToFloat64(bid[0])
		if err != nil {
			continue
		}
		amount, err := util.ToFloat64(bid[1])
		if err != nil {
			continue
		}
		depth.Bids = append(depth.Bids, model.DepthItem{Price: price, Amount: amount})
	}

	// 解析卖单
	for _, item := range asksData {
		ask, ok := item.([]interface{})
		if !ok || len(ask) < 2 {
			continue
		}
		price, err := util.ToFloat64(ask[0])
		if err != nil {
			continue
		}
		amount, err := util.ToFloat64(ask[1])
		if err != nil {
			continue
		}
		depth.Asks = append(depth.Asks, model.DepthItem{Price: price, Amount: amount})
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.depthHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler(depth)
	}
}

// handleTickerMessage 处理行情数据消息
func (ws *WebSocket) handleTickerMessage(symbol string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance] Failed to marshal ticker data: %v", err)
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
		logger.Errorf("[Binance] Failed to unmarshal ticker data: %v", err)
		return
	}

	// 创建行情对象
	ticker := &model.Ticker{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[tickerData.Symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance] Unknown symbol: %s", tickerData.Symbol)
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
func (ws *WebSocket) handleTickerUpdateMessage(msg map[string]interface{}) {
	// 解析行情数据
	symbol, _ := msg["s"].(string)

	// 创建行情对象
	ticker := &model.Ticker{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance] Unknown symbol: %s", symbol)
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
func (ws *WebSocket) handleKlineMessage(symbol string, streamType string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance] Failed to marshal kline data: %v", err)
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
		logger.Errorf("[Binance] Failed to unmarshal kline data: %v", err)
		return
	}

	// 提取K线周期
	parts := strings.Split(streamType, "_")
	if len(parts) < 2 {
		logger.Errorf("[Binance] Invalid kline stream format: %s", streamType)
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
		logger.Errorf("[Binance] Unknown symbol: %s", klineData.Symbol)
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
func (ws *WebSocket) handleKlineUpdateMessage(msg map[string]interface{}) {
	// 解析K线数据
	symbol, _ := msg["s"].(string)
	kData, ok := msg["k"].(map[string]interface{})
	if !ok {
		logger.Errorf("[Binance] Invalid kline data format")
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
		logger.Errorf("[Binance] Unknown symbol: %s", symbol)
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

// handleTradeMessage 处理交易数据消息
func (ws *WebSocket) handleTradeMessage(symbol string, data interface{}) {
	// 将data转换为JSON字节
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("[Binance] Failed to marshal trade data: %v", err)
		return
	}

	// 解析交易数据
	var tradeData struct {
		EventType          string `json:"e"`
		EventTime          int64  `json:"E"`
		Symbol             string `json:"s"`
		TradeID            int64  `json:"t"`
		Price              string `json:"p"`
		Quantity           string `json:"q"`
		BuyerOrderID       int64  `json:"b"`
		SellerOrderID      int64  `json:"a"`
		TradeTime          int64  `json:"T"`
		IsBuyerMarketMaker bool   `json:"m"`
		IsIgnore           bool   `json:"M"`
	}

	if err := json.Unmarshal(dataBytes, &tradeData); err != nil {
		logger.Errorf("[Binance] Failed to unmarshal trade data: %v", err)
		return
	}

	// 创建交易对象
	trade := model.Trade{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[tradeData.Symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance] Unknown symbol: %s", tradeData.Symbol)
		return
	}
	trade.Pair = pair

	// 设置交易数据
	trade.Tid = fmt.Sprintf("%d", tradeData.TradeID)
	trade.Price, _ = util.ToFloat64(tradeData.Price)
	trade.Amount, _ = util.ToFloat64(tradeData.Quantity)
	trade.Timestamp = tradeData.TradeTime

	// 设置交易方向
	if tradeData.IsBuyerMarketMaker {
		trade.Direction = "sell"
	} else {
		trade.Direction = "buy"
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.tradeHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler([]model.Trade{trade})
	}
}

// handleTradeUpdateMessage 处理交易更新消息
func (ws *WebSocket) handleTradeUpdateMessage(msg map[string]interface{}) {
	// 解析交易数据
	symbol, _ := msg["s"].(string)
	tradeID, _ := msg["t"].(float64)
	price, _ := msg["p"].(string)
	quantity, _ := msg["q"].(string)
	tradeTime, _ := msg["T"].(float64)
	isBuyerMarketMaker, _ := msg["m"].(bool)

	// 创建交易对象
	trade := model.Trade{}

	// 设置交易对
	ws.mutex.RLock()
	pair, ok := ws.currencyPairM[symbol]
	ws.mutex.RUnlock()
	if !ok {
		logger.Errorf("[Binance] Unknown symbol: %s", symbol)
		return
	}
	trade.Pair = pair

	// 设置交易数据
	trade.Tid = fmt.Sprintf("%d", int64(tradeID))
	trade.Price, _ = util.ToFloat64(price)
	trade.Amount, _ = util.ToFloat64(quantity)
	trade.Timestamp = int64(tradeTime)

	// 设置交易方向
	if isBuyerMarketMaker {
		trade.Direction = "sell"
	} else {
		trade.Direction = "buy"
	}

	// 调用处理器
	ws.mutex.RLock()
	handler, ok := ws.tradeHandlers[strings.ToLower(symbol)]
	ws.mutex.RUnlock()

	if ok && handler != nil {
		handler([]model.Trade{trade})
	}
}

// reverseAdaptKlinePeriod 将币安K线周期转换为KlinePeriod
func reverseAdaptKlinePeriod(interval string) model.KlinePeriod {
	switch interval {
	case "1m":
		return model.KLINE_PERIOD_1MIN
	case "3m":
		return model.KLINE_PERIOD_3MIN
	case "5m":
		return model.KLINE_PERIOD_5MIN
	case "15m":
		return model.KLINE_PERIOD_15MIN
	case "30m":
		return model.KLINE_PERIOD_30MIN
	case "1h":
		return model.KLINE_PERIOD_60MIN
	case "2h":
		return model.KLINE_PERIOD_2H
	case "4h":
		return model.KLINE_PERIOD_4H
	case "6h":
		return model.KLINE_PERIOD_6H
	case "8h":
		return model.KLINE_PERIOD_8H
	case "12h":
		return model.KLINE_PERIOD_12H
	case "1d":
		return model.KLINE_PERIOD_1DAY
	case "3d":
		return model.KLINE_PERIOD_3DAY
	case "1w":
		return model.KLINE_PERIOD_1WEEK
	case "1M":
		return model.KLINE_PERIOD_1MONTH
	default:
		return model.KLINE_PERIOD_1DAY
	}
}
