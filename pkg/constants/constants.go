package constants

const (
	DevopsNamespace = "devops-system"

	// ConfigmapEnvVarPostfix is a postfix for configmap envVar
	ConfigmapEnvVarPostfix = "CONFIGMAP"
	// SecretEnvVarPostfix is a postfix for secret envVar
	SecretEnvVarPostfix = "SECRET"

	// ReloaderAutoAnnotation is an annotation to detect changes in secrets
	ReloaderAutoAnnotation = "reloader.efunds.com/auto"

	EnvVarPrefix = "EFUNDS_"

	KubesphereAppVersion = "app.kubernetes.io/version"
	KubesphereAppName    = "app.kubernetes.io/name"

	NginxUpstreamAnnotation = "nginx.ingress.kubernetes.io/upstream-vhost"
	NginxServiceUpstreamAnnotation = "nginx.ingress.kubernetes.io/service-upstream"

	NginxSendTimeoutAnnotation = "nginx.ingress.kubernetes.io/proxy-send-timeout"
	NginxReadTimeoutAnnotation = "nginx.ingress.kubernetes.io/proxy-read-timeout"
	NginxConnectTimeoutAnnotation = "nginx.ingress.kubernetes.io/proxy-connect-timeout"

	DefaultNginxSendTimeoutAnnotationValue = "5"
	DefaultNginxReadTimeoutAnnotationValue = "5"
	DefaultNginxConnectTimeoutAnnotationValue = "5"
	DefaultNginxServiceUpstreamAnnotationValue = "true"
)
