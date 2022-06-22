package key

import (
	"strings"
)

const (
	TagCluster       = "giantswarm.io/cluster"
	TagInstallation  = "giantswarm.io/installation"
	TagOrganization  = "giantswarm.io/organization"
	TagStack         = "giantswarm.io/stack"
	TagCloudProvider = "kubernetes.io/cluster/%s"
)

func ARNPrefix(region string) string {
	arnPrefix := "aws"
	if strings.HasPrefix(region, "cn-") {
		arnPrefix = "aws-cn"
	}
	return arnPrefix
}
