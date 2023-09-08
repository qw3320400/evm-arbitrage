package protocol

import (
	"bytes"
	"fmt"
	"monitor/storage"
	"strconv"
)

var (
	_ storage.DataUpdate = &StateFromLogUpdate{}
	_ DataConvert        = &StateFromLogUpdate{}
)

type DataConvert interface {
	ToFileData() []byte
}

type StateFromLogUpdate struct {
	BlockNumber uint64
	TxIndex     uint
	LogIndex    uint
	Timestamp   int64
}

func (old *StateFromLogUpdate) NeedUpdate(new interface{}) bool {
	var newState *StateFromLogUpdate
	switch v := new.(type) {
	case *StateFromLogUpdate:
		newState = v
	case *UniswapV2Pair:
		newState = v.StateFromLogUpdate
	}
	return newState.BlockNumber > old.BlockNumber ||
		(newState.BlockNumber == old.BlockNumber && newState.TxIndex > old.TxIndex) ||
		(newState.BlockNumber == old.BlockNumber && newState.TxIndex == old.TxIndex && newState.LogIndex > old.LogIndex)
}

func (old *StateFromLogUpdate) Expired(now int64) bool {
	return now > old.Timestamp+86400
}

func (s *StateFromLogUpdate) ToFileData() []byte {
	if s == nil {
		return []byte{}
	}
	return []byte(fmt.Sprintf("%d,%d,%d,%d",
		s.BlockNumber,
		s.TxIndex,
		s.LogIndex,
		s.Timestamp,
	))
}

func (s *StateFromLogUpdate) FromFileData(body []byte) error {
	if s == nil {
		return fmt.Errorf("s is nil")
	}
	words := bytes.Split(body, []byte(","))
	if len(words) != 4 {
		return fmt.Errorf("data format error %s", string(body))
	}
	var err error
	s.BlockNumber, err = strconv.ParseUint(string(words[0]), 10, 64)
	if err != nil {
		return fmt.Errorf("data format error %s", string(body))
	}
	txIndex, err := strconv.ParseUint(string(words[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("data format error %s", string(body))
	}
	s.TxIndex = uint(txIndex)
	logIndex, err := strconv.ParseUint(string(words[2]), 10, 64)
	if err != nil {
		return fmt.Errorf("data format error %s", string(body))
	}
	s.LogIndex = uint(logIndex)
	timestamp, err := strconv.ParseInt(string(words[3]), 10, 64)
	if err != nil {
		return fmt.Errorf("data format error %s", string(body))
	}
	s.Timestamp = timestamp
	return nil
}

func FileDataToStorage(key string, line []byte) (interface{}, interface{}, error) {
	if len(line) == 0 {
		return nil, nil, fmt.Errorf("line is nil")
	}
	switch key {
	case storage.StoreKeyUniswapv2Pairs:
		data := &UniswapV2Pair{}
		err := data.FromFileData(line)
		if err != nil {
			return nil, nil, fmt.Errorf("from file data fail %s", err)
		}
		return data.Address, data, nil
	default:
		return nil, nil, fmt.Errorf("key error %s", key)
	}
}
