package model

import "errors"

type MsgResult struct {
	Result chan interface{}
}

func (m *MsgResult) GetResult() (interface{}, error) {
	result, ok := <-m.Result
	if !ok {
		return nil, errors.New("返回值错误")
	}
	return result, nil
}
