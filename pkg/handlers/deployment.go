package handlers

import (
	"github.com/hchenc/mutator/pkg/constants"
	"github.com/hchenc/mutator/pkg/utils/crypto"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type FilterChain interface {
	Filter([]appsv1.Deployment, client.Object, map[string]string) []appsv1.Deployment
}

type DeploymentHandler struct {
	err           []error
	shaValue      string
	mountType     string
	volumeNameMap map[string]string
	input         client.Object
	filters       []FilterChain
	deployments   []appsv1.Deployment
}

func NewDeploymentHandler(deployments []appsv1.Deployment, input client.Object) *DeploymentHandler {
	return &DeploymentHandler{
		deployments: deployments,
		//input:       input,
		mountType: input.GetObjectKind().GroupVersionKind().Kind,
	}
}

func (handler *DeploymentHandler) WithFilter(filter FilterChain) *DeploymentHandler {
	handler.filters = append(handler.filters, filter)
	return handler
}

func (handler *DeploymentHandler) GetDeploymentList() []appsv1.Deployment {
	return handler.deployments
}

func (handler *DeploymentHandler) GetSHAValue() string {
	return handler.shaValue
}

func (handler *DeploymentHandler) GetMountType() string {
	return handler.mountType
}

func (handler *DeploymentHandler) GetInput() client.Object {
	return handler.input
}

func (handler *DeploymentHandler) GetRecordMap() map[string]string {
	return handler.volumeNameMap
}

func (handler *DeploymentHandler) setUp() {
	var err error
	switch object := handler.input.(type) {
	case *corev1.ConfigMap:
		handler.shaValue, err = crypto.GetSHAfromConfigmap(object)
	case *corev1.Secret:
		handler.shaValue, err = crypto.GetSHAfromSecret(object.Data)
	}
	if err != nil {
		handler.err = append(handler.err, err)
	}
}

func (handler *DeploymentHandler) For(object client.Object) *DeploymentHandler {
	handler.input = object
	handler.setUp()
	return handler
}

func (handler *DeploymentHandler) Complete() *DeploymentHandler {
	for _, filter := range handler.filters {
		handler.deployments = filter.Filter(handler.deployments, handler.input, nil)
	}
	return handler
}

func (handler *DeploymentHandler) Record() *DeploymentHandler {
	handler.volumeNameMap = map[string]string{}
	for _, filter := range handler.filters {
		handler.deployments = filter.Filter(handler.deployments, handler.input, handler.volumeNameMap)
	}
	return handler
}

type ReloadOrNotFilter struct {
	FilterAnnotation string `validate:"required"`
}

func (filter ReloadOrNotFilter) Filter(deployments []appsv1.Deployment, object client.Object, desireMap map[string]string) []appsv1.Deployment {
	var filteredDeploymentList []appsv1.Deployment
	for _, deployment := range deployments {
		if value, exist := deployment.Annotations[filter.FilterAnnotation]; exist {
			if noAction, _ := strconv.ParseBool(value); !noAction {
				continue
			}
		}
		filteredDeploymentList = append(filteredDeploymentList, deployment)
	}
	return filteredDeploymentList
}

type VolumeNameFilter struct {
}

func (filter VolumeNameFilter) Filter(deployments []appsv1.Deployment, object client.Object, desireMap map[string]string) []appsv1.Deployment {
	var filteredDeploymentList []appsv1.Deployment
	for _, deployment := range deployments {
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == object.GetName() {
				filteredDeploymentList = append(filteredDeploymentList, deployment)
				if desireMap != nil {
					desireMap[deployment.Name] = volume.Name
				}
				break
			} else if volume.Secret != nil && volume.Secret.SecretName == object.GetName() {
				filteredDeploymentList = append(filteredDeploymentList, deployment)
				if desireMap != nil {
					desireMap[deployment.Name] = volume.Name
				}
				break
			} else {
				continue
			}
		}
	}
	return filteredDeploymentList
}

func updateContainerEnvVars(handler DeploymentHandler, deployment *appsv1.Deployment) constants.Result {
	envVar := constants.EnvVarPrefix + crypto.ConvertToEnvVarName(handler.GetInput().GetName()) + "_" + crypto.ConvertToEnvVarName(handler.GetMountType())
	for id := range deployment.Spec.Template.Spec.Containers {
		for _, volume := range deployment.Spec.Template.Spec.Containers[id].VolumeMounts {
			if volume.Name != handler.GetRecordMap()[deployment.Name] {
				continue
			}
			result := updateEnvVar(deployment.Spec.Template.Spec.Containers[id], envVar, handler.GetSHAValue())
			if result == constants.NoEnvVarFound {
				e := corev1.EnvVar{
					Name:  envVar,
					Value: handler.GetSHAValue(),
				}
				deployment.Spec.Template.Spec.Containers[id].Env = append(deployment.Spec.Template.Spec.Containers[id].Env, e)
				return constants.Updated
			}
			return result
		}
	}
	return constants.NoContainerFound

}

func updateEnvVar(container corev1.Container, envVar string, shaData string) constants.Result {
	envs := container.Env
	for j := range envs {
		if envs[j].Name == envVar {
			if envs[j].Value != shaData {
				envs[j].Value = shaData
				return constants.Updated
			}
			return constants.NotUpdated
		}
	}
	return constants.NoEnvVarFound
}
