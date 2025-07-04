package model

const (
	Kline_1min  KlinePeriod = "1min"
	Kline_5min              = "5min"
	Kline_15min             = "15min"
	Kline_30min             = "30min"
	Kline_60min             = "60min"
	Kline_1h                = "1h"
	Kline_4h                = "4h" // 与 KLINE_PERIOD_4H 相同
	Kline_6h                = "6h" // 与 KLINE_PERIOD_6H 相同
	Kline_1day              = "1day"
	Kline_1week             = "1week"
)

// 币安WebSocket API使用的K线周期常量
const (
	KLINE_PERIOD_1MIN  KlinePeriod = "1m"
	KLINE_PERIOD_3MIN  KlinePeriod = "3m"
	KLINE_PERIOD_5MIN  KlinePeriod = "5m"
	KLINE_PERIOD_15MIN KlinePeriod = "15m"
	KLINE_PERIOD_30MIN KlinePeriod = "30m"
	KLINE_PERIOD_60MIN KlinePeriod = "1h" // 与 KLINE_PERIOD_1H 相同
	// KLINE_PERIOD_1H 已删除，使用 KLINE_PERIOD_60MIN 代替
	KLINE_PERIOD_2H  KlinePeriod = "2h"
	KLINE_PERIOD_4H  KlinePeriod = "4h"
	KLINE_PERIOD_6H  KlinePeriod = "6h"
	KLINE_PERIOD_8H  KlinePeriod = "8h"
	KLINE_PERIOD_12H KlinePeriod = "12h"
	KLINE_PERIOD_1D  KlinePeriod = "1d"
	KLINE_PERIOD_3D  KlinePeriod = "3d"
	KLINE_PERIOD_1W  KlinePeriod = "1w"
	KLINE_PERIOD_1M  KlinePeriod = "1M"
)

// 兼容性常量，用于WebSocket处理
const (
	KLINE_PERIOD_1DAY   = KLINE_PERIOD_1D
	KLINE_PERIOD_3DAY   = KLINE_PERIOD_3D
	KLINE_PERIOD_1WEEK  = KLINE_PERIOD_1W
	KLINE_PERIOD_1MONTH = KLINE_PERIOD_1M
)

const (
	OrderStatus_Pending      OrderStatus = 1
	OrderStatus_Finished                 = 2
	OrderStatus_Canceled                 = 3
	OrderStatus_PartFinished             = 4
	OrderStatus_Canceling                = 5
)

// 订单状态常量，用于WebSocket处理
const (
	ORDER_UNFINISH    = OrderStatus_Pending
	ORDER_PART_FINISH = OrderStatus_PartFinished
	ORDER_FINISH      = OrderStatus_Finished
	ORDER_CANCEL      = OrderStatus_Canceled
	ORDER_REJECT      = OrderStatus_Canceled // 拒绝订单视为取消
)

const (
	Spot_Buy          OrderSide = "buy"
	Spot_Sell         OrderSide = "sell"
	Futures_OpenBuy   OrderSide = "futures_open_buy"
	Futures_OpenSell  OrderSide = "futures_open_sell"
	Futures_CloseBuy  OrderSide = "futures_close_buy"
	Futures_CloseSell OrderSide = "futures_close_sell"
)

// 交易方向常量，用于WebSocket处理
const (
	BUY  = Spot_Buy
	SELL = Spot_Sell
)

const (
	OrderType_Limit    OrderType = "limit"
	OrderType_Market   OrderType = "market"
	OrderType_opponent OrderType = "opponent"
)

// 订单类型常量，用于WebSocket处理
const (
	LIMIT       = OrderType_Limit
	MARKET      = OrderType_Market
	STOP        = "stop"
	STOP_MARKET = "stop_market"
)

// coin const list
// a-z排序
const (
	ADA  = "ADA"
	ATOM = "ATOM"
	AAVE = "AAVE"
	ALGO = "ALGO"
	AR   = "AR"

	BTC  = "BTC"
	BNB  = "BNB"
	BSV  = "BSV"
	BCH  = "BCH"
	BUSD = "BUSD"

	CEL = "CEL"
	CRV = "CRV"

	DAI  = "DAI"
	DCR  = "DCR"
	DOT  = "DOT"
	DOGE = "DOGE"
	DASH = "DASH"
	DYDX = "DYDX"

	ETH  = "ETH"
	ETHW = "ETHW"
	ETC  = "ETC"
	EOS  = "EOS"
	ENJ  = "ENJ"
	ENS  = "ENS"

	FLOW = "FLOW"
	FIL  = "FIL"
	FLM  = "FLM"

	GALA = "GALA"
	GAS  = "GAS"

	HT = "HT"

	IOTA = "IOTA"
	IOST = "IOST"

	KSM = "KSM"

	LTC = "LTC"
	LDO = "LDO"

	MINA = "MINA"
	MEME = "MEME"

	NEO  = "NEO"
	NEAR = "NEAR"

	OP   = "OP"
	OKB  = "OKB"
	OKT  = "OKT"
	ORDI = "ORDI"

	PLG  = "PLG"
	PERP = "PERP"
	PEPE = "PEPE"

	QTUM = "QTUM"

	RACA = "RACA"
	RVN  = "RVN"

	STORJ = "STORJ"
	SOL   = "SOL"
	SHIB  = "SHIB"
	SC    = "SC"
	SAND  = "SAND"
	SUSHI = "SUSHI"
	SUI   = "SUI"

	TRX   = "TRX"
	TRADE = "TRADE"
	TRB   = "TRB"

	USD  = "USD"
	USDT = "USDT"
	USDC = "USDC"
	UNI  = "UNI"

	VELO = "VELO"

	WBTC  = "WBTC"
	WAVES = "WAVES"

	XRP = "XRP"
	XTZ = "XTZ"

	YFI  = "YFI"
	YFII = "YFII"

	ZEC  = "ZEC"
	ZYRO = "ZYRO"
)

// exchange name const list
const (
	OKX     = "okx.com"
	BINANCE = "binance.com"
)

const (
	Order_Client_ID__Opt_Key = "OrderClientID"
)

const (
	TWO_WAY_POSITION_MODE = "TWO_WAY_POSITION_MODE"
	ONE_WAY_POSITION_MODE = "ONE_WAY_POSITION_MODE"
)
