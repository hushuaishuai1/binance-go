package goex

import (
	"github.com/nntaoli-project/goex/v2/binance"
	"github.com/nntaoli-project/goex/v2/httpcli"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/websocket"
	"reflect"
)

var (
	DefaultHttpCli = httpcli.Cli
	DefaultWsCli   = websocket.WsCli
)

var (
	Binance = binance.New()
)

// NewWithApiKey 使用API密钥创建Binance实例，包括WebSocket支持
func NewWithApiKey(apiKey, secretKey string) *binance.Binance {
	return binance.NewWithApiKey(apiKey, secretKey)
}

func SetDefaultHttpCli(cli httpcli.IHttpClient) {
	logger.Infof("use new http client implement: %s", reflect.TypeOf(cli).Elem().String())
	httpcli.Cli = cli
}

func SetDefaultWsCli(cli websocket.IWebSocketClient) {
	logger.Infof("use new websocket client implement: %s", reflect.TypeOf(cli).Elem().String())
	websocket.WsCli = cli
}
