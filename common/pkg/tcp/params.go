package tcp

type (
	AuthorizeParams struct {
		Username string `json:"username"`
	}

	SubmitParams struct {
		JobID       int64  `json:"job_id"`
		ClientNonce string `json:"client_nonce"`
		Result      string `json:"result"`
	}

	JobParams struct {
		JobID       int64  `json:"job_id"`
		ServerNonce string `json:"server_nonce"`
	}
)

func (p AuthorizeParams) ToMessage(id int64) Message {
	return Message{
		ID:     &id,
		Method: "authorize",
		Params: map[string]any{
			"username": p.Username,
		},
	}
}

func (p SubmitParams) ToMessage(id int64) Message {
	return Message{
		ID:     &id,
		Method: "submit",
		Params: map[string]any{
			"job_id":       p.JobID,
			"client_nonce": p.ClientNonce,
			"result":       p.Result,
		},
	}
}

func (p JobParams) ToMessage() Message {
	return Message{
		ID:     nil,
		Method: "job",
		Params: map[string]any{
			"job_id":       p.JobID,
			"server_nonce": p.ServerNonce,
		},
	}
}
