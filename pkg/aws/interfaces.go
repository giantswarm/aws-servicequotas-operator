package aws

import (
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"k8s.io/apimachinery/pkg/runtime"
)

// Session represents an AWS session
type Session interface {
	Session() awsclient.ConfigProvider
}

// ClusterScoper is the interface for a workload cluster scope
type ClusterScoper interface {
	Session

	// ARN returns the workload cluster assumed role to operate.
	ARN() string
	// Cluster returns the AWS infrastructure cluster.
	Cluster() runtime.Object
	// Cluster returns the AWS infrastructure cluster name.
	ClusterName() string
	// Cluster returns the AWS infrastructure cluster namespace.
	ClusterNamespace() string
	// Region returns the AWS infrastructure cluster object region.
	Region() string
}
