package binance

import (
	"github.com/nntaoli-project/goex/v2/binance/futures/fapi"
	"github.com/nntaoli-project/goex/v2/binance/spot"
)

type Binance struct {
	Spot      *spot.Spot
	Swap      *fapi.FApi
	SpotWs    *spot.WebSocket
	FuturesWs *fapi.WebSocket
}

func New() *Binance {
	return &Binance{
		Spot: spot.New(),
		Swap: fapi.NewFApi(),
	}
}

// NewWithApiKey 使用API密钥创建Binance实例，包括WebSocket支持
func NewWithApiKey(apiKey, secretKey string) *Binance {
	return &Binance{
		Spot:      spot.New(),
		Swap:      fapi.NewFApi(),
		SpotWs:    spot.NewWebSocket(),
		FuturesWs: fapi.NewWebSocket(apiKey, secretKey),
	}
}
