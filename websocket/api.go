package websocket

import (
	"github.com/nntaoli-project/goex/v2/model"
)

// IWebSocketAPI 定义WebSocket API的通用接口
type IWebSocketAPI interface {
	// GetName 获取交易所名称
	// 返回值:
	//   - string: 交易所名称
	GetName() string

	// Connect 连接到WebSocket服务器
	// 返回值:
	//   - error: 连接错误信息
	Connect() error

	// Close 关闭WebSocket连接
	// 返回值:
	//   - error: 关闭连接时的错误信息
	Close() error

	// IsConnected 检查是否已连接
	// 返回值:
	//   - bool: 是否已连接
	IsConnected() bool
}

// IPubWebSocket 定义公共WebSocket API接口
type IPubWebSocket interface {
	IWebSocketAPI

	// SubscribeDepth 订阅深度数据
	// 参数:
	//   - pair: 交易对
	//   - size: 深度大小
	//   - handler: 深度数据处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeDepth(pair model.CurrencyPair, size int, handler func(*model.Depth), opts ...model.OptionParameter) error

	// SubscribeTicker 订阅行情数据
	// 参数:
	//   - pair: 交易对
	//   - handler: 行情数据处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeTicker(pair model.CurrencyPair, handler func(*model.Ticker), opts ...model.OptionParameter) error

	// SubscribeKline 订阅K线数据
	// 参数:
	//   - pair: 交易对
	//   - period: K线周期
	//   - handler: K线数据处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeKline(pair model.CurrencyPair, period model.KlinePeriod, handler func([]model.Kline), opts ...model.OptionParameter) error

	// SubscribeTrade 订阅交易数据
	// 参数:
	//   - pair: 交易对
	//   - handler: 交易数据处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeTrade(pair model.CurrencyPair, handler func([]model.Trade), opts ...model.OptionParameter) error

	// UnsubscribeDepth 取消订阅深度数据
	// 参数:
	//   - pair: 交易对
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeDepth(pair model.CurrencyPair, opts ...model.OptionParameter) error

	// UnsubscribeTicker 取消订阅行情数据
	// 参数:
	//   - pair: 交易对
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeTicker(pair model.CurrencyPair, opts ...model.OptionParameter) error

	// UnsubscribeKline 取消订阅K线数据
	// 参数:
	//   - pair: 交易对
	//   - period: K线周期
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeKline(pair model.CurrencyPair, period model.KlinePeriod, opts ...model.OptionParameter) error

	// UnsubscribeTrade 取消订阅交易数据
	// 参数:
	//   - pair: 交易对
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeTrade(pair model.CurrencyPair, opts ...model.OptionParameter) error
}

// IPrvWebSocket 定义私有WebSocket API接口
type IPrvWebSocket interface {
	IWebSocketAPI

	// SubscribeOrder 订阅订单更新
	// 参数:
	//   - handler: 订单更新处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeOrder(handler func(*model.Order), opts ...model.OptionParameter) error

	// SubscribeAccount 订阅账户更新
	// 参数:
	//   - handler: 账户更新处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeAccount(handler func(map[string]model.Account), opts ...model.OptionParameter) error

	// UnsubscribeOrder 取消订阅订单更新
	// 参数:
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeOrder(opts ...model.OptionParameter) error

	// UnsubscribeAccount 取消订阅账户更新
	// 参数:
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeAccount(opts ...model.OptionParameter) error
}

// IFuturesPubWebSocket 定义期货公共WebSocket API接口
type IFuturesPubWebSocket interface {
	IPubWebSocket

	// SubscribeFundingRate 订阅资金费率
	// 参数:
	//   - pair: 交易对
	//   - handler: 资金费率处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeFundingRate(pair model.CurrencyPair, handler func(*model.FundingRate), opts ...model.OptionParameter) error

	// UnsubscribeFundingRate 取消订阅资金费率
	// 参数:
	//   - pair: 交易对
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeFundingRate(pair model.CurrencyPair, opts ...model.OptionParameter) error
}

// IFuturesPrvWebSocket 定义期货私有WebSocket API接口
type IFuturesPrvWebSocket interface {
	IPrvWebSocket

	// SubscribePosition 订阅持仓更新
	// 参数:
	//   - handler: 持仓更新处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribePosition(handler func([]model.FuturesPosition), opts ...model.OptionParameter) error

	// SubscribeFuturesAccount 订阅期货账户更新
	// 参数:
	//   - handler: 期货账户更新处理函数
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	SubscribeFuturesAccount(handler func(map[string]model.FuturesAccount), opts ...model.OptionParameter) error

	// UnsubscribePosition 取消订阅持仓更新
	// 参数:
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribePosition(opts ...model.OptionParameter) error

	// UnsubscribeFuturesAccount 取消订阅期货账户更新
	// 参数:
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	UnsubscribeFuturesAccount(opts ...model.OptionParameter) error
}
