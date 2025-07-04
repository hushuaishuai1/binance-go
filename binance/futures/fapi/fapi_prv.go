package fapi

import (
	"errors"
	"github.com/nntaoli-project/goex/v2/binance/common"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/options"
	"github.com/nntaoli-project/goex/v2/util"
	"net/http"
	"net/url"
)

// Prv 币安期货私有API实现
// 包含需要身份验证的币安期货交易所API接口实现
type Prv struct {
	*FApi
	*common.AuthClient
}

// GetAccount 获取账户资产信息
// 参数:
//   - currency: 币种，可为空字符串获取所有币种资产
//
// 返回值:
//   - map[string]model.Account: 账户资产信息，键为币种名称
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 使用示例:
//
//	accounts, _, err := prvApi.GetAccount("")
//	if err != nil {
//	  // 处理错误
//	}
//	// 使用accounts
func (p *Prv) GetAccount(currency string) (map[string]model.Account, []byte, error) {
	param := &url.Values{}
	responseBody, err := p.AuthClient.DoAuthRequest(http.MethodGet, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.GetAccountUri, param, nil)
	if err != nil {
		return nil, responseBody, err
	}
	accounts, err := p.UnmarshalOpts.GetAccountResponseUnmarshaler(responseBody)
	return accounts, responseBody, err
}

// CreateOrder 创建订单
// 参数:
//   - pair: 交易对
//   - qty: 数量
//   - price: 价格
//   - side: 订单方向(Futures_OpenBuy/Futures_OpenSell/Futures_CloseBuy/Futures_CloseSell)
//   - orderTy: 订单类型(OrderType_Limit/OrderType_Market等)
//   - opt: 可选参数，如ClientOrderID等
//
// 返回值:
//   - *model.Order: 订单信息
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 注意:
//   - 限价单(OrderType_Limit)的价格*数量必须大于等于5.0 USDT
//   - 默认使用GTC(Good Till Cancel)时效策略
//
// 使用示例:
//
//	order, _, err := prvApi.CreateOrder(
//	  pair, 0.01, 30000, model.Futures_OpenBuy, model.OrderType_Limit)
func (p *Prv) CreateOrder(pair model.CurrencyPair, qty, price float64, side model.OrderSide, orderTy model.OrderType, opt ...model.OptionParameter) (order *model.Order, responseBody []byte, err error) {
	if orderTy == model.OrderType_Limit && qty*price < 5.0 { //币安规则
		return nil, nil, errors.New("MIN NOTIONAL must >= 5.0 USDT")
	}

	var param = url.Values{}
	param.Set("symbol", pair.Symbol)
	param.Set("price", util.FloatToString(price, pair.PricePrecision))
	param.Set("quantity", util.FloatToString(qty, pair.QtyPrecision))
	param.Set("type", common.AdaptOrderTypeToString(orderTy))
	param.Set("side", common.AdaptOrderSideToString(side))
	param.Set("timeInForce", "GTC")
	param.Set("newOrderRespType", "ACK")

	switch side {
	case model.Futures_OpenSell, model.Futures_CloseSell:
		param.Set("positionSide", "SHORT")
	case model.Futures_OpenBuy, model.Futures_CloseBuy:
		param.Set("positionSide", "LONG")
	}

	util.MergeOptionParams(&param, opt...)           //合并参数
	common.AdaptOrderClientIDOptionParameter(&param) //client id

	responseBody, err = p.AuthClient.DoAuthRequest(http.MethodPost, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.NewOrderUri, &param, nil)
	if err != nil {
		return nil, responseBody, err
	}

	ord, err := p.UnmarshalOpts.CreateOrderResponseUnmarshaler(responseBody)
	if ord != nil {
		ord.Pair = pair
		ord.Price = price
		ord.Qty = qty
		ord.Side = side
		ord.OrderTy = orderTy
	}

	return ord, responseBody, err
}

// GetOrderInfo 获取订单信息
// 参数:
//   - pair: 交易对
//   - id: 订单ID
//   - opt: 可选参数
//
// 返回值:
//   - *model.Order: 订单信息
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 使用示例:
//
//	order, _, err := prvApi.GetOrderInfo(pair, "123456")
//	if err != nil {
//	  // 处理错误
//	}
//	// 使用order信息
func (p *Prv) GetOrderInfo(pair model.CurrencyPair, id string, opt ...model.OptionParameter) (order *model.Order, responseBody []byte, err error) {
	param := &url.Values{}
	param.Set("symbol", pair.Symbol)
	param.Set("orderId", id)

	util.MergeOptionParams(param, opt...)

	data, err := p.AuthClient.DoAuthRequest(http.MethodGet, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.GetOrderUri, param, nil)
	if err != nil {
		return nil, data, err
	}

	order, err = p.UnmarshalOpts.GetOrderInfoResponseUnmarshaler(data)
	if err != nil {
		return nil, data, err
	}

	order.Pair = pair

	return
}

// GetPendingOrders 获取当前未完成订单列表
// 参数:
//   - pair: 交易对
//   - opt: 可选参数
//
// 返回值:
//   - []model.Order: 未完成订单列表
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 使用示例:
//
//	orders, _, err := prvApi.GetPendingOrders(pair)
//	if err != nil {
//	  // 处理错误
//	}
//	for _, order := range orders {
//	  // 处理每个订单
//	}
func (p *Prv) GetPendingOrders(pair model.CurrencyPair, opt ...model.OptionParameter) (orders []model.Order, responseBody []byte, err error) {
	param := &url.Values{}
	param.Set("symbol", pair.Symbol)

	util.MergeOptionParams(param, opt...)

	data, err := p.AuthClient.DoAuthRequest(http.MethodGet, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.GetPendingOrdersUri, param, nil)
	if err != nil {
		return nil, data, err
	}

	orders, err = p.UnmarshalOpts.GetPendingOrdersResponseUnmarshaler(data)
	if err != nil {
		return nil, data, err
	}

	for i, _ := range orders {
		orders[i].Pair = pair
	}

	return orders, data, nil
}

// GetHistoryOrders 获取历史订单列表
// 参数:
//   - pair: 交易对
//   - opt: 可选参数，如时间范围等
//
// 返回值:
//   - []model.Order: 历史订单列表
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 注意:
//   - 默认返回最多500条记录
//
// 使用示例:
//
//	orders, _, err := prvApi.GetHistoryOrders(pair)
//	if err != nil {
//	  // 处理错误
//	}
//	for _, order := range orders {
//	  // 处理每个订单
//	}
func (p *Prv) GetHistoryOrders(pair model.CurrencyPair, opt ...model.OptionParameter) (orders []model.Order, responseBody []byte, err error) {
	param := &url.Values{}
	param.Set("symbol", pair.Symbol)
	param.Set("limit", "500")

	util.MergeOptionParams(param, opt...)

	data, err := p.AuthClient.DoAuthRequest(http.MethodGet, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.GetHistoryOrdersUri, param, nil)
	if err != nil {
		return nil, data, err
	}

	orders, err = p.UnmarshalOpts.GetHistoryOrdersResponseUnmarshaler(data)
	if err != nil {
		return nil, data, err
	}

	for i, _ := range orders {
		orders[i].Pair = pair
	}

	return orders, data, nil
}

// CancelOrder 取消订单
// 参数:
//   - pair: 交易对
//   - id: 订单ID
//   - opt: 可选参数
//
// 返回值:
//   - []byte: API响应原始数据
//   - error: 错误信息，如果为nil则表示取消成功
//
// 使用示例:
//
//	_, err := prvApi.CancelOrder(pair, "123456")
//	if err != nil {
//	  // 处理取消失败
//	} else {
//	  // 取消成功
//	}
func (p *Prv) CancelOrder(pair model.CurrencyPair, id string, opt ...model.OptionParameter) (responseBody []byte, err error) {
	param := &url.Values{}
	param.Set("symbol", pair.Symbol)
	param.Set("orderId", id)

	util.MergeOptionParams(param, opt...)

	data, err := p.AuthClient.DoAuthRequest(http.MethodDelete, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.CancelOrderUri, param, nil)
	if err != nil {
		return data, err
	}

	err = p.UnmarshalOpts.CancelOrderResponseUnmarshaler(data)

	return data, err
}

// GetFuturesAccount 获取期货账户信息
// 参数:
//   - currency: 币种
//
// 返回值:
//   - map[string]model.FuturesAccount: 期货账户信息
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 注意:
//   - 此方法尚未实现，调用会导致panic
func (p *Prv) GetFuturesAccount(currency string) (acc map[string]model.FuturesAccount, responseBody []byte, err error) {
	panic("not implement")
}

// GetPositions 获取持仓信息
// 参数:
//   - pair: 交易对
//   - opts: 可选参数
//
// 返回值:
//   - []model.FuturesPosition: 持仓信息列表
//   - []byte: API响应原始数据
//   - error: 错误信息
//
// 使用示例:
//
//	positions, _, err := prvApi.GetPositions(pair)
//	if err != nil {
//	  // 处理错误
//	}
//	for _, position := range positions {
//	  // 处理每个持仓
//	}
func (p *Prv) GetPositions(pair model.CurrencyPair, opts ...model.OptionParameter) (positions []model.FuturesPosition, responseBody []byte, err error) {
	param := &url.Values{}
	param.Set("symbol", pair.Symbol)

	util.MergeOptionParams(param, opts...)

	data, err := p.AuthClient.DoAuthRequest(http.MethodGet, p.AuthClient.UriOpts.Endpoint+p.AuthClient.UriOpts.GetPositionsUri, param, nil)
	if err != nil {
		return nil, data, err
	}

	pos, err := p.UnmarshalOpts.GetPositionsResponseUnmarshaler(data)
	if err != nil {
		return nil, data, err
	}

	for i, _ := range pos {
		pos[i].Pair = pair
	}

	return pos, data, nil
}



// NewPrvApi 创建币安期货私有API实例
// 参数:
//   - fapi: 币安期货API实例
//   - opts: API选项，如API密钥、密钥等
//
// 返回值:
//   - *Prv: 币安期货私有API实例
//
// 使用示例:
//
//	prvApi := NewPrvApi(fapi,
//	  options.WithApiKey("your-api-key"),
//	  options.WithApiSecretKey("your-secret-key"))
func NewPrvApi(fapi *FApi, opts ...options.ApiOption) *Prv {
	var prv = &Prv{
		AuthClient: &common.AuthClient{},
	}
	prv.FApi = fapi
	for _, opt := range opts {
		opt(&prv.AuthClient.ApiOpts)
	}
	return prv
}
