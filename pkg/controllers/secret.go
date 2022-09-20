package controllers

import (
	"context"
	"github.com/hchenc/mutator/pkg/constants"
	filter "github.com/hchenc/mutator/pkg/controllers/predicates"
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

type SecretOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *logrus.Logger
}

func (s *SecretOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &corev1.Secret{}

	if err := s.Get(ctx, req.NamespacedName, secret); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			s.Logger.Errorf("Failed to reconcile secret for %s", err)
			return ctrl.Result{}, err
		}
	}
	log := s.Logger.WithFields(logger.GetObjectFields(secret))

	if secret.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	deploymentList := &appsv1.DeploymentList{}

	if err := s.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
		log.Errorf("Failed to list deployments for %s", err)
		return ctrl.Result{}, err
	}

	handler := handlers.NewDeploymentHandler(deploymentList.Items, secret).
		For(secret).
		WithFilter(
			&handlers.ReloadOrNotFilter{FilterAnnotation: constants.ReloaderAutoAnnotation}).
		WithFilter(
			&handlers.VolumeNameFilter{}).
		Record()

	for _, deployment := range handler.GetDeploymentList() {
		result := updateContainerEnvVars(handler, &deployment)
		if result == constants.Updated {
			if err := s.Update(ctx, &deployment, &client.UpdateOptions{FieldManager: "Reloader"}); err != nil {
				log.Errorf("Update deployment failed for %s", err)
			} else {
				log.Info("Reload deployment succeeded")
			}
		}
	}
	return ctrl.Result{}, nil

}

func (s *SecretOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(
			predicate.And(
				&filter.NamespaceUpdatePredicate{
					IncludeNamespaces: filter.DefaultIncludeNamespaces,
				},
				&filter.UpdatePredicate{},
			),
		).
		Complete(s)
}
