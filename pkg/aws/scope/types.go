package scope

import (
	"github.com/giantswarm/aws-servicequotas-operator/pkg/aws"
)

// ServiceQuotasScope is a scope for use with the ServiceQuotas reconciling service in cluster
type ServiceQuotasScope interface {
	aws.ClusterScoper
}
