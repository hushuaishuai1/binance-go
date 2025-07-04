package fapi

import (
	"errors"
	"fmt"
	"github.com/nntaoli-project/goex/v2/binance/common"
	. "github.com/nntaoli-project/goex/v2/httpcli"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/util"
	"net/http"
	"net/url"
)

// DoNoAuthRequest 执行不需要认证的HTTP请求
// 参数:
//   - httpMethod: HTTP方法，如GET、POST等
//   - reqUrl: 请求URL
//   - params: 请求参数
//
// 返回值:
//   - []byte: 响应数据
//   - []byte: 响应数据的副本
//   - error: 错误信息
//
// 注意:
//   - GET请求会将参数附加到URL中
//   - 其他请求会将参数放在请求体中
func (f *FApi) DoNoAuthRequest(httpMethod, reqUrl string, params *url.Values) ([]byte, []byte, error) {
	reqBody := ""
	if http.MethodGet == httpMethod {
		reqUrl += "?" + params.Encode()
	}

	responseBody, err := Cli.DoRequest(httpMethod, reqUrl, reqBody, nil)
	if err != nil {

	}

	return responseBody, responseBody, err
}

// GetName 获取交易所名称
// 返回值:
//   - string: 返回交易所名称，固定为"binance.com"
func (f *FApi) GetName() string {
	return "binance.com"
}

// GetExchangeInfo 获取交易所支持的所有期货交易对信息
// 返回值:
//   - map[string]model.CurrencyPair: 交易对信息映射，key为交易对的symbol，value为交易对的详细信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 在使用其他API前，建议先调用此方法获取交易对信息
//   - 返回的交易对信息包含价格精度、数量精度、最小交易量等重要信息
func (f *FApi) GetExchangeInfo() (map[string]model.CurrencyPair, []byte, error) {
	data, body, err := f.DoNoAuthRequest(http.MethodGet, f.UriOpts.Endpoint+f.UriOpts.GetExchangeInfoUri, &url.Values{})
	if err != nil {
		logger.Errorf("[GetExchangeInfo] http request error, body: %s", string(body))
		return nil, body, err
	}

	m, err := f.UnmarshalOpts.GetExchangeInfoResponseUnmarshaler(data)
	if err != nil {
		logger.Errorf("[GetExchangeInfo] unmarshaler data error, err: %s", err.Error())
		return nil, body, err
	}

	f.currencyPairM = m

	return m, body, err
}

// NewCurrencyPair 创建新的期货交易对
// 参数:
//   - baseSym: 基础货币符号，如BTC
//   - quoteSym: 计价货币符号，如USDT
//   - opts: 可选参数，可以指定合约类型
//
// 返回值:
//   - model.CurrencyPair: 交易对信息，包含Symbol、BaseCurrency、QuoteCurrency等
//   - error: 错误信息，如果交易对不存在则返回错误
//
// 注意:
//   - 使用此方法前必须先调用GetExchangeInfo方法
//   - 默认创建永续合约(PERPETUAL)交易对
//   - 可以通过opts参数指定合约类型，如"contractAlias":"PERPETUAL"表示永续合约
func (f *FApi) NewCurrencyPair(baseSym, quoteSym string, opts ...model.OptionParameter) (model.CurrencyPair, error) {
	var (
		contractAlias string
		currencyPair  model.CurrencyPair
	)

	if len(opts) == 0 {
		contractAlias = "PERPETUAL"
	} else if opts[0].Key == "contractAlias" {
		contractAlias = opts[0].Value
	}

	currencyPair = f.currencyPairM[baseSym+quoteSym+contractAlias]
	if currencyPair.Symbol == "" {
		return currencyPair, errors.New("not found currency pair")
	}

	return currencyPair, nil
}

// GetDepth 获取期货币对的深度数据
// 参数:
//   - pair: 交易对信息
//   - limit: 返回的深度数量
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - *model.Depth: 深度数据，包含买单(bids)和卖单(asks)
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - bids按价格降序排列
//   - asks按价格升序排列
func (f *FApi) GetDepth(pair model.CurrencyPair, limit int, opt ...model.OptionParameter) (depth *model.Depth, responseBody []byte, err error) {
	params := url.Values{}
	params.Set("symbol", pair.Symbol)
	params.Set("limit", fmt.Sprint(limit))

	util.MergeOptionParams(&params, opt...)

	data, responseBody, err := f.DoNoAuthRequest(http.MethodGet, f.UriOpts.Endpoint+f.UriOpts.DepthUri, &params)
	if err != nil {
		return nil, responseBody, err
	}

	dep, err := f.UnmarshalOpts.DepthUnmarshaler(data)
	dep.Pair = pair

	return dep, responseBody, err
}

// GetTicker 获取期货币对的行情数据
// 参数:
//   - pair: 交易对信息
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - *model.Ticker: 行情数据，包含最新价、买一价、卖一价、24小时最高价、24小时最低价、24小时成交量等信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 此方法尚未实现，调用会抛出panic异常
func (f *FApi) GetTicker(pair model.CurrencyPair, opt ...model.OptionParameter) (ticker *model.Ticker, responseBody []byte, err error) {
	//TODO implement me
	panic("implement me")
}

// GetKline 获取期货K线数据
// 参数:
//   - pair: 交易对信息
//   - period: K线周期，如1min、5min、15min、30min、60min、1h、4h、1day、1week等
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - []model.Kline: K线数据数组，包含开盘价、收盘价、最高价、最低价、成交量等信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 默认返回100条数据，可以通过opt参数修改limit值来获取更多或更少的数据
func (f *FApi) GetKline(pair model.CurrencyPair, period model.KlinePeriod, opt ...model.OptionParameter) (klines []model.Kline, responseBody []byte, err error) {
	var param = url.Values{}
	param.Set("symbol", pair.Symbol)
	param.Set("interval", common.AdaptKlinePeriodToSymbol(period))
	param.Set("limit", "100")

	util.MergeOptionParams(&param, opt...)

	data, responseBody, err := f.DoNoAuthRequest(http.MethodGet, f.UriOpts.Endpoint+f.UriOpts.KlineUri, &param)
	if err != nil {
		return nil, responseBody, err
	}

	klines, err = f.UnmarshalOpts.KlineUnmarshaler(data)

	for i, _ := range klines {
		klines[i].Pair = pair
	}

	return klines, responseBody, err
}
