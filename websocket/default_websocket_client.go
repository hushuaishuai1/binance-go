package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"sync"
	"time"
)

// DefaultWebSocketClient 默认的WebSocket客户端实现
type DefaultWebSocketClient struct {
	conn                *websocket.Conn
	url                 string
	mutex               sync.RWMutex
	connected           bool
	handlers            map[string]func([]byte)
	errorHandler        func(error)
	connectedHandler    func()
	disconnectedHandler func(error)
	pingInterval        time.Duration
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxMessageSize      int64
	subscriptions       map[string][]model.CurrencyPair
	subscriptionsMutex  sync.RWMutex
	done                chan struct{}
}

// NewDefaultWebSocketClient 创建一个新的默认WebSocket客户端
func NewDefaultWebSocketClient() *DefaultWebSocketClient {
	return &DefaultWebSocketClient{
		handlers:       make(map[string]func([]byte)),
		pingInterval:   30 * time.Second,
		readTimeout:    60 * time.Second,
		writeTimeout:   10 * time.Second,
		maxMessageSize: 1024 * 1024, // 1MB
		subscriptions:  make(map[string][]model.CurrencyPair),
		done:           make(chan struct{}),
	}
}

// Connect 连接到WebSocket服务器
func (c *DefaultWebSocketClient) Connect(url string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return errors.New("already connected")
	}

	c.url = url

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.connected = true
	c.done = make(chan struct{})

	// 设置连接参数
	c.conn.SetReadLimit(c.maxMessageSize)

	// 启动消息接收协程
	go c.receiveMessages()

	// 启动心跳协程
	go c.keepAlive()

	// 调用连接成功处理器
	if c.connectedHandler != nil {
		c.connectedHandler()
	}

	// 重新订阅之前的频道
	c.resubscribe()

	return nil
}

// 重新订阅之前的频道
func (c *DefaultWebSocketClient) resubscribe() {
	c.subscriptionsMutex.RLock()
	defer c.subscriptionsMutex.RUnlock()

	for channel, pairs := range c.subscriptions {
		err := c.Subscribe(channel, pairs)
		if err != nil {
			logger.Errorf("Failed to resubscribe to channel %s: %v", channel, err)
		}
	}
}

// receiveMessages 接收消息的协程
func (c *DefaultWebSocketClient) receiveMessages() {
	defer func() {
		c.mutex.Lock()
		c.connected = false
		c.mutex.Unlock()

		if c.conn != nil {
			_ = c.conn.Close()
		}

		close(c.done)

		if c.disconnectedHandler != nil {
			c.disconnectedHandler(errors.New("connection closed"))
		}
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			// 设置读取超时
			_ = c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))

			// 读取消息
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if c.errorHandler != nil {
					c.errorHandler(err)
				}
				return
			}

			// 处理消息
			c.handleMessage(message)
		}
	}
}

// handleMessage 处理接收到的消息
func (c *DefaultWebSocketClient) handleMessage(message []byte) {
	// 这里需要根据具体交易所的消息格式解析频道信息
	// 简单实现：遍历所有处理器，让它们自己判断是否处理该消息
	for _, handler := range c.handlers {
		if handler != nil {
			handler(message)
		}
	}
}

// keepAlive 保持连接活跃的协程
func (c *DefaultWebSocketClient) keepAlive() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.RLock()
			if !c.connected {
				c.mutex.RUnlock()
				return
			}

			// 设置写入超时
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

			// 发送ping消息
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			c.mutex.RUnlock()

			if err != nil {
				if c.errorHandler != nil {
					c.errorHandler(err)
				}
				return
			}
		case <-c.done:
			return
		}
	}
}

// Close 关闭WebSocket连接
func (c *DefaultWebSocketClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return errors.New("not connected")
	}

	// 发送关闭消息
	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// 关闭done通道，通知所有协程退出
	close(c.done)

	// 关闭连接
	err := c.conn.Close()
	c.connected = false

	return err
}

// Subscribe 订阅特定主题
func (c *DefaultWebSocketClient) Subscribe(channel string, pairs []model.CurrencyPair, opts ...model.OptionParameter) error {
	c.mutex.RLock()
	if !c.connected {
		c.mutex.RUnlock()
		return errors.New("not connected")
	}
	c.mutex.RUnlock()

	// 保存订阅信息
	c.subscriptionsMutex.Lock()
	c.subscriptions[channel] = pairs
	c.subscriptionsMutex.Unlock()

	// 这里需要根据具体交易所实现订阅逻辑
	// 由子类实现具体的订阅消息格式和发送逻辑
	return errors.New("subscribe method should be implemented by specific exchange")
}

// Unsubscribe 取消订阅特定主题
func (c *DefaultWebSocketClient) Unsubscribe(channel string, pairs []model.CurrencyPair, opts ...model.OptionParameter) error {
	c.mutex.RLock()
	if !c.connected {
		c.mutex.RUnlock()
		return errors.New("not connected")
	}
	c.mutex.RUnlock()

	// 移除订阅信息
	c.subscriptionsMutex.Lock()
	delete(c.subscriptions, channel)
	c.subscriptionsMutex.Unlock()

	// 这里需要根据具体交易所实现取消订阅逻辑
	// 由子类实现具体的取消订阅消息格式和发送逻辑
	return errors.New("unsubscribe method should be implemented by specific exchange")
}

// SetHandler 设置消息处理器
func (c *DefaultWebSocketClient) SetHandler(channel string, handler func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.handlers[channel] = handler
}

// SetErrorHandler 设置错误处理器
func (c *DefaultWebSocketClient) SetErrorHandler(handler func(error)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.errorHandler = handler
}

// SetConnectedHandler 设置连接成功处理器
func (c *DefaultWebSocketClient) SetConnectedHandler(handler func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.connectedHandler = handler
}

// SetDisconnectedHandler 设置连接断开处理器
func (c *DefaultWebSocketClient) SetDisconnectedHandler(handler func(error)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.disconnectedHandler = handler
}

// IsConnected 检查是否已连接
func (c *DefaultWebSocketClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.connected
}

// SendMessage 发送消息
func (c *DefaultWebSocketClient) SendMessage(message []byte) error {
	c.mutex.RLock()
	if !c.connected {
		c.mutex.RUnlock()
		return errors.New("not connected")
	}

	// 设置写入超时
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

	// 发送消息
	err := c.conn.WriteMessage(websocket.TextMessage, message)
	c.mutex.RUnlock()

	return err
}

// 初始化全局WebSocket客户端
func init() {
	WsCli = NewDefaultWebSocketClient()
}
