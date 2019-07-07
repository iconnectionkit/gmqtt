package metric

import "github.com/prometheus/client_golang/prometheus"

type Metric struct {
	PublishReceivedTotal prometheus.Counter
	PublishSentTotal prometheus.Counter
	ConnectedTotal prometheus.Counter // 总链接数， label cleanSession等
	ConnectedCurrent prometheus.Gauge //当前连接数 label cleanSession等
	MessageDroppedTotal prometheus.Counter //丢弃总数量， label qos
	MessageRetryTotal prometheus.Counter //消息重试次数
}