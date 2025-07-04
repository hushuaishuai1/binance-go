package fapi

import (
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/options"
)

// FApi 币安期货API实现
// 包含币安期货交易所的API接口实现
type FApi struct {
	currencyPairM map[string]model.CurrencyPair

	UriOpts       options.UriOptions
	UnmarshalOpts options.UnmarshalerOptions
}

// NewFApi 创建币安期货API实例
// 返回值:
//   - *FApi: 币安期货API实例
//
// 注意:
//   - 此方法会初始化API的URI和反序列化选项
//   - 默认使用币安期货USDT合约的API端点
func NewFApi() *FApi {
	f := &FApi{
		UriOpts: options.UriOptions{
			Endpoint:            "https://fapi.binance.com",
			KlineUri:            "/fapi/v1/klines",
			TickerUri:           "/fapi/v1/ticker/24hr",
			DepthUri:            "/fapi/v1/depth",
			NewOrderUri:         "/fapi/v1/order",
			GetOrderUri:         "/fapi/v1/order",
			GetHistoryOrdersUri: "/fapi/v1/allOrders",
			GetPendingOrdersUri: "/fapi/v1/openOrders",
			CancelOrderUri:      "/fapi/v1/order",
			GetAccountUri:       "/fapi/v2/balance",
			GetPositionsUri:     "/fapi/v2/positionRisk",
			GetExchangeInfoUri:  "/fapi/v1/exchangeInfo",
		},
		UnmarshalOpts: options.UnmarshalerOptions{
			GetExchangeInfoResponseUnmarshaler:  UnmarshalGetExchangeInfoResponse,
			DepthUnmarshaler:                    UnmarshalDepthResponse,
			KlineUnmarshaler:                    UnmarshalKlinesResponse,
			GetAccountResponseUnmarshaler:       UnmarshalGetAccountResponse,
			CreateOrderResponseUnmarshaler:      UnmarshalCreateOrderResponse,
			CancelOrderResponseUnmarshaler:      UnmarshalCancelOrderResponse,
			GetOrderInfoResponseUnmarshaler:     UnmarshalGetOrderInfoResponse,
			GetPendingOrdersResponseUnmarshaler: UnmarshalGetPendingOrdersResponse,
			GetHistoryOrdersResponseUnmarshaler: UnmarshalGetHistoryOrdersResponse,
			GetPositionsResponseUnmarshaler:     UnmarshalGetPositionsResponse,
		},
	}

	return f
}

// WithUriOption 设置URI选项
// 参数:
//   - opts: URI选项函数
//
// 返回值:
//   - *FApi: 当前API实例，用于链式调用
//
// 使用示例:
//
//	api := NewFApi().WithUriOption(
//	  options.WithEndpoint("https://testnet.binancefuture.com"))
func (f *FApi) WithUriOption(opts ...options.UriOption) *FApi {
	for _, opt := range opts {
		opt(&f.UriOpts)
	}
	return f
}

// WithUnmarshalOption 设置反序列化选项
// 参数:
//   - opts: 反序列化选项函数
//
// 返回值:
//   - *FApi: 当前API实例，用于链式调用
//
// 使用示例:
//
//	api := NewFApi().WithUnmarshalOption(
//	  options.WithDepthUnmarshaler(customUnmarshaler))
func (f *FApi) WithUnmarshalOption(opts ...options.UnmarshalerOption) *FApi {
	for _, opt := range opts {
		opt(&f.UnmarshalOpts)
	}
	return f
}

// NewPrvApi 创建币安期货私有API实例
// 参数:
//   - opts: API选项，如API密钥、密钥等
//
// 返回值:
//   - *Prv: 币安期货私有API实例
//
// 使用示例:
//
//	prvApi := fApi.NewPrvApi(
//	  options.WithApiKey("your-api-key"),
//	  options.WithApiSecretKey("your-secret-key"))
func (f *FApi) NewPrvApi(opts ...options.ApiOption) *Prv {
	api := NewPrvApi(f, opts...)
	api.AuthClient.UriOpts = f.UriOpts
	return api
}
