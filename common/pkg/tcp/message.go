package tcp

import (
	"bufio"
	"encoding/json"
	"net"

	"tcp-message-processor/common/pkg/method"
)

type (
	Message struct {
		ID     *int64         `json:"id"`
		Method string         `json:"method,omitempty"`
		Params map[string]any `json:"params,omitempty"`
		Result any            `json:"result,omitempty"`
		Error  string         `json:"error,omitempty"`
	}
)

func ReadMessage(conn net.Conn) (Message, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}

func WriteMessage(conn net.Conn, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}

func NewSuccessResponse(id int64, result any) *Message {
	return &Message{
		ID:     &id,
		Result: result,
	}
}

func NewErrorResponse(id int64, errMsg string) *Message {
	return &Message{
		ID:     &id,
		Result: false,
		Error:  errMsg,
	}
}

func (m Message) IsJob() bool {
	return m.Method == method.Job.String()
}
