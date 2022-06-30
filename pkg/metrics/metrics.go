package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// metrics label.
	MetricsLabelTenant = "tenant_id"

	// metrics subscribe name.
	MetricsNameSubNum = "subscribe_num"

	// metrics subscribe entity name.
	MetricsNameSubEntitiesNum = "subscribe_entities_num"

	// metrics subscribe name.
	MetricsNameSubMax = "subscribe_max"

	// metrics subscribe entity name.
	MetricsNameSubEntitiesMax = "subscribe_entities_max"
)

var CollectorSubscribeMax = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameSubMax,
		Help: "subscribe max.",
	},
	[]string{MetricsLabelTenant},
)

var CollectorSubscribeEntitiesMax = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameSubEntitiesMax,
		Help: "subscribe entities max.",
	},
	[]string{MetricsLabelTenant},
)

var CollectorSubscribeNum = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameSubNum,
		Help: "subscribe num.",
	},
	[]string{MetricsLabelTenant},
)

var CollectorSubscribeEntitiesNum = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameSubEntitiesNum,
		Help: "subscribe entities num.",
	},
	[]string{MetricsLabelTenant},
)
var Metrics = []prometheus.Collector{CollectorSubscribeEntitiesNum, CollectorSubscribeNum}
