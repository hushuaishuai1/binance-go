package spot

import (
	"errors"
	"fmt"
	. "github.com/nntaoli-project/goex/v2/httpcli"
	"github.com/nntaoli-project/goex/v2/logger"
	. "github.com/nntaoli-project/goex/v2/model"
	. "github.com/nntaoli-project/goex/v2/util"
	"net/http"
	"net/url"
)

// GetName 获取交易所名称
// 返回值:
//   - string: 返回交易所名称，固定为"binance.com"
func (s *Spot) GetName() string {
	return "binance.com"
}

// GetDepth 获取币对的深度数据
// 参数:
//   - pair: 交易对信息
//   - size: 返回的深度数量
//   - opts: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - *Depth: 深度数据，包含买单(bids)和卖单(asks)
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - bids按价格降序排列
//   - asks按价格升序排列
func (s *Spot) GetDepth(pair CurrencyPair, size int, opts ...OptionParameter) (*Depth, []byte, error) {
	params := url.Values{}
	params.Set("symbol", pair.Symbol)
	params.Set("limit", fmt.Sprint(size))
	MergeOptionParams(&params, opts...)

	reqUrl := fmt.Sprintf("%s%s", s.UriOpts.Endpoint, s.UriOpts.DepthUri)
	data, err := s.DoNoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, data, err
	}
	logger.Debugf("[GetDepth] %s", string(data))
	dep, err := s.UnmarshalerOpts.DepthUnmarshaler(data)
	return dep, data, err
}

// GetTicker 获取币对的行情数据
// 参数:
//   - pair: 交易对信息
//   - opt: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - *Ticker: 行情数据，包含最新价、买一价、卖一价、24小时最高价、24小时最低价、24小时成交量等信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 如果在opt中传入"symbols"参数，将会覆盖单个symbol的设置
func (s *Spot) GetTicker(pair CurrencyPair, opt ...OptionParameter) (*Ticker, []byte, error) {
	params := url.Values{}
	params.Set("symbol", pair.Symbol)

	if len(opt) > 0 {
		for _, p := range opt {
			if p.Key == "symbols" {
				params.Del("symbol") //only symbol or symbols
			}
			params.Add(p.Key, p.Value)
		}
	}

	data, err := s.DoNoAuthRequest(http.MethodGet,
		fmt.Sprintf("%s%s", s.UriOpts.Endpoint, s.UriOpts.TickerUri), &params, nil)
	if err != nil {
		return nil, data, fmt.Errorf("%w%s", err, errors.New(string(data)))
	}

	tk, err := s.UnmarshalerOpts.TickerUnmarshaler(data)
	if err != nil {
		return nil, data, err
	}

	tk.Pair = pair

	return tk, data, err
}

// GetKline 获取K线数据
// 参数:
//   - pair: 交易对信息
//   - period: K线周期，如1min、5min、15min、30min、60min、1h、4h、1day、1week等
//   - opts: 可选参数，可以传递额外的请求参数
//
// 返回值:
//   - []Kline: K线数据数组，包含开盘价、收盘价、最高价、最低价、成交量等信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 默认返回1000条数据，可以通过opts参数修改
func (s *Spot) GetKline(pair CurrencyPair, period KlinePeriod, opts ...OptionParameter) ([]Kline, []byte, error) {
	params := url.Values{}
	params.Set("limit", "1000")
	params.Set("symbol", pair.Symbol)
	params.Set("interval", adaptKlinePeriod(period))

	MergeOptionParams(&params, opts...)

	reqUrl := fmt.Sprintf("%s%s", s.UriOpts.Endpoint, s.UriOpts.KlineUri)
	respBody, err := s.DoNoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, respBody, err
	}

	klines, err := s.UnmarshalerOpts.KlineUnmarshaler(respBody)
	return klines, respBody, err
}

// GetExchangeInfo 获取交易所支持的所有交易对信息
// 返回值:
//   - map[string]CurrencyPair: 交易对信息映射，key为交易对的symbol，value为交易对的详细信息
//   - []byte: 原始响应数据
//   - error: 错误信息
//
// 注意:
//   - 在使用其他API前，建议先调用此方法获取交易对信息
//   - 返回的交易对信息包含价格精度、数量精度、最小交易量等重要信息
func (s *Spot) GetExchangeInfo() (map[string]CurrencyPair, []byte, error) {
	body, err := s.DoNoAuthRequest(http.MethodGet, s.UriOpts.Endpoint+s.UriOpts.GetExchangeInfoUri, &url.Values{}, nil)
	if err != nil {
		logger.Errorf("[GetExchangeInfo] http request error, body: %s", string(body))
		return nil, body, err
	}

	m, err := s.UnmarshalerOpts.GetExchangeInfoResponseUnmarshaler(body)
	if err != nil {
		logger.Errorf("[GetExchangeInfo] unmarshaler data error, err: %s", err.Error())
		return nil, body, err
	}

	s.currencyPairM = m

	return m, body, err
}

// NewCurrencyPair 创建新的交易对
// 参数:
//   - baseSym: 基础货币符号，如BTC
//   - quoteSym: 计价货币符号，如USDT
//
// 返回值:
//   - CurrencyPair: 交易对信息，包含Symbol、BaseCurrency、QuoteCurrency等
//   - error: 错误信息，如果交易对不存在则返回错误
//
// 注意:
//   - 使用此方法前必须先调用GetExchangeInfo方法
//   - 返回的CurrencyPair对象包含了交易所对该交易对的所有限制信息
func (s *Spot) NewCurrencyPair(baseSym, quoteSym string) (CurrencyPair, error) {
	currencyPair := s.currencyPairM[baseSym+quoteSym]
	if currencyPair.Symbol == "" {
		return currencyPair, errors.New("not found currency pair")
	}
	return currencyPair, nil
}

// DoNoAuthRequest 执行不需要认证的HTTP请求
// 参数:
//   - method: HTTP方法，如GET、POST等
//   - reqUrl: 请求URL
//   - params: 请求参数
//   - headers: 请求头
//
// 返回值:
//   - []byte: 响应数据
//   - error: 错误信息
//
// 注意:
//   - GET请求会将参数附加到URL中
//   - 其他请求会将参数放在请求体中
func (s *Spot) DoNoAuthRequest(method, reqUrl string, params *url.Values, headers map[string]string) ([]byte, error) {
	var reqBody string

	if method == http.MethodGet {
		reqUrl += "?" + params.Encode()
	} else {
		reqBody = params.Encode()
	}

	responseData, err := Cli.DoRequest(method, reqUrl, reqBody, headers)
	if err != nil {
		return responseData, err
	}

	return responseData, err
}
