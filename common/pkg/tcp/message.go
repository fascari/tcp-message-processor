package tcp

import (
	"tcp-message-processor/common/pkg/method"
)

type Message struct {
	ID     *int64         `json:"id"`
	Method string         `json:"method,omitempty"`
	Params map[string]any `json:"params,omitempty"`
	Result any            `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

func SuccessResponse(id int64, result any) Message {
	return Message{
		ID:     &id,
		Result: result,
	}
}

func ErrorResponse(id int64, errMsg string) Message {
	return Message{
		ID:     &id,
		Result: false,
		Error:  errMsg,
	}
}

func (m Message) IsJob() bool {
	return m.Method == method.Job.String()
}
