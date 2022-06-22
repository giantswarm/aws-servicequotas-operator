package scope

import (
	awsclient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/component-base/version"
	"sigs.k8s.io/cluster-api/util/record"

	"github.com/giantswarm/aws-servicequotas-operator/pkg/aws"
)

// AWSClients contains all the aws clients used by the scopes
type AWSClients struct {
	ServiceQuotas *servicequotas.ServiceQuotas
}

// NewServiceQuotasClient creates a new ServiceQuotas API client for a given session
func NewServiceQuotasClient(session aws.Session, arn string, target runtime.Object) *servicequotas.ServiceQuotas {
	Client := servicequotas.New(session.Session(), &awsclient.Config{Credentials: stscreds.NewCredentials(session.Session(), arn)})
	Client.Handlers.Build.PushFrontNamed(getUserAgentHandler())
	Client.Handlers.Complete.PushBack(recordAWSPermissionsIssue(target))

	return Client
}

func getUserAgentHandler() request.NamedHandler {
	return request.NamedHandler{
		Name: "aws-servicequotas-operator/user-agent",
		Fn:   request.MakeAddToUserAgentHandler("awscluster", version.Get().String()),
	}
}

func recordAWSPermissionsIssue(target runtime.Object) func(r *request.Request) {
	return func(r *request.Request) {
		if awsErr, ok := r.Error.(awserr.Error); ok {
			switch awsErr.Code() {
			case "AuthFailure", "UnauthorizedOperation", "NoCredentialProviders":
				record.Warnf(target, awsErr.Code(), "Operation %s failed with a credentials or permission issue", r.Operation.Name)
			}
		}
	}
}
