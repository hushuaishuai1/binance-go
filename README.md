# binance go API 使用文档

## 项目简介
- 统一并标准化各大主流数字货币交易平台的接口。
- 当前支持交易所：Binance（币安）

## 目录结构
```
├── api.go                  // 核心接口定义（IPubRest、IPrvRest等）
├── binance/                // 币安实现
│   ├── spot/               // 现货相关实现
│   │   ├── spot.go         // 现货API主入口
│   │   ├── prv.go          // 现货私有API（需密钥）
│   │   └── pub.go          // 现货公有API
│   └── futures/fapi/       // USDT永续合约相关实现
│       ├── fapi.go         // 合约API主入口
│       ├── fapi_prv.go     // 合约私有API（需密钥）
│       └── fapi_pub.go     // 合约公有API
├── model/                  // 通用数据结构
├── options/                // 配置与解包选项
├── util/                   // 工具函数
├── websocket/              // WebSocket接口
```

## API 核心接口说明

### 公共接口（无需授权）
接口定义见 `api.go` 的 `IPubRest`：
```go
// 获取深度、Ticker、K线、交易所信息、创建交易对
GetDepth(pair, limit, ...opt) (depth, resp, err)
GetTicker(pair, ...opt) (ticker, resp, err)
GetKline(pair, period, ...opt) (klines, resp, err)
GetExchangeInfo() (map[CurrencyPair], resp, err)
NewCurrencyPair(base, quote, ...opt) (CurrencyPair, err)
```

### 私有接口（需API密钥）
接口定义见 `api.go` 的 `IPrvRest`：
```go
GetAccount(coin) (map[string]Account, resp, err)
CreateOrder(pair, qty, price, side, orderTy, ...opt) (order, resp, err)
GetOrderInfo(pair, id, ...opt) (order, resp, err)
GetPendingOrders(pair, ...opt) (orders, resp, err)
GetHistoryOrders(pair, ...opt) (orders, resp, err)
CancelOrder(pair, id, ...opt) (resp, err)
```

### 合约专用接口
见 `IFuturesPubRest`、`IFuturesPrvRest`，如资金费率、持仓、合约账户等。

## 现货 Spot API 用法

### 创建现货API实例
```go
import (
    goexv2 "github.com/nntaoli-project/goex/v2"
    "github.com/nntaoli-project/goex/v2/model"
    "github.com/nntaoli-project/goex/v2/options"
)

spot := goexv2.Binance.Spot
prvApi := spot.NewPrvApi(
    options.WithApiKey("your-api-key"),
    options.WithApiSecretKey("your-secret-key"))
```

### 主要接口示例
```go
// 获取交易所信息（必须调用）
_, _, err := spot.GetExchangeInfo()

// 获取Ticker
btcUSDT, _ := spot.NewCurrencyPair(model.BTC, model.USDT)
ticker, _, _ := spot.GetTicker(btcUSDT)

// 下单
order, resp, err := prvApi.CreateOrder(btcUSDT, 0.01, 18000, model.Spot_Buy, model.OrderType_Limit)

// 查询订单
ord, resp, err := prvApi.GetOrderInfo(btcUSDT, "orderId")

// 查询未完成订单
orders, resp, err := prvApi.GetPendingOrders(btcUSDT)

// 查询历史订单
orders, resp, err := prvApi.GetHistoryOrders(btcUSDT)

// 撤单
resp, err := prvApi.CancelOrder(btcUSDT, "orderId")
```

## 合约 Futures API 用法

### 创建合约API实例
```go
import (
    goexv2 "github.com/nntaoli-project/goex/v2"
    "github.com/nntaoli-project/goex/v2/model"
    "github.com/nntaoli-project/goex/v2/options"
)

fapi := goexv2.Binance.Swap
prvApi := fapi.NewPrvApi(
    options.WithApiKey("your-api-key"),
    options.WithApiSecretKey("your-secret-key"))
```

### 主要接口示例
```go
// 获取合约账户资产
accounts, resp, err := prvApi.GetAccount("")

// 下单
order, resp, err := prvApi.CreateOrder(pair, 0.01, 30000, model.Futures_OpenBuy, model.OrderType_Limit)

// 查询订单
ord, resp, err := prvApi.GetOrderInfo(pair, "orderId")

// 查询未完成订单
orders, resp, err := prvApi.GetPendingOrders(pair)

// 查询历史订单
orders, resp, err := prvApi.GetHistoryOrders(pair)

// 撤单
resp, err := prvApi.CancelOrder(pair, "orderId")
```

## WebSocket API 示例
```go
import (
    goexv2 "github.com/nntaoli-project/goex/v2"
    "github.com/nntaoli-project/goex/v2/model"
    "log"
    "time"
)

binance := goexv2.NewWithApiKey("your_api_key", "your_api_secret")
_, _, err := binance.Spot.GetExchangeInfo()
btcUSDT, _ := binance.Spot.NewCurrencyPair(model.BTC, model.USDT)
err = binance.SpotWs.Connect()
defer binance.SpotWs.Close()
err = binance.SpotWs.SubscribeDepth(btcUSDT, 20, func(depth *model.Depth) {
    log.Printf("Depth update: %s, Asks: %d, Bids: %d", depth.Pair.Symbol, len(depth.Asks), len(depth.Bids))
})
err = binance.SpotWs.SubscribeKline(btcUSDT, model.KLINE_PERIOD_1MIN, func(klines []model.Kline) {
    for _, k := range klines {
        log.Printf("Kline: %s, Time: %d, Open: %.2f, Close: %.2f", k.Pair.Symbol, k.Timestamp, k.Open, k.Close)
    }
})
err = binance.SpotWs.SubscribeTrade(btcUSDT, func(trades []model.Trade) {
    for _, t := range trades {
        log.Printf("Trade: %s, ID: %s, Price: %.2f, Amount: %.8f", t.Pair.Symbol, t.Tid, t.Price, t.Amount)
    }
})
time.Sleep(5 * time.Minute)
```

## FAQ
- 如何指定订单的ClientID？
```go
ord, resp, err := prvApi.CreateOrder(btcUSDTCurrencyPair, 0.01, 23000,
    model.Spot_Buy, model.OrderType_Limit,
    model.OptionParameter{}.OrderClientID("goex123027892"))
```

## 代码合并与API简化建议
- 现货与合约API结构高度一致，建议统一接口命名与参数风格。
- 公共接口（如GetTicker、GetAccount等）可通过接口继承与组合进一步抽象。
- 统一OptionParameter参数传递方式，减少重复代码。
- 推荐所有API实例均通过WithXXXOption链式配置，便于扩展。

## 致谢
- JetBrains 提供的开源支持
- 社区贡献者

# binance-go
