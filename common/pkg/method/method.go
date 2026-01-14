package method

const (
	Authorize Method = "authorize"
	Submit    Method = "submit"
	Job       Method = "job"
)

type Method string

func (m Method) String() string {
	return string(m)
}
