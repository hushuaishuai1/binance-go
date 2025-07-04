package model

// Trade 定义交易数据结构
type Trade struct {
	Pair      CurrencyPair `json:"pair"`      // 交易对
	Tid       string       `json:"tid"`       // 交易ID
	Price     float64      `json:"price"`     // 成交价格
	Amount    float64      `json:"amount"`    // 成交数量
	Direction string       `json:"direction"` // 交易方向，buy或sell
	Timestamp int64        `json:"timestamp"` // 成交时间戳
}
