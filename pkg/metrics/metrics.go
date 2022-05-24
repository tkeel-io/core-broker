package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// metrics label.
	MetricsLabelTenant = "tenant_id"

	// metrics subscribe name.
	MetricsNameSubNum = "subscribe_num"

	// metrics subscribe entity name.
	MetricsNameSubEntitiesNum = "subscribe_entities_num"
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
