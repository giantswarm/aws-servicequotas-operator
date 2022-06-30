package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricNamespace = "aws_servicequotas_operator"
	metricSubsystem = "quota"

	labelAccountId        = "account_id"
	labelCluster          = "cluster_id"
	labelNamespace        = "cluster_namespace"
	labelServiceName      = "service_name"
	labelQuotaDescription = "quota_description"
	labelQuotaCode        = "quota_code"
	labelQuotaValue       = "quota_value"
)

var (
	errorLabels = []string{labelAccountId, labelCluster, labelNamespace, labelServiceName, labelQuotaDescription, labelQuotaCode, labelQuotaValue}
	infoLabels  = []string{labelAccountId, labelCluster, labelNamespace, labelServiceName, labelQuotaDescription, labelQuotaCode}

	QuotaIncreaseErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "increase_request_errors",
			Help:      "Number of service quota increase request errors",
		},
		errorLabels,
	)
	QuotaAppliedErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "applied_request_errors",
			Help:      "Number of applied quota request errors",
		},
		errorLabels,
	)
	QuotaHistoryErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "history_request_errors",
			Help:      "Number of service quota history request errors",
		},
		errorLabels,
	)

	QuotaAppliedValues = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "applied_values",
			Help:      "Number of applied quota values",
		},
		infoLabels,
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(QuotaAppliedValues, QuotaHistoryErrors, QuotaAppliedErrors, QuotaIncreaseErrors)
}
