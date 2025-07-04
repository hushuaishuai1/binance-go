package websocket

import (
	"github.com/nntaoli-project/goex/v2/model"
)

// IWebSocketClient 定义WebSocket客户端接口
type IWebSocketClient interface {
	// Connect 连接到WebSocket服务器
	// 参数:
	//   - url: WebSocket服务器地址
	// 返回值:
	//   - error: 连接错误信息
	Connect(url string) error

	// Close 关闭WebSocket连接
	// 返回值:
	//   - error: 关闭连接时的错误信息
	Close() error

	// Subscribe 订阅特定主题
	// 参数:
	//   - channel: 频道名称
	//   - pairs: 交易对列表
	//   - opts: 可选参数
	// 返回值:
	//   - error: 订阅错误信息
	Subscribe(channel string, pairs []model.CurrencyPair, opts ...model.OptionParameter) error

	// Unsubscribe 取消订阅特定主题
	// 参数:
	//   - channel: 频道名称
	//   - pairs: 交易对列表
	//   - opts: 可选参数
	// 返回值:
	//   - error: 取消订阅错误信息
	Unsubscribe(channel string, pairs []model.CurrencyPair, opts ...model.OptionParameter) error

	// SetHandler 设置消息处理器
	// 参数:
	//   - channel: 频道名称
	//   - handler: 消息处理函数
	SetHandler(channel string, handler func([]byte))

	// SetErrorHandler 设置错误处理器
	// 参数:
	//   - handler: 错误处理函数
	SetErrorHandler(handler func(error))

	// SetConnectedHandler 设置连接成功处理器
	// 参数:
	//   - handler: 连接成功处理函数
	SetConnectedHandler(handler func())

	// SetDisconnectedHandler 设置连接断开处理器
	// 参数:
	//   - handler: 连接断开处理函数
	SetDisconnectedHandler(handler func(error))

	// IsConnected 检查是否已连接
	// 返回值:
	//   - bool: 是否已连接
	IsConnected() bool

	// SendMessage 发送消息
	// 参数:
	//   - message: 要发送的消息
	// 返回值:
	//   - error: 发送错误信息
	SendMessage(message []byte) error
}

// 全局WebSocket客户端
var (
	WsCli IWebSocketClient
)
