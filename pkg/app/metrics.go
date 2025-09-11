package app

// Metrics 定义了计数器行为接口
type Metrics interface {
	Increment(key string)
	GetCounter(key string) int64
}
