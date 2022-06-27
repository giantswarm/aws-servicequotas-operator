package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricNamespace = "aws_servicequotas_operator"
	metricSubsystem = "cluster"

	labelCluster          = "cluster_id"
	labelNamespace        = "cluster_namespace"
	labelServiceName      = "service_name"
	labelQuotaDescription = "quota_description"
	labelQuotaCode        = "quota_code"
	labelQuotaValue       = "quota_value"
)

var (
	labels = []string{labelCluster, labelNamespace, labelServiceName, labelQuotaDescription, labelQuotaCode, labelQuotaValue}

	QuotaIncreaseErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "quota_increase_request_errors",
			Help:      "Number of service quota increase request errors",
		},
		labels,
	)
	QuotaAppliedErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "quota_applied_request_errors",
			Help:      "Number of applied quota request errors",
		},
		labels,
	)
	QuotaHistoryErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "quota_history_request_errors",
			Help:      "Number of service quota history request errors",
		},
		labels,
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(QuotaHistoryErrors, QuotaAppliedErrors, QuotaIncreaseErrors)
}
