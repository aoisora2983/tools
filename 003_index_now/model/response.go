package model

type ResponseDetail struct {
	Target  string
	Message string
}
type Response struct {
	Code    string
	Message string
	Details []ResponseDetail
}
