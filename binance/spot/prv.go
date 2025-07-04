package spot

import (
	"fmt"
	"github.com/nntaoli-project/goex/v2/binance/common"
	. "github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/options"
	. "github.com/nntaoli-project/goex/v2/util"
	"net/http"
	"net/url"
)

// PrvApi 币安现货私有API实现
// 包含需要API密钥认证的交易相关接口
type PrvApi struct {
	*Spot
	*common.AuthClient
}

// NewPrvApi 创建币安现货私有API实例
// 参数:
//   - apiOpts: API选项，如API密钥、密钥等
//
// 返回值:
//   - *PrvApi: 币安现货私有API实例
//
// 使用示例:
//
//	prvApi := NewPrvApi(
//	  options.WithApiKey("your-api-key"),
//	  options.WithApiSecretKey("your-secret-key"))
func NewPrvApi(apiOpts ...options.ApiOption) *PrvApi {
	s := &PrvApi{
		AuthClient: &common.AuthClient{},
	}
	for _, opt := range apiOpts {
		opt(&s.AuthClient.ApiOpts)
	}
	return s
}

// GetAccount 获取账户资产信息
// 参数:
//   - coin: 币种，如果为空则返回所有币种的资产信息
//
// 返回值:
//   - map[string]Account: 账户资产信息，key为币种名称，value为账户信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 默认会过滤掉余额为0的资产
//   - 需要API密钥权限
func (s *PrvApi) GetAccount(coin string) (map[string]Account, []byte, error) {
	params := url.Values{}
	params.Set("omitZeroBalances", "true")
	reqUrl := fmt.Sprintf("%s%s", s.AuthClient.UriOpts.Endpoint, s.AuthClient.UriOpts.GetAccountUri)
	data, err := s.AuthClient.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, data, err
	}
	accounts, err := s.UnmarshalerOpts.GetAccountResponseUnmarshaler(data)
	return accounts, data, err
}

// CreateOrder 创建订单
// 参数:
//   - pair: 交易对信息
//   - qty: 交易数量
//   - price: 交易价格
//   - side: 交易方向，买入或卖出
//   - orderTy: 订单类型，如限价单、市价单等
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - *Order: 订单信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 默认使用GTC(Good Till Cancel)时效策略
//   - 可以通过opt参数传入clientOrderId来指定客户端订单ID
//   - 需要API密钥交易权限
func (s *PrvApi) CreateOrder(pair CurrencyPair, qty, price float64, side OrderSide, orderTy OrderType, opt ...OptionParameter) (*Order, []byte, error) {
	var params = url.Values{}
	params.Set("symbol", pair.Symbol)
	params.Set("side", adaptOrderSide(side))
	params.Set("type", adaptOrderType(orderTy))
	params.Set("timeInForce", "GTC")
	params.Set("quantity", FloatToString(qty, pair.QtyPrecision))
	params.Set("price", FloatToString(price, pair.PricePrecision))
	params.Set("newOrderRespType", "ACK")

	MergeOptionParams(&params, opt...)
	common.AdaptOrderClientIDOptionParameter(&params)

	data, err := s.AuthClient.DoAuthRequest(http.MethodPost,
		fmt.Sprintf("%s%s", s.AuthClient.UriOpts.Endpoint, s.AuthClient.UriOpts.NewOrderUri), &params, nil)
	if err != nil {
		return nil, data, err
	}

	ord, err := s.UnmarshalerOpts.CreateOrderResponseUnmarshaler(data)
	if err != nil {
		return nil, data, err
	}

	ord.Pair = pair
	ord.Price = price
	ord.Qty = qty
	ord.Status = OrderStatus_Pending
	ord.Side = side
	ord.OrderTy = orderTy

	return ord, data, nil
}

// GetOrderInfo 获取订单信息
// 参数:
//   - pair: 交易对信息
//   - id: 订单ID
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - *Order: 订单信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 如果id为空，可以通过opt参数传入clientOrderId来查询订单
//   - 需要API密钥交易权限
func (s *PrvApi) GetOrderInfo(pair CurrencyPair, id string, opt ...OptionParameter) (*Order, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", s.AuthClient.UriOpts.Endpoint, s.AuthClient.UriOpts.GetOrderUri)
	params := url.Values{}
	params.Set("symbol", pair.Symbol)

	if id != "" {
		params.Set("orderId", id)
	}

	MergeOptionParams(&params, opt...)
	adaptClientOrderId(&params)

	resp, err := s.AuthClient.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, resp, err
	}

	ord, err := s.UnmarshalerOpts.GetOrderInfoResponseUnmarshaler(resp)
	if ord != nil {
		ord.Pair = pair
	}

	return ord, resp, err
}

// GetPendingOrders 获取当前未完成的订单
// 参数:
//   - pair: 交易对信息
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - []Order: 订单信息数组
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 返回的是当前未完成（挂单中）的订单列表
//   - 需要API密钥交易权限
func (s *PrvApi) GetPendingOrders(pair CurrencyPair, opt ...OptionParameter) ([]Order, []byte, error) {
	var params = url.Values{}
	params.Set("symbol", pair.Symbol)
	MergeOptionParams(&params, opt...)
	data, err := s.AuthClient.DoAuthRequest(http.MethodGet, fmt.Sprintf("%s%s", s.AuthClient.UriOpts.Endpoint, s.AuthClient.UriOpts.GetPendingOrdersUri), &params, nil)
	if err != nil {
		return nil, data, err
	}
	orders, err := s.UnmarshalerOpts.GetPendingOrdersResponseUnmarshaler(data)
	return orders, data, err
}

// GetHistoryOrders 获取历史订单
// 参数:
//   - pair: 交易对信息
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - []Order: 订单信息数组
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 默认返回最近100条历史订单
//   - 可以通过opt参数修改limit值来获取更多或更少的订单
//   - 需要API密钥交易权限
func (s *PrvApi) GetHistoryOrders(pair CurrencyPair, opt ...OptionParameter) ([]Order, []byte, error) {
	params := url.Values{}
	params.Set("symbol", pair.Symbol)
	params.Set("limit", "100")
	MergeOptionParams(&params, opt...)
	reqUrl := fmt.Sprintf("%s%s", s.AuthClient.UriOpts.Endpoint, s.AuthClient.UriOpts.GetHistoryOrdersUri)
	data, err := s.AuthClient.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, data, err
	}
	orders, err := s.UnmarshalerOpts.GetHistoryOrdersResponseUnmarshaler(data)
	return orders, data, err
}

// CancelOrder 取消订单
// 参数:
//   - pair: 交易对信息
//   - id: 订单ID
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 如果id为空，可以通过opt参数传入clientOrderId来取消订单
//   - 需要API密钥交易权限
func (s *PrvApi) CancelOrder(pair CurrencyPair, id string, opt ...OptionParameter) ([]byte, error) {
	var params = url.Values{}
	params.Set("symbol", pair.Symbol)
	if id != "" {
		params.Set("orderId", id)
	}

	MergeOptionParams(&params, opt...)
	adaptClientOrderId(&params)

	data, err := s.AuthClient.DoAuthRequest(http.MethodDelete, fmt.Sprintf("%s%s", s.AuthClient.UriOpts.Endpoint, s.AuthClient.UriOpts.CancelOrderUri), &params, nil)
	if err != nil {
		return data, err
	}
	return data, s.UnmarshalerOpts.CancelOrderResponseUnmarshaler(data)
}
