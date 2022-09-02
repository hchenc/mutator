package predicates

import "sigs.k8s.io/controller-runtime/pkg/event"

type NamespaceCreatePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

func (n NamespaceCreatePredicate) Create(e event.CreateEvent) bool {
	namespace := e.Object.GetNamespace()

	if exists, verified := checkIndexKey(n.IncludeNamespaces, namespace); verified {
		return exists
	}

	if exists, verified := checkIndexKey(n.ExcludeNamespaces, namespace); verified {
		return !exists
	}

	return false

}