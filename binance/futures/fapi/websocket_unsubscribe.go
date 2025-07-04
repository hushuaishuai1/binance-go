package fapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/util"
	"strings"
)

// UnsubscribeDepth 取消订阅深度数据
func (ws *WebSocketBase) UnsubscribeDepth(pair model.CurrencyPair, opts ...model.OptionParameter) error {
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
func (ws *WebSocketBase) UnsubscribeTicker(pair model.CurrencyPair, opts ...model.OptionParameter) error {
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
func (ws *WebSocketBase) UnsubscribeKline(pair model.CurrencyPair, period model.KlinePeriod, opts ...model.OptionParameter) error {
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
func (ws *WebSocketBase) UnsubscribeTrade(pair model.CurrencyPair, opts ...model.OptionParameter) error {
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
			fmt.Sprintf("%s@aggTrade", symbol),
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

// UnsubscribeFundingRate 取消订阅资金费率
func (ws *WebSocketBase) UnsubscribeFundingRate(pair model.CurrencyPair, opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 转换交易对格式为小写
	symbol := strings.ToLower(pair.Symbol)

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.fundingRateHandlers, symbol)
	ws.mutex.Unlock()

	// 构造取消订阅消息
	unsubMsg := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{
			fmt.Sprintf("%s@markPrice", symbol),
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

// UnsubscribeOrder 取消订阅订单更新
func (ws *WebSocketBase) UnsubscribeOrder(opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.orderHandlers, "order")
	ws.mutex.Unlock()

	// 如果没有其他私有订阅，则取消订阅listenKey
	if len(ws.orderHandlers) == 0 && len(ws.positionHandlers) == 0 && len(ws.accountHandlers) == 0 {
		// 构造取消订阅消息
		unsubMsg := map[string]interface{}{
			"method": "UNSUBSCRIBE",
			"params": []string{
				fmt.Sprintf("%s", ws.listenKey),
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

	return nil
}

// UnsubscribeAccount 取消订阅账户更新
func (ws *WebSocketBase) UnsubscribeAccount(opts ...model.OptionParameter) error {
	// 账户更新和订单更新使用同一个listenKey，所以这里直接复用UnsubscribeOrder的逻辑
	return errors.New("use UnsubscribeFuturesAccount instead")
}

// UnsubscribePosition 取消订阅持仓更新
func (ws *WebSocketBase) UnsubscribePosition(opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.positionHandlers, "position")
	ws.mutex.Unlock()

	// 如果没有其他私有订阅，则取消订阅listenKey
	if len(ws.orderHandlers) == 0 && len(ws.positionHandlers) == 0 && len(ws.accountHandlers) == 0 {
		// 构造取消订阅消息
		unsubMsg := map[string]interface{}{
			"method": "UNSUBSCRIBE",
			"params": []string{
				fmt.Sprintf("%s", ws.listenKey),
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

	return nil
}

// UnsubscribeFuturesAccount 取消订阅期货账户更新
func (ws *WebSocketBase) UnsubscribeFuturesAccount(opts ...model.OptionParameter) error {
	if !ws.IsConnected() {
		return errors.New("not connected")
	}

	// 移除处理器
	ws.mutex.Lock()
	delete(ws.accountHandlers, "account")
	ws.mutex.Unlock()

	// 如果没有其他私有订阅，则取消订阅listenKey
	if len(ws.orderHandlers) == 0 && len(ws.positionHandlers) == 0 && len(ws.accountHandlers) == 0 {
		// 构造取消订阅消息
		unsubMsg := map[string]interface{}{
			"method": "UNSUBSCRIBE",
			"params": []string{
				fmt.Sprintf("%s", ws.listenKey),
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

	return nil
}
