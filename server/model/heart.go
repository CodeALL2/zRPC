package model

type HeartRequest struct {
	seq int //请求序列号
}

type HeartResponse struct {
	ack int //返回序列号
}
