package spot

import (
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"net/url"
)

func adaptKlinePeriod(period model.KlinePeriod) string {
	switch period {
	// 支持所有K线周期常量
	case model.Kline_1min, model.KLINE_PERIOD_1MIN:
		return "1m"
	case model.KLINE_PERIOD_3MIN:
		return "3m"
	case model.Kline_5min, model.KLINE_PERIOD_5MIN:
		return "5m"
	case model.Kline_15min, model.KLINE_PERIOD_15MIN:
		return "15m"
	case model.Kline_30min, model.KLINE_PERIOD_30MIN:
		return "30m"
	case model.Kline_1h: // 与 model.KLINE_PERIOD_60MIN 相同
		return "1h"
	case model.KLINE_PERIOD_2H:
		return "2h"
	case model.Kline_4h: // 与 model.KLINE_PERIOD_4H 相同
		return "4h"
	case model.Kline_6h: // 与 model.KLINE_PERIOD_6H 相同
		return "6h"
	case model.KLINE_PERIOD_8H:
		return "8h"
	case model.KLINE_PERIOD_12H:
		return "12h"
	case model.Kline_1day, model.KLINE_PERIOD_1DAY:
		return "1d"
	case model.KLINE_PERIOD_3DAY:
		return "3d"
	case model.Kline_1week, model.KLINE_PERIOD_1WEEK:
		return "1w"
	case model.KLINE_PERIOD_1MONTH:
		return "1M"
	}
	return string(period)
}

func adaptOrderSide(s model.OrderSide) string {
	switch s {
	case model.Spot_Buy:
		return "BUY"
	case model.Spot_Sell:
		return "SELL"
	default:
		logger.Warnf("[adapt side] order side:%s error", s)
	}
	return string(s)
}

func adaptOrderType(ty model.OrderType) string {
	switch ty {
	case model.OrderType_Limit:
		return "LIMIT"
	case model.OrderType_Market:
		return "MARKET"
	default:
		logger.Warnf("[adapt order type] order typ unknown")
	}
	return string(ty)
}

func adaptOrderStatus(st string) model.OrderStatus {
	switch st {
	case "NEW":
		return model.OrderStatus_Pending
	case "FILLED":
		return model.OrderStatus_Finished
	case "CANCELED":
		return model.OrderStatus_Canceled
	case "PARTIALLY_FILLED":
		return model.OrderStatus_PartFinished
	}
	return model.OrderStatus(-1)
}

func adaptOrderOrigSide(side string) model.OrderSide {
	switch side {
	case "BUY":
		return model.Spot_Buy
	case "SELL":
		return model.Spot_Sell
	default:
		logger.Warnf("[adaptOrderOrigSide] unknown order origin side: %s", side)
	}
	return model.OrderSide(side)
}

func adaptOrderOrigType(ty string) model.OrderType {
	switch ty {
	case "LIMIT":
		return model.OrderType_Limit
	case "MARKET":
		return model.OrderType_Market
	default:
		return model.OrderType(ty)
	}
}

func adaptOrderOrigStatus(st string) model.OrderStatus {
	switch st {
	case "NEW":
		return model.OrderStatus_Pending
	case "FILLED":
		return model.OrderStatus_Finished
	case "CANCELED":
		return model.OrderStatus_Canceled
	case "PARTIALLY_FILLED":
		return model.OrderStatus_PartFinished
	default:
		return model.OrderStatus(-1)
	}
}

func adaptClientOrderId(params *url.Values) {
	cid := params.Get(model.Order_Client_ID__Opt_Key)
	if cid != "" {
		params.Set("origClientOrderId", cid) //clOrdId
		params.Del(model.Order_Client_ID__Opt_Key)
	}
}
