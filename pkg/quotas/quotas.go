package quotas

import (
	"context"
	"fmt"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/servicequotas"

	"github.com/giantswarm/aws-servicequotas-operator/pkg/aws/scope"
	"github.com/giantswarm/aws-servicequotas-operator/pkg/aws/services/quotas"
	ctrlmetrics "github.com/giantswarm/aws-servicequotas-operator/pkg/metrics"
)

type QuotasService struct {
	Client client.Client
	Scope  *scope.ClusterScope
	DryRun bool

	Quotas *quotas.Service
}

type QuotaCodeValue struct {
	Description string

	Code  *string
	Value *float64
}

func New(scope *scope.ClusterScope, client client.Client, dryRun bool) *QuotasService {
	return &QuotasService{
		Scope:  scope,
		Client: client,
		DryRun: dryRun,

		Quotas: quotas.NewService(scope),
	}
}

func (s *QuotasService) Reconcile(ctx context.Context) {
	s.Scope.Info("Reconciling AWSCluster CR for service quotas ")

	// Default quotas we want to set only for workload cluster accounts
	quotas := map[string][]QuotaCodeValue{
		"s3": {
			{
				Description: "Number of buckets",

				Code:  aws.String("L-DC2B2D3D"),
				Value: aws.Float64(1000),
			},
		},
		"vpc": {
			// TODO: Apply in Management Account
			//{
			//Description: "Routes per route table",

			//Code:  aws.String("L-93826ACB"),
			//Value: aws.Float64(200),
			//},
			{
				Description: "VPCs per Region",

				Code:  aws.String("L-F678F1CE"),
				Value: aws.Float64(50),
			},
			{
				Description: "NAT gateways per Availability Zone",

				Code:  aws.String("L-FE5A380F"),
				Value: aws.Float64(50),
			},
			{
				Description: "IPv4 CIDR blocks per VPC",

				Code:  aws.String("L-83CA0A9D"),
				Value: aws.Float64(50),
			},
		},
		"ec2": {
			{
				Description: "EC2-VPC Elastic IPs",

				Code:  aws.String("L-0263D0A3"),
				Value: aws.Float64(50),
			},
			{
				Description: "Running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances",

				Code:  aws.String("L-1216C47A"),
				Value: aws.Float64(250),
			},
		},
		"elasticloadbalancing": {
			{
				Description: "Application Load Balancers per Region",

				Code:  aws.String("L-53DA6B97"),
				Value: aws.Float64(100),
			},
			{
				Description: "Classic Load Balancers per Region",

				Code:  aws.String("L-E9E9831D"),
				Value: aws.Float64(100),
			},
		},
		"autoscaling": {
			{
				Description: "Auto Scaling groups per region",

				Code:  aws.String("L-CDE20ADC"),
				Value: aws.Float64(250),
			},
			{
				Description: "Launch configurations per region",

				Code:  aws.String("L-6B80B8FA"),
				Value: aws.Float64(500),
			},
		},
	}

	for serviceCode, quotasPerService := range quotas {
		for _, quotaCodeValue := range quotasPerService {
			var (
				err           error
				historyOutput *servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput
				appliedOutput *servicequotas.GetServiceQuotaOutput
				increaseQuota bool
			)
			// Get the current quota value, sometimes it is not available, e.g. for S3 buckets
			appliedInput := &servicequotas.GetServiceQuotaInput{
				ServiceCode: &serviceCode,
				QuotaCode:   quotaCodeValue.Code,
			}
			appliedOutput, err = s.Quotas.Client.GetServiceQuota(appliedInput)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					switch awsErr.Code() {
					case servicequotas.ErrCodeNoSuchResourceException:
						// fall through
					default:
						ctrlmetrics.QuotaAppliedErrors.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code, strconv.Itoa(int(*quotaCodeValue.Value))).Inc()
						s.Scope.Error(err, "Failed to get applied service quota")
						continue
					}
				} else {
					ctrlmetrics.QuotaAppliedErrors.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code, strconv.Itoa(int(*quotaCodeValue.Value))).Inc()
					s.Scope.Error(err, "Failed to get applied service quota")
					continue
				}
			}
			if appliedOutput.Quota != nil {
				ctrlmetrics.QuotaAppliedValues.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code).Set(*appliedOutput.Quota.Value)
				if *quotaCodeValue.Value > *appliedOutput.Quota.Value {
					increaseQuota = true
				} else {
					continue
				}
			}

			// Check if the quota recently has been changed, this is helpful when we don't get the applied quota, e.g. S3 buckets
			historyInput := &servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaInput{
				ServiceCode: &serviceCode,
				QuotaCode:   quotaCodeValue.Code,
			}
			for {
				historyOutput, err = s.Quotas.Client.ListRequestedServiceQuotaChangeHistoryByQuota(historyInput)
				if err != nil {
					ctrlmetrics.QuotaHistoryErrors.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code, strconv.Itoa(int(*quotaCodeValue.Value))).Inc()
					s.Scope.Error(err, "Failed to list requested service quota change history by quota")
					break
				}
				if historyOutput.NextToken == nil {
					break
				}
				historyOutput.NextToken = historyInput.NextToken
			}

			if historyOutput != nil {
				var count int
				for _, r := range historyOutput.RequestedQuotas {
					ctrlmetrics.QuotaAppliedValues.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code).Set(*r.DesiredValue)
					if (*quotaCodeValue.Value > *r.DesiredValue) &&
						(*r.QuotaCode == *quotaCodeValue.Code) &&
						(*r.ServiceCode == serviceCode) {
						count++
					}
				}
				if count == len(historyOutput.RequestedQuotas) {
					increaseQuota = true
				}
			}

			if increaseQuota {
				if !s.DryRun {
					if s.Scope.AccountId() != s.Scope.ManagementAccountId() {
						s.Scope.Info(fmt.Sprintf("Setting quota for Service %s: Code %s Desired Value: %v", quotaCodeValue.Description, *quotaCodeValue.Code, *quotaCodeValue.Value))
						increaseRequests := []*servicequotas.RequestServiceQuotaIncreaseInput{
							{
								DesiredValue: quotaCodeValue.Value,
								QuotaCode:    quotaCodeValue.Code,
								ServiceCode:  &serviceCode,
							},
						}
						for _, r := range increaseRequests {
							_, err = s.Quotas.Client.RequestServiceQuotaIncrease(r)
							if err != nil {
								if awsErr, ok := err.(awserr.Error); ok {
									switch awsErr.Code() {
									case servicequotas.ErrCodeResourceAlreadyExistsException:
										s.Scope.Info("Service quota already requested, skipping")
										continue
									case servicequotas.ErrCodeIllegalArgumentException:
										s.Scope.Info("Current service quota value is already greater, skipping")
										continue
									default:
										ctrlmetrics.QuotaIncreaseErrors.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code, strconv.Itoa(int(*quotaCodeValue.Value))).Inc()
										s.Scope.Error(err, "Failed to request service quota increase")
										continue
									}
								} else {
									ctrlmetrics.QuotaIncreaseErrors.WithLabelValues(s.Scope.AccountId(), s.Scope.ClusterName(), s.Scope.ClusterNamespace(), serviceCode, quotaCodeValue.Description, *quotaCodeValue.Code, strconv.Itoa(int(*quotaCodeValue.Value))).Inc()
									s.Scope.Error(err, "Failed to request service quota increase")
									continue
								}
							}
							s.Scope.Info(fmt.Sprintf("Quota successfully requested for Service %s: Code %s, Desired Value: %v", quotaCodeValue.Description, *quotaCodeValue.Code, *quotaCodeValue.Value), s.Scope.ClusterNamespace(), s.Scope.ClusterName())
						}
					} else {
						s.Scope.Info("Ignoring quota requests, workload account matches with management account")
					}
				} else {
					s.Scope.Info(fmt.Sprintf("Would set quota for Service %s: Code %s, Desired Value: %v ", quotaCodeValue.Description, *quotaCodeValue.Code, *quotaCodeValue.Value))
				}
			}
		}
	}
}
