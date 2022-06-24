//nolint
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

var (
	arn    string
	code   string
	region string
)

var rootCmd = &cobra.Command{
	Use:   "codes",
	Short: "Get service quotas",
	Run: func(cmd *cobra.Command, args []string) {
		getQuotas()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&arn, "arn", "", "AWS assumed role name, e.g. `arn:aws:iam::180547736195:role/GiantSwarmAWSOperator`")
	rootCmd.MarkPersistentFlagRequired("arn")
	rootCmd.PersistentFlags().StringVar(&code, "code", "", "Service quota code, e.g. `s3`")
	rootCmd.MarkPersistentFlagRequired("code")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region, e.g. `us-east-1`")
	rootCmd.MarkPersistentFlagRequired("region")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getQuotas() {
	sess, err := sessionForRegion(region)
	if err != nil {
		fmt.Println(err)
	}
	awsClientConfig := &aws.Config{Credentials: stscreds.NewCredentials(sess, arn)}

	stsClient := sts.New(sess, awsClientConfig)
	_, err = stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Println(err)
	}

	client := servicequotas.New(sess, awsClientConfig)

	inputDefault := &servicequotas.ListAWSDefaultServiceQuotasInput{ServiceCode: aws.String(code)}
	defaultQuotas := make([]*servicequotas.ServiceQuota, 0)
	for {
		outputDefault, err := client.ListAWSDefaultServiceQuotas(inputDefault)
		if err != nil {
			fmt.Println(err)
		}
		defaultQuotas = append(defaultQuotas, outputDefault.Quotas...)
		if outputDefault.NextToken == nil {
			break
		}
		inputDefault.NextToken = outputDefault.NextToken
	}

	input := &servicequotas.ListServiceQuotasInput{ServiceCode: aws.String(code)}
	appliedQuotas := make([]*servicequotas.ServiceQuota, 0)
	for {
		output, err := client.ListServiceQuotas(input)
		if err != nil {
			fmt.Println(err)
		}
		appliedQuotas = append(appliedQuotas, output.Quotas...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	i := map[string][]*servicequotas.ServiceQuota{}
	i["default"] = defaultQuotas
	i["applied"] = appliedQuotas
	json.NewEncoder(os.Stdout).Encode(i)
}

func sessionForRegion(region string) (*session.Session, error) {
	ns, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return ns, nil
}
