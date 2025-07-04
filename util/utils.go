package util

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/spf13/cast"
	"io/ioutil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// FloatToString 保留的小数点位数,去除末尾多余的0(StripTrailingZeros)
func FloatToString(v float64, n int) string {
	ret := strconv.FormatFloat(v, 'f', n, 64)
	return strconv.FormatFloat(cast.ToFloat64(ret), 'f', -1, 64) //StripTrailingZeros
}

//// IsoTime
//// Get iso format time
//// eg: 2018-03-16T18:02:48.284Z
//
//func IsoTime() string {
//	utcTime := time.Now().UTC()
//	iso := utcTime.String()
//	isoBytes := []byte(iso)
//	iso = string(isoBytes[:10]) + "T" + string(isoBytes[11:23]) + "Z"
//	return iso
//}

func ValuesToJson(v url.Values) ([]byte, error) {
	paramMap := make(map[string]interface{})
	for k, vv := range v {
		if len(vv) == 1 {
			paramMap[k] = vv[0]
		} else {
			paramMap[k] = vv
		}
	}
	return json.Marshal(paramMap)
}

func GzipUnCompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func FlateUnCompress(data []byte) ([]byte, error) {
	return ioutil.ReadAll(flate.NewReader(bytes.NewReader(data)))
}

func GenerateOrderClientId(size int) string {
	uuidStr := strings.Replace(uuid.New().String(), "-", "", 32)
	return "goex-" + uuidStr[0:size-5]
}

func MergeOptionParams(params *url.Values, opts ...model.OptionParameter) {
	for _, opt := range opts {
		params.Set(opt.Key, opt.Value)
	}
}

// ToFloat64 将各种类型转换为float64
func ToFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("nil value")
	}

	switch v := v.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	default:
		return 0, fmt.Errorf("cannot convert %v(%s) to float64", v, reflect.TypeOf(v))
	}
}

// ToInt64 将各种类型转换为int64
func ToInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, errors.New("nil value")
	}

	switch v := v.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %v(%s) to int64", v, reflect.TypeOf(v))
	}
}

// ToInt 将各种类型转换为int
func ToInt(v interface{}) (int, error) {
	i64, err := ToInt64(v)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}
