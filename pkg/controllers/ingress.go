package controllers

import (
	"context"
	"fmt"
	"github.com/hchenc/mutator/pkg/constants"
	"github.com/hchenc/mutator/pkg/controllers/predicates"
	"github.com/hchenc/mutator/pkg/handlers"
	"github.com/hchenc/mutator/pkg/utils/logger"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type IngressOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *logrus.Logger
}

func (i *IngressOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ingress := &v1.Ingress{}

	err := i.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			i.Logger.Errorf("Failed to reconcile ingress for %s", err)
			return ctrl.Result{}, err
		}
	}
	log := i.Logger.WithFields(logger.GetObjectFields(ingress))

	if ingress.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	handler := handlers.NewIngressHandler(ingress)

	serviceName := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name
	namespaceName := ingress.Namespace
	upstreamVhost := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespaceName)

	handler.
		Process(handlers.AnnotationPatchFilter, constants.NginxUpstreamAnnotation, upstreamVhost).
		Process(handlers.AnnotationPatchFilter, constants.NginxServiceUpstreamAnnotation, constants.DefaultNginxServiceUpstreamAnnotationValue).
		Process(handlers.AnnotationNotPatchFilter, constants.NginxConnectTimeoutAnnotation, constants.DefaultNginxConnectTimeoutAnnotationValue).
		Process(handlers.AnnotationNotPatchFilter, constants.NginxReadTimeoutAnnotation, constants.DefaultNginxReadTimeoutAnnotationValue).
		Process(handlers.AnnotationNotPatchFilter, constants.NginxSendTimeoutAnnotation, constants.DefaultNginxSendTimeoutAnnotationValue)

	if handler.Changed {
		if err := i.Update(ctx, ingress); err != nil {
			log.Errorf("Update ingress failed for %s", err)
			return ctrl.Result{}, err
		}
		log.Info("Update ingress annotation succeeded")

	}
	return reconcile.Result{}, nil
}

func (i *IngressOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Ingress{}).
		WithEventFilter(
			&predicates.NamespaceCreatePredicate{
				IncludeNamespaces: predicates.DefaultIncludeNamespaces,
			}).
		Complete(i)
}
