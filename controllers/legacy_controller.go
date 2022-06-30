/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"regexp"
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/aws-servicequotas-operator/pkg/aws/scope"
	"github.com/giantswarm/aws-servicequotas-operator/pkg/quotas"
)

// AWSLegcyClusterReconciler reconciles a Giant Swarm AWSCluster object
type AWSLegacyClusterReconciler struct {
	client.Client
	DryRun bool
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=giantswarm.io.giantswarm.io,resources=awsclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=giantswarm.io.giantswarm.io,resources=awsclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=giantswarm.io.giantswarm.io,resources=awsclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AWSCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AWSLegacyClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	logger := r.Log.WithValues("namespace", req.Namespace, "cluster", req.Name)

	cluster := &infrastructurev1alpha3.AWSCluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, nil
	}

	// fetch ARN from the cluster to assume role for creating dependencies
	credentialName := cluster.Spec.Provider.CredentialSecret.Name
	credentialNamespace := cluster.Spec.Provider.CredentialSecret.Namespace
	var credentialSecret = &v1.Secret{}
	if err = r.Get(ctx, types.NamespacedName{Namespace: credentialNamespace, Name: credentialName}, credentialSecret); err != nil {
		logger.Error(err, "failed to get credential secret")
		return ctrl.Result{}, microerror.Mask(err)
	}

	secretByte, ok := credentialSecret.Data["aws.awsoperator.arn"]
	if !ok {
		logger.Error(err, "Unable to extract ARN from secret")
		return ctrl.Result{}, microerror.Mask(fmt.Errorf("Unable to extract ARN from secret %s for cluster %s", credentialName, cluster.Name))

	}

	// convert secret data secretByte into string
	arn := string(secretByte)

	// extract AccountID from ARN
	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
	accountID := re.FindAllString(arn, 1)[0]

	if accountID == "" {
		logger.Error(err, "Unable to extract Account ID from ARN")
		return ctrl.Result{}, microerror.Mask(fmt.Errorf("Unable to extract Account ID from ARN %s", string(arn)))

	}

	// create the cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		AccountId:        accountID,
		ARN:              arn,
		ClusterName:      cluster.Name,
		ClusterNamespace: cluster.Namespace,
		Region:           cluster.Spec.Provider.Region,

		Logger:  logger,
		Cluster: cluster,
	})
	if err != nil {
		return reconcile.Result{}, microerror.Mask(err)
	}

	if cluster.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}
	quotas.New(clusterScope, r.Client, r.DryRun).Reconcile(ctx)
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: time.Minute * 30,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSLegacyClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha3.AWSCluster{}).
		Complete(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	return nil
}
