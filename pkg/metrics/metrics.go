package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricNamespace = "aws_servicequotas_operator"
	metricSubsystem = "cluster"

	labelCluster     = "cluster_id"
	labelNamespace   = "cluster_namespace"
	labelServiceName = "service_name"
	labelQuotaName   = "quota_name"
	labelQuotaValue  = "quota_value"
)

var (
	labels = []string{labelCluster, labelNamespace, labelServiceName, labelQuotaName, labelQuotaValue}

	QuotaIncreaseErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "quota_increase_errors",
			Help:      "Number of service quota increase errors",
		},
		labels,
	)
	QuotaAppliedErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "quota_applied_errors",
			Help:      "Number of get applied quota errors",
		},
		labels,
	)
	QuotaHistoryErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "quota_history_errors",
			Help:      "Number of service quota history errors",
		},
		labels,
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(QuotaHistoryErrors, QuotaAppliedErrors, QuotaIncreaseErrors)
}
