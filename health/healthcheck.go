package health

import "net/http"

type Healthcheck struct {
	livenessChecks  []CheckFn
	readinessChecks []CheckFn
}

func New() *Healthcheck {
	return &Healthcheck{}
}

func (x *Healthcheck) AddLivenessCheck(check CheckFn) {
	x.livenessChecks = append(x.livenessChecks, check)
}

func (x *Healthcheck) AddReadinessCheck(check CheckFn) {
	x.readinessChecks = append(x.readinessChecks, check)
}

func (x *Healthcheck) LivenessHandler(writer http.ResponseWriter, request *http.Request) {
	probeHandler(writer, request, x.livenessChecks)
}

func (x *Healthcheck) ReadinessHandler(writer http.ResponseWriter, request *http.Request) {
	probeHandler(writer, request, x.readinessChecks)
}

func probeHandler(writer http.ResponseWriter, request *http.Request, checks []CheckFn) {
	if request.Method != http.MethodGet {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := http.StatusOK
	for _, check := range checks {
		err := check()
		if err != nil {
			status = http.StatusInternalServerError
		}
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	writer.Header().Set("Pragma", "no-cache")
	writer.Header().Set("Expires", "0")

	writer.WriteHeader(status)
}
