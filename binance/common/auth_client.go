package common

import (
	"github.com/nntaoli-project/goex/v2/httpcli"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/options"
	"net/url"
)

// AuthClient 封装了需要认证的HTTP请求逻辑
type AuthClient struct {
	ApiOpts options.ApiOptions
	UriOpts options.UriOptions
}

// DoAuthRequest 执行需要认证的HTTP请求
// 参数:
//   - method: HTTP方法，如GET、POST、DELETE等
//   - reqUrl: 请求URL
//   - params: 请求参数
//   - header: 请求头
//
// 返回值:
//   - []byte: 响应数据
//   - error: 错误信息
//
// 注意:
//   - 会自动添加API密钥到请求头
//   - 会自动对请求参数进行签名
//   - 所有参数都会附加到URL中，即使是POST请求
func (ac *AuthClient) DoAuthRequest(method, reqUrl string, params *url.Values, header map[string]string) ([]byte, error) {
	if header == nil {
		header = make(map[string]string, 2)
	}
	header["X-MBX-APIKEY"] = ac.ApiOpts.Key
	SignParams(params, ac.ApiOpts.Secret)

	reqUrl += "?" + params.Encode()

	respBody, err := httpcli.Cli.DoRequest(method, reqUrl, "", header)
	logger.Debugf("[DoAuthRequest] response body: %s", string(respBody))
	return respBody, err
}