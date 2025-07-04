package fapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/util"
	"strings"
)

// adaptKlinePeriod 将模型的KlinePeriod转换为Binance的K线周期
func adaptKlinePeriod(period model.KlinePeriod) string {
	switch period {
	case model.Kline_1min:
		return "1m"
	case model.KLINE_PERIOD_3MIN:
		return "3m"
	case model.Kline_5min:
		return "5m"
	case model.Kline_15min:
		return "15m"
	case model.Kline_30min:
		return "30m"
	case model.Kline_1h:
		return "1h"
	case model.KLINE_PERIOD_2H:
		return "2h"
	case model.Kline_4h:
		return "4h"
	case model.Kline_6h:
		return "6h"
	case model.KLINE_PERIOD_8H:
		return "8h"
	case model.KLINE_PERIOD_12H:
		return "12h"
	case model.Kline_1day:
		return "1d"
	case model.KLINE_PERIOD_3DAY:
		return "3d"
	case model.Kline_1week:
		return "1w"
	case model.KLINE_PERIOD_1MONTH:
		return "1M"
	default:
		return string(period)
	}
}

// SubscribeDepth 订阅深度数据
func (ws *WebSocketBase) SubscribeDepth(pair model.CurrencyPair, size int, handler func(*model.Depth), opts ...model.OptionParameter) error {
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
func (ws *WebSocketBase) SubscribeTicker(pair model.CurrencyPair, handler func(*model.Ticker), opts ...model.OptionParameter) error {
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
func (ws *WebSocketBase) SubscribeKline(pair model.CurrencyPair, period model.KlinePeriod, handler func([]model.Kline), opts ...model.OptionParameter) error {
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
func (ws *WebSocketBase) SubscribeTrade(pair model.CurrencyPair, handler func([]model.Trade), opts ...model.OptionParameter) error {
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
			fmt.Sprintf("%s@aggTrade", symbol),
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

// SubscribeFundingRate 订阅资金费率
func (ws *WebSocketBase) SubscribeFundingRate(pair model.CurrencyPair, handler func(*model.FundingRate), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 设置处理器
	ws.mutex.Lock()
	ws.fundingRateHandlers[symbol] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@markPrice", symbol),
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

// SubscribeOrder 订阅订单更新
func (ws *WebSocketBase) SubscribeOrder(handler func(*model.Order), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	if ws.apiKey == "" || ws.apiSecret == "" {
		return errors.New("API key and secret are required for private WebSocket")
	}

	// 获取listenKey
	err := ws.getListenKey()
	if err != nil {
		return err
	}

	// 设置处理器
	ws.mutex.Lock()
	ws.orderHandlers["order"] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s", ws.listenKey),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	// 启动listenKey续期协程
	go ws.keepAliveListenKey()

	return ws.ws.SendMessage(msgBytes)
}

// SubscribeAccount 订阅账户更新
func (ws *WebSocketBase) SubscribeAccount(handler func(map[string]model.Account), opts ...model.OptionParameter) error {
	// 账户更新和订单更新使用同一个listenKey，所以这里直接复用SubscribeOrder的逻辑
	return errors.New("use SubscribeFuturesAccount instead")
}

// SubscribePosition 订阅持仓更新
func (ws *WebSocketBase) SubscribePosition(handler func([]model.FuturesPosition), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	if ws.apiKey == "" || ws.apiSecret == "" {
		return errors.New("API key and secret are required for private WebSocket")
	}

	// 获取listenKey
	err := ws.getListenKey()
	if err != nil {
		return err
	}

	// 设置处理器
	ws.mutex.Lock()
	ws.positionHandlers["position"] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s", ws.listenKey),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	// 启动listenKey续期协程
	go ws.keepAliveListenKey()

	return ws.ws.SendMessage(msgBytes)
}

// SubscribeFuturesAccount 订阅期货账户更新
func (ws *WebSocketBase) SubscribeFuturesAccount(handler func(map[string]model.FuturesAccount), opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	if ws.apiKey == "" || ws.apiSecret == "" {
		return errors.New("API key and secret are required for private WebSocket")
	}

	// 获取listenKey
	err := ws.getListenKey()
	if err != nil {
		return err
	}

	// 设置处理器
	ws.mutex.Lock()
	ws.accountHandlers["account"] = handler
	ws.mutex.Unlock()

	// 构造订阅消息
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s", ws.listenKey),
		},
		"id": util.GenerateOrderClientId(32),
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(subMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	// 启动listenKey续期协程
	go ws.keepAliveListenKey()

	return ws.ws.SendMessage(msgBytes)
}
