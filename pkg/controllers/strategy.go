package controllers

import (
	"context"
	"github.com/hchenc/application/pkg/apis/app/v1beta1"
	servicemeshv1alpha2 "github.com/hchenc/mutator/pkg/apis/servicemesh/v1alpha2"
	"github.com/hchenc/mutator/pkg/constants"
	filter "github.com/hchenc/mutator/pkg/controllers/predicates"
	"github.com/hchenc/mutator/pkg/utils/logger"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type StrategyOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *logrus.Logger
}

func (s *StrategyOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	strategy := &servicemeshv1alpha2.Strategy{}

	if err := s.Get(ctx, req.NamespacedName, strategy); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			s.Logger.Errorf("Failed to reconcile strategy for %s", err)
			return ctrl.Result{}, err
		}
	}

	log := s.Logger.WithFields(logger.GetObjectFields(strategy))

	if strategy.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if appName, exist := strategy.Labels[constants.KubesphereAppName]; exist {
		application := &v1beta1.Application{}

		appNamespacedName := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      appName,
		}
		err := s.Get(ctx, appNamespacedName, application)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			} else {
				log.Errorf("Failed to get application for %s", err)
				return ctrl.Result{}, err
			}
		}

		application.Labels[constants.KubesphereAppVersion] = strategy.Spec.GovernorVersion
		if err := s.Update(ctx, application); err != nil {
			log.Errorf("Update application failed for %s", err)
		}

		log.Info("Update application succeeded")
		return reconcile.Result{}, nil
	} else {
		log.Errorf("No label <%s> exists", constants.KubesphereAppName)
		return reconcile.Result{}, nil
	}
}

func (s *StrategyOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicemeshv1alpha2.Strategy{}).
		WithEventFilter(
			predicate.And(
				&filter.NamespaceUpdatePredicate{
					IncludeNamespaces: filter.DefaultIncludeNamespaces,
				},
				&strategyUpdatePredicate{},
			),
		).
		Complete(s)
}

type strategyUpdatePredicate struct {
	filter.NamespaceUpdatePredicate
}

func (s strategyUpdatePredicate) Update(e event.UpdateEvent) bool {
	if references := e.ObjectOld.GetOwnerReferences(); references == nil {
		return false
	}
	return true
}
