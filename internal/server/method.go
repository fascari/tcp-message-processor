package server

const (
	MethodAuthorize Method = "authorize"
	MethodSubmit    Method = "submit"
	MethodJob       Method = "job"
)

type Method string
