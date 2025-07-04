package fapi

import (
	"github.com/nntaoli-project/goex/v2/model"
)

// reverseAdaptKlinePeriod 将Binance的K线周期转换为模型的KlinePeriod
func reverseAdaptKlinePeriod(interval string) model.KlinePeriod {
	switch interval {
	case "1m":
		return model.Kline_1min
	case "3m":
		return model.KLINE_PERIOD_3MIN
	case "5m":
		return model.Kline_5min
	case "15m":
		return model.Kline_15min
	case "30m":
		return model.Kline_30min
	case "1h":
		return model.Kline_1h
	case "2h":
		return model.KLINE_PERIOD_2H
	case "4h":
		return model.Kline_4h
	case "6h":
		return model.Kline_6h
	case "8h":
		return model.KLINE_PERIOD_8H
	case "12h":
		return model.KLINE_PERIOD_12H
	case "1d":
		return model.Kline_1day
	case "3d":
		return model.KLINE_PERIOD_3DAY
	case "1w":
		return model.Kline_1week
	case "1M":
		return model.KLINE_PERIOD_1MONTH
	default:
		return model.KlinePeriod(interval)
	}
}

// WebSocket 币安期货WebSocket客户端
type WebSocket struct {
	*WebSocketBase
}

// NewWebSocket 创建币安期货WebSocket客户端
func NewWebSocket(apiKey, secretKey string) *WebSocket {
	return &WebSocket{
		WebSocketBase: NewWebSocketBase(apiKey, secretKey),
	}
}
