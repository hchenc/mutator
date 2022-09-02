package controllers

import (
	"context"
	"github.com/hchenc/mutator/pkg/constants"
	"github.com/hchenc/mutator/pkg/controllers/predicates"
	"github.com/hchenc/mutator/pkg/handlers"
	"github.com/hchenc/mutator/pkg/utils/logger"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ConfigMapOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *logrus.Logger
}

func (c *ConfigMapOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	configmap := &corev1.ConfigMap{}

	if err := c.Get(ctx, req.NamespacedName, configmap); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			c.Logger.Errorf("Failed to reconcile secret for %s", err)
			return ctrl.Result{}, err
		}
	}

	log := c.Logger.WithFields(logger.GetObjectFields(configmap))

	if configmap.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	deploymentList := &appsv1.DeploymentList{}

	if err := c.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
		log.Errorf("Failed to list deployments for %s", err)
		return ctrl.Result{}, err
	}

	handler := handlers.NewDeploymentHandler(deploymentList.Items, configmap).
		For(configmap).
		WithFilter(
			&handlers.ReloadOrNotFilter{FilterAnnotation: constants.ReloaderAutoAnnotation}).
		WithFilter(
			&handlers.VolumeNameFilter{}).
		Record()

	for _, deployment := range handler.GetDeploymentList() {
		result := updateContainerEnvVars(handler, &deployment)
		if result == constants.Updated {
			if err := c.Update(ctx, &deployment, &client.UpdateOptions{FieldManager: "Reloader"}); err != nil {
				log.Errorf("Update deployment failed for %s", err)
			} else {
				log.Info("Update deployment succeeded")
			}
		}
	}
	return ctrl.Result{}, nil

}

func (c *ConfigMapOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(
			predicates.And(
				&predicates.NamespaceUpdatePredicate{
					IncludeNamespaces: predicates.DefaultIncludeNamespaces,
				},
				&predicates.UpdatePredicate{},
			),
		).
		Complete(c)
}
