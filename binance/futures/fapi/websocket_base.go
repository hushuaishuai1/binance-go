package fapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/options"
	"github.com/nntaoli-project/goex/v2/websocket"
	"net/http"
	"sync"
	"time"
)

// WebSocketBase 币安期货WebSocket API实现
type WebSocketBase struct {
	name                string
	ws                  websocket.IWebSocketClient
	baseURL             string
	connected           bool
	mutex               sync.RWMutex
	depthHandlers       map[string]func(*model.Depth)
	tickerHandlers      map[string]func(*model.Ticker)
	klineHandlers       map[string]func([]model.Kline)
	tradeHandlers       map[string]func([]model.Trade)
	fundingRateHandlers map[string]func(*model.FundingRate)
	positionHandlers    map[string]func([]model.FuturesPosition)
	accountHandlers     map[string]func(map[string]model.FuturesAccount)
	orderHandlers       map[string]func(*model.Order)
	errorHandler        func(error)
	connectedHandler    func()
	disconnectedHandler func(error)
	currencyPairM       map[string]model.CurrencyPair
	apiKey              string
	apiSecret           string
	listenKey           string
	listenKeyExpireTime time.Time
}

// NewWebSocketBase 创建币安期货WebSocket API
func NewWebSocketBase(apiKey, apiSecret string) *WebSocketBase {
	ws := &WebSocketBase{
		name:                "binance.com",
		baseURL:             "wss://fstream.binance.com/ws",
		depthHandlers:       make(map[string]func(*model.Depth)),
		tickerHandlers:      make(map[string]func(*model.Ticker)),
		klineHandlers:       make(map[string]func([]model.Kline)),
		tradeHandlers:       make(map[string]func([]model.Trade)),
		fundingRateHandlers: make(map[string]func(*model.FundingRate)),
		positionHandlers:    make(map[string]func([]model.FuturesPosition)),
		accountHandlers:     make(map[string]func(map[string]model.FuturesAccount)),
		orderHandlers:       make(map[string]func(*model.Order)),
		currencyPairM:       make(map[string]model.CurrencyPair),
		apiKey:              apiKey,
		apiSecret:           apiSecret,
	}

	// 设置WebSocket客户端
	ws.ws = websocket.WsCli

	// 设置消息处理器
	ws.ws.SetHandler("message", func(msg []byte) {
		ws.handleMessage(msg)
	})

	// 设置错误处理器
	ws.ws.SetErrorHandler(func(err error) {
		logger.Errorf("[Binance Futures] WebSocket error: %v", err)
		if ws.errorHandler != nil {
			ws.errorHandler(err)
		}
	})

	// 设置连接成功处理器
	ws.ws.SetConnectedHandler(func() {
		ws.mutex.Lock()
		ws.connected = true
		ws.mutex.Unlock()

		logger.Info("[Binance Futures] WebSocket connected")
		if ws.connectedHandler != nil {
			ws.connectedHandler()
		}
	})

	// 设置连接断开处理器
	ws.ws.SetDisconnectedHandler(func(err error) {
		ws.mutex.Lock()
		ws.connected = false
		ws.mutex.Unlock()

		logger.Errorf("[Binance Futures] WebSocket disconnected: %v", err)
		if ws.disconnectedHandler != nil {
			ws.disconnectedHandler(err)
		}
	})

	return ws
}

// GetName 获取交易所名称
func (ws *WebSocketBase) GetName() string {
	return ws.name
}

// Connect 连接到WebSocket服务器
func (ws *WebSocketBase) Connect() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if ws.connected {
		return errors.New("already connected")
	}

	// 获取交易对信息
	fapi := NewFApi()
	currencyPairM, _, err := fapi.GetExchangeInfo()
	if err != nil {
		return fmt.Errorf("failed to get exchange info: %w", err)
	}
	ws.currencyPairM = currencyPairM

	// 连接到WebSocket服务器
	return ws.ws.Connect(ws.baseURL)
}

// Close 关闭WebSocket连接
func (ws *WebSocketBase) Close() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !ws.connected {
		return errors.New("not connected")
	}

	return ws.ws.Close()
}

// IsConnected 检查是否已连接
func (ws *WebSocketBase) IsConnected() bool {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	return ws.connected
}

// getListenKey 获取listenKey
func (ws *WebSocketBase) getListenKey() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	// 如果listenKey已存在且未过期，则直接返回
	if ws.listenKey != "" && time.Now().Before(ws.listenKeyExpireTime) {
		return nil
	}

	// 创建一个新的Prv对象来获取listenKey
	fapi := NewFApi()
	prv := fapi.NewPrvApi(options.WithApiKey(ws.apiKey), options.WithApiSecretKey(ws.apiSecret))

	// 获取listenKey
	url := "https://fapi.binance.com/fapi/v1/listenKey"
	headers := map[string]string{
		"X-MBX-APIKEY": ws.apiKey,
	}

	resp, err := prv.DoAuthRequest(http.MethodPost, url, nil, headers)
	if err != nil {
		return fmt.Errorf("failed to get listenKey: %w", err)
	}

	// 解析响应
	var result struct {
		ListenKey string `json:"listenKey"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to unmarshal listenKey response: %w", err)
	}

	ws.listenKey = result.ListenKey
	ws.listenKeyExpireTime = time.Now().Add(30 * time.Minute)

	return nil
}

// keepAliveListenKey 保持listenKey活跃
func (ws *WebSocketBase) keepAliveListenKey() {
	ticker := time.NewTicker(25 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ws.mutex.RLock()
			if !ws.connected {
				ws.mutex.RUnlock()
				return
			}

			// 检查是否还有私有订阅
			hasPrivateSub := len(ws.orderHandlers) > 0 || len(ws.positionHandlers) > 0 || len(ws.accountHandlers) > 0
			ws.mutex.RUnlock()

			if !hasPrivateSub {
				return
			}

			// 续期listenKey
			fapi := NewFApi()
			prv := fapi.NewPrvApi(options.WithApiKey(ws.apiKey), options.WithApiSecretKey(ws.apiSecret))
			url := "https://fapi.binance.com/fapi/v1/listenKey"
			headers := map[string]string{
				"X-MBX-APIKEY": ws.apiKey,
			}

			_, err := prv.DoAuthRequest(http.MethodPut, url, nil, headers)
			if err != nil {
				logger.Errorf("[Binance Futures] Failed to keep alive listenKey: %v", err)
			} else {
				ws.mutex.Lock()
				ws.listenKeyExpireTime = time.Now().Add(30 * time.Minute)
				ws.mutex.Unlock()
				logger.Debug("[Binance Futures] ListenKey renewed successfully")
			}
		}
	}
}
