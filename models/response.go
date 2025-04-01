package models

type Response struct {
	Success bool   `json:"success" form:"success"`
	Msg     string `json:"msg" form:"msg"`
	Body    interface{}
}