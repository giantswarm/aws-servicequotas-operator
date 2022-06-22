package quotas

import (
	"github.com/aws/aws-sdk-go/service/servicequotas/servicequotasiface"

	"github.com/giantswarm/aws-servicequotas-operator/pkg/aws/scope"
)

// Service holds a collection of interfaces.
type Service struct {
	scope  scope.ServiceQuotasScope
	Client servicequotasiface.ServiceQuotasAPI
}

// NewService returns a new service given the Cloudformation api client.
func NewService(clusterScope scope.ServiceQuotasScope) *Service {
	return &Service{
		scope:  clusterScope,
		Client: scope.NewServiceQuotasClient(clusterScope, clusterScope.ARN(), clusterScope.Cluster()),
	}
}
