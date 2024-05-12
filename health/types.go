package health

import "net/http"

type CheckFn func() error

type Checker interface {
	ChecksAdder
	LivenessHandler(writer http.ResponseWriter, request *http.Request)
	ReadinessHandler(writer http.ResponseWriter, request *http.Request)
}

type ChecksAdder interface {
	AddLivenessCheck(check CheckFn)
	AddReadinessCheck(check CheckFn)
}
