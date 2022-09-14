# aws-servicequotas-operator

The `aws-servicequotas-operator` ensures all service quotas are set in each AWS account where workload clusters are running. It only create requests to set the recommended quotas from our [documentation](https://docs.giantswarm.io/getting-started/cloud-provider-accounts/aws/#limits), it won't decrease quotas which are already higher than recommended.

In case you want to add a new quota for a service, you can run the the CLI tool insides `/codes`. Take the `ServiceCode` from `servicecodes.json` and run:

`AWS_REGION=$REGION AWS_ACCESS_KEY_ID=$KEY AWS_SECRET_ACCESS_KEY=$SECRET go run codes/main.go --arn arn:aws:iam::ACCOUNT_ID:role/GiantSwarmAWSOperator --code $SERVICECODE --region $REGION`

This will return all quota codes for your service, e.g.:

```json
    {
      "Adjustable": true,
      "ErrorReason": null,
      "GlobalQuota": false,
      "Period": null,
      "QuotaArn": "arn:aws:servicequotas:eu-west-1:ACCOUNT_ID:autoscaling/L-6B80B8FA",
      "QuotaCode": "L-6B80B8FA",
      "QuotaName": "Launch configurations per region",
      "ServiceCode": "autoscaling",
      "ServiceName": "Amazon EC2 Auto Scaling",
      "Unit": "None",
      "UsageMetric": null,
      "Value": 500
    }
```

You can only add quota codes which are `adjustable`.

Once you have the information of the `QuotaCode`, you can add it to `pkg/quotas/quotas.go`. There's a map for quotas which will be applied.

The key of quotas is the `ServiceCode` which can contain multiple quotas. Add the `QuodaCode from above and ensure it has a reasonable value. 

Also make sure to update the [documentation](https://docs.giantswarm.io/getting-started/cloud-provider-accounts/aws/#limits) once you add new quotas.

After changing everything you only need to release and `aws-servicequotas-operator` with the new version gets applied in each AWS installation.