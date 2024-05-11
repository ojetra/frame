package frame

import "time"

const (
	GracefulShutdownTimeoutSecondDefault = int64(10)
)

type Config struct {
	HTTPServer                    HTTPServer `json:"http_server"`
	GracefulShutdownTimeoutSecond int64      `json:"graceful_shutdown_timeout_second"`
}

type HTTPServer struct {
	Port int `json:"port"`
}

func (x *Config) GetGracefulShutdownTimeoutSecond() time.Duration {
	gracefulShutdownTimeoutSecond := GracefulShutdownTimeoutSecondDefault
	if x.GracefulShutdownTimeoutSecond != 0 {
		gracefulShutdownTimeoutSecond = x.GracefulShutdownTimeoutSecond
	}
	return time.Duration(gracefulShutdownTimeoutSecond) * time.Second
}
