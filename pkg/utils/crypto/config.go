package crypto

//Config contains rolling upgrade configuration parameters
type Config struct {
	ResourceNamespace   string
	ResourceName        string
	ResourceAnnotations map[string]string
	SHAValue            string
	Type                string
}

//// GetConfigmapConfig provides utility config for configmap
//func GetConfigmapConfig(configmap *v1.ConfigMap) Config {
//	return Config{
//		ResourceNamespace:   configmap.Namespace,
//		ResourceName:        configmap.Name,
//		ResourceAnnotations: configmap.Annotations,
//		SHAValue:            GetSHAfromConfigmap(configmap),
//		Type:                constants.ConfigmapEnvVarPostfix,
//	}
//}
//
//// GetSecretConfig provides utility config for secret
//func GetSecretConfig(secret *v1.Secret) Config {
//	return Config{
//		ResourceNamespace:   secret.Namespace,
//		ResourceName:        secret.Name,
//		ResourceAnnotations: secret.Annotations,
//		SHAValue:            GetSHAfromSecret(secret.Data),
//		Type:                constants.SecretEnvVarPostfix,
//	}
//}
