package controllers

import (
	"context"
	"github.com/hchenc/mutator/pkg/controllers/predicates"
	"github.com/hchenc/mutator/pkg/utils/logger"
	"github.com/sirupsen/logrus"
	"istio.io/api/networking/v1alpha3"
	virtualservicev1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)


type VirtualServiceOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *logrus.Logger
}

func (v *VirtualServiceOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	vs := &virtualservicev1alpha3.VirtualService{}

	if err := v.Get(ctx, req.NamespacedName, vs); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			v.Logger.Errorf("Failed to reconcile secret for %s", err)
			return ctrl.Result{}, err
		}
	}

	log := v.Logger.WithFields(logger.GetObjectFields(vs))

	if vs.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	for _, http := range vs.Spec.Http {
		if http.Retries == nil {
			http.Retries = &v1alpha3.HTTPRetry{
				Attempts: int32(0),
			}
		}
	}
	if err := v.Update(ctx, vs); err != nil {
		log.Errorf("Update virtualService failed for %s", err)
		return ctrl.Result{}, err
	}
	log.Info("Update virtualService succeeded")
	return ctrl.Result{}, nil
}

func (v *VirtualServiceOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&virtualservicev1alpha3.VirtualService{}).
		WithEventFilter(
			&predicates.NamespaceCreatePredicate{
				IncludeNamespaces: predicates.DefaultIncludeNamespaces,
			}).
		Complete(v)
}