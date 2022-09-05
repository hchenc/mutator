package handlers

import (
	v1 "k8s.io/api/networking/v1"
)

type filter func(ingress *v1.Ingress, annotationKey, annotationValue string) bool

type IngressHandler struct {
	ingress *v1.Ingress
	Changed bool
}

func NewIngressHandler(ingress *v1.Ingress) *IngressHandler {
	return &IngressHandler{ingress: ingress}
}

func (handler *IngressHandler) Process(f filter, annotationKey, annotationValue string) *IngressHandler {
	handler.Changed = f(handler.ingress, annotationKey, annotationValue) || handler.Changed
	return handler
}

func AnnotationPatchFilter(ingress *v1.Ingress, annotationKey, annotationValue string) bool {
	if value, exists := ingress.Annotations[annotationKey]; exists {
		if value == annotationValue {
			return false
		}
	}
	ingress.Annotations[annotationKey] = annotationValue
	return true
}

func AnnotationNotPatchFilter(ingress *v1.Ingress, annotationKey, annotationValue string) bool {
	if uv, exists := ingress.Annotations[annotationKey]; exists && uv != "" {
		return false
	} else {
		ingress.Annotations[annotationKey] = annotationValue
		return true
	}
}
