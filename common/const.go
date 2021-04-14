package common

const (
	PresetPath = "/preset/api/v1.10/"
)

type ResponseData struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
