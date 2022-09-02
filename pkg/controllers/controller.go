package controllers

import (
	"github.com/hchenc/mutator/pkg/constants"
	"github.com/hchenc/mutator/pkg/handlers"
	"github.com/hchenc/mutator/pkg/utils/crypto"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func updateContainerEnvVars(handler *handlers.DeploymentHandler, deployment *appsv1.Deployment) constants.Result {
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
