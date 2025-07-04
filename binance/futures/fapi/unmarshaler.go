package fapi

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/nntaoli-project/goex/v2/binance/common"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/spf13/cast"
	"time"
)

// UnmarshalGetExchangeInfoResponse 解析交易所信息响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - map[string]model.CurrencyPair: 交易对信息映射，键为交易对标识
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货交易所信息，提取交易对详情、精度和交易限制
func UnmarshalGetExchangeInfoResponse(data []byte) (map[string]model.CurrencyPair, error) {
	var (
		err             error
		currencyPairMap = make(map[string]model.CurrencyPair, 40)
	)

	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var (
			currencyPair model.CurrencyPair
		)

		currencyPair.ContractVal = 1
		currencyPair.ContractValCurrency = model.USDT

		err = jsonparser.ObjectEach(value, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
			valStr := string(val)
			switch string(key) {
			case "symbol":
				currencyPair.Symbol = valStr
			case "baseAsset":
				currencyPair.BaseSymbol = valStr
			case "quoteAsset":
				currencyPair.QuoteSymbol = valStr
			case "contractType":
				currencyPair.ContractAlias = valStr
			case "pricePrecision":
				currencyPair.PricePrecision = cast.ToInt(valStr)
			case "quantityPrecision":
				currencyPair.QtyPrecision = cast.ToInt(valStr)
			case "deliveryDate":
				currencyPair.ContractDeliveryDate = cast.ToInt64(valStr)
			case "onboardDate":

			case "filters":
				_, err = jsonparser.ArrayEach(val, func(filterData []byte, dataType jsonparser.ValueType, offset int, err error) {
					filterType, _ := jsonparser.GetString(filterData, "filterType")
					if filterType == "LOT_SIZE" {
						var (
							minQty []byte
							maxQty []byte
						)

						minQty, _, _, err = jsonparser.Get(filterData, "minQty")
						maxQty, _, _, err = jsonparser.Get(filterData, "maxQty")

						currencyPair.MinQty = cast.ToFloat64(string(minQty))
						currencyPair.MaxQty = cast.ToFloat64(string(maxQty))
					}

					if filterType == "MARKET_LOT_SIZE" {

					}

					//if filterType == "MIN_NOTIONAL" {
					//	currencyPair.MinQty, err = jsonparser.GetFloat(filterData, "notional")
					//	if err != nil {
					//		logger.Errorf("[UnmarshalGetExchangeInfoResponse] get notional error: %s", err.Error())
					//	}
					//}
				})
			}
			return err
		})

		k := fmt.Sprintf("%s%s%s", currencyPair.BaseSymbol, currencyPair.QuoteSymbol, currencyPair.ContractAlias)
		currencyPairMap[k] = currencyPair

	}, "symbols")

	return currencyPairMap, err
}

// UnmarshalDepthResponse 解析深度数据响应
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - *model.Depth: 深度数据，包含买单和卖单
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货深度数据，提取时间戳、买单和卖单信息
func UnmarshalDepthResponse(data []byte) (*model.Depth, error) {
	var (
		dep model.Depth
		err error
	)

	err = jsonparser.ObjectEach(data,
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			switch string(key) {
			case "E":
				dep.UTime = time.UnixMilli(cast.ToInt64(string(value)))
			case "asks":
				items, _ := unmarshalDepthItem(value)
				dep.Asks = items
			case "bids":
				items, _ := unmarshalDepthItem(value)
				dep.Bids = items
			}
			return nil
		})

	return &dep, err
}

// unmarshalDepthItem 解析深度数据项
// 参数:
//   - data: 深度数据项的原始数据
//
// 返回值:
//   - model.DepthItems: 深度数据项列表
//   - error: 错误信息
//
// 注意:
//   - 此函数解析深度数据中的单个项，提取价格和数量信息
func unmarshalDepthItem(data []byte) (model.DepthItems, error) {
	var items model.DepthItems
	_, err := jsonparser.ArrayEach(data, func(asksItemData []byte, dataType jsonparser.ValueType, offset int, err error) {
		item := model.DepthItem{}
		i := 0
		_, err = jsonparser.ArrayEach(asksItemData, func(itemVal []byte, dataType jsonparser.ValueType, offset int, err error) {
			valStr := string(itemVal)
			switch i {
			case 0:
				item.Price = cast.ToFloat64(valStr)
			case 1:
				item.Amount = cast.ToFloat64(valStr)
			}
			i += 1
		})
		items = append(items, item)
	})
	return items, err
}

// UnmarshalKlinesResponse 解析K线数据响应
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - []model.Kline: K线数据列表
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货K线数据，提取时间戳、开盘价、最高价、最低价、收盘价和成交量
func UnmarshalKlinesResponse(data []byte) ([]model.Kline, error) {
	var (
		err    error
		klines []model.Kline
	)

	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var (
			i = 0
			k model.Kline
		)

		_, err = jsonparser.ArrayEach(value, func(val []byte, dataType jsonparser.ValueType, offset int, err error) {
			switch i {
			case 0:
				k.Timestamp, _ = jsonparser.ParseInt(val)
			case 1:
				k.Open = cast.ToFloat64(string(val))
			case 2:
				k.High = cast.ToFloat64(string(val))
			case 3:
				k.Low = cast.ToFloat64(string(val))
			case 4:
				k.Close = cast.ToFloat64(string(val))
			case 5:
				k.Vol = cast.ToFloat64(string(val))
			}
			i += 1
		})

		klines = append(klines, k)
	})

	return klines, err
}

// UnmarshalGetAccountResponse 解析账户信息响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - map[string]model.Account: 账户信息映射，键为币种名称
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货账户信息，提取币种、总余额和可用余额
func UnmarshalGetAccountResponse(data []byte) (map[string]model.Account, error) {
	var accounts = make(map[string]model.Account, 4)
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var acc model.Account
		err = jsonparser.ObjectEach(value, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
			valStr := string(val)
			switch string(key) {
			case "asset":
				acc.Coin = valStr
			case "balance":
				acc.Balance = cast.ToFloat64(valStr)
			case "availableBalance":
				acc.AvailableBalance = cast.ToFloat64(valStr)
			}
			return nil
		})
		accounts[acc.Coin] = acc
	})
	return accounts, err
}

// UnmarshalCreateOrderResponse 解析创建订单响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - *model.Order: 订单信息
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货创建订单响应，提取订单ID、客户端订单ID、已成交数量和平均成交价格
//   - 默认订单状态设置为待处理(Pending)
func UnmarshalCreateOrderResponse(data []byte) (*model.Order, error) {
	var order model.Order
	order.Status = model.OrderStatus_Pending
	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		valStr := string(value)
		switch string(key) {
		case "clientOrderId":
			order.CId = valStr
		case "orderId":
			order.Id = valStr
		case "executedQty":
			order.ExecutedQty = cast.ToFloat64(valStr)
		case "avgPrice":
			order.PriceAvg = cast.ToFloat64(valStr)
		}
		return nil
	})
	return &order, err
}

// UnmarshalGetPendingOrdersResponse 解析未完成订单列表响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - []model.Order: 未完成订单列表
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货未完成订单列表，调用UnmarshalOrderResponse处理每个订单
func UnmarshalGetPendingOrdersResponse(data []byte) ([]model.Order, error) {
	var orders []model.Order
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var ord model.Order
		ord, err = UnmarshalOrderResponse(value)
		orders = append(orders, ord)
	})
	return orders, err
}

// UnmarshalGetOrderInfoResponse 解析订单信息响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - *model.Order: 订单信息
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货单个订单信息，调用UnmarshalOrderResponse处理订单数据
func UnmarshalGetOrderInfoResponse(data []byte) (*model.Order, error) {
	ord, err := UnmarshalOrderResponse(data)
	return &ord, err
}

// UnmarshalGetHistoryOrdersResponse 解析历史订单列表响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - []model.Order: 历史订单列表
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货历史订单列表，调用UnmarshalOrderResponse处理每个订单
func UnmarshalGetHistoryOrdersResponse(data []byte) ([]model.Order, error) {
	var (
		orders []model.Order
		err    error
	)

	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var ord model.Order
		ord, err = UnmarshalOrderResponse(value)
		if err != nil {
			return
		}
		orders = append(orders, ord)
	})

	return orders, err
}

// UnmarshalOrderResponse 解析订单响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - model.Order: 订单信息
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货订单数据，提取订单ID、价格、数量、状态、时间等信息
//   - 根据side和positionSide确定订单方向(开多/开空/平多/平空)
func UnmarshalOrderResponse(data []byte) (ord model.Order, err error) {
	var (
		positionSide string
		side         string
	)

	err = jsonparser.ObjectEach(data, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
		valStr := string(val)
		switch string(key) {
		case "orderId":
			ord.Id = valStr
		case "clientOrderId":
			ord.CId = valStr
		case "price":
			ord.Price = cast.ToFloat64(valStr)
		case "origQty":
			ord.Qty = cast.ToFloat64(valStr)
		case "executeQty":
			ord.ExecutedQty = cast.ToFloat64(valStr)
		case "time":
			ord.CreatedAt = cast.ToInt64(valStr)
		case "updateTime":
			ord.FinishedAt = cast.ToInt64(valStr)
		case "status":
			ord.Status = common.AdaptStringToOrderStatus(valStr)
		case "side":
			side = valStr
		case "positionSide":
			positionSide = valStr
		case "type":
			ord.OrderTy = common.AdaptStringToOrderType(valStr)
		}
		return nil
	})

	if ord.Status == model.OrderStatus_Canceled {
		ord.CanceledAt = ord.FinishedAt
	}

	ord.Side = common.AdaptStringToFuturesOrderSide(side, positionSide)

	return
}

// UnmarshalCancelOrderResponse 解析取消订单响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - error: 错误信息，如果为nil则表示取消成功
//
// 注意:
//   - 此函数解析币安期货取消订单响应，检查是否存在错误码
//   - 如果存在错误码，则返回整个响应作为错误信息
func UnmarshalCancelOrderResponse(data []byte) error {
	_, err := jsonparser.GetString(data, "code")
	if err == nil {
		return errors.New(string(data))
	}
	return nil
}

// UnmarshalGetPositionsResponse 解析持仓信息响应数据
// 参数:
//   - data: API响应的原始数据
//
// 返回值:
//   - []model.FuturesPosition: 持仓信息列表
//   - error: 错误信息
//
// 注意:
//   - 此函数解析币安期货持仓信息，提取杠杆倍数、持仓数量、开仓均价、强平价格、未实现盈亏等
//   - 根据positionSide确定持仓方向(多头/空头)
func UnmarshalGetPositionsResponse(data []byte) ([]model.FuturesPosition, error) {
	var positions []model.FuturesPosition
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var (
			pos     model.FuturesPosition
			posSide string
		)

		err = jsonparser.ObjectEach(value, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
			valStr := string(val)
			switch string(key) {
			case "leverage":
				pos.Lever = cast.ToFloat64(valStr)
			case "positionAmt":
				pos.Qty = cast.ToFloat64(valStr)
				pos.AvailQty = pos.Qty
			case "entryPrice":
				pos.AvgPx = cast.ToFloat64(valStr)
			case "liquidationPrice":
				pos.LiqPx = cast.ToFloat64(valStr)
			case "unRealizedProfit":
				pos.Upl = cast.ToFloat64(valStr)
			case "positionSide":
				posSide = valStr
			}
			return nil
		})

		if posSide == "LONG" {
			pos.PosSide = model.Futures_OpenBuy
		}

		if posSide == "SHORT" {
			pos.PosSide = model.Futures_OpenSell
		}

		if posSide == "BOTH" {
			if pos.Qty < 0 {
				pos.PosSide = model.Futures_OpenSell
			} else {
				pos.PosSide = model.Futures_OpenBuy
			}
		}

		positions = append(positions, pos)
	})
	return positions, err
}
