package server

import (
	"context"

	application "github.com/hchenc/application/pkg/apis/app/v1beta1"
	servicemeshv1alpha2 "github.com/hchenc/mutator/pkg/apis/servicemesh/v1alpha2"
	"github.com/hchenc/mutator/pkg/constants"
	"github.com/hchenc/mutator/pkg/controllers"
	"github.com/hchenc/mutator/pkg/utils/logger"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	urlruntime "k8s.io/apimachinery/pkg/util/runtime"

	virtualservicev1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

var (
	leaderElection          bool
	leaderElectionNamespace string
	leaderElectionID        string
	log                     = logger.GetLogger()
)

func NewServerCommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "server",
		Short: "Run kubernetes and kubesphere resource operator to patch target data",
		Run: func(cmd *cobra.Command, args []string) {
			var c = NewKubernetesConfig()
			if config, err := clientcmd.BuildConfigFromFlags("", c.KubeConfigPath); err == nil {
				c.KubeConfig = config
			}
			server(c.KubeConfig, signals.SetupSignalHandler())
		},
	}
	runCmd.Flags().BoolVar(&leaderElection, "leader-election", false, "leader election")
	runCmd.Flags().StringVar(&leaderElectionNamespace, "leader-election-namespace", constants.DevopsNamespace, "leader election namespace")
	runCmd.Flags().StringVar(&leaderElectionID, "leader-election-id", "resource-mutator-leader-election", "leader election id")

	return runCmd
}

func server(kubeConfig *rest.Config, ctx context.Context) {
	scheme := runtime.NewScheme()
	mgrOptions := manager.Options{
		Scheme: scheme,
		Port:   9443,
	}
	if leaderElection {
		mgrOptions.LeaderElection = true
		mgrOptions.LeaderElectionID = leaderElectionID
		mgrOptions.LeaderElectionNamespace = leaderElectionNamespace
	}
	log.Info("setting up manager")

	mgr, err := manager.New(kubeConfig, mgrOptions)
	if err != nil {
		log.Fatalf("unable to set up mutator controllers manager", err)
	}

	urlruntime.Must(corev1.AddToScheme(mgr.GetScheme()))
	urlruntime.Must(appsv1.AddToScheme(mgr.GetScheme()))
	urlruntime.Must(servicemeshv1alpha2.AddToScheme(mgr.GetScheme()))
	urlruntime.Must(application.AddToScheme(mgr.GetScheme()))
	urlruntime.Must(v1.AddToScheme(mgr.GetScheme()))
	urlruntime.Must(virtualservicev1alpha3.AddToScheme(mgr.GetScheme()))

	if err := (&controllers.ConfigMapOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: log,
	}).SetupWithManager(mgr); err != nil {
		log.Fatal(err)
	}

	if err := (&controllers.IngressOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: log,
	}).SetupWithManager(mgr); err != nil {
		log.Fatal(err)
	}

	if err := (&controllers.SecretOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: log,
	}).SetupWithManager(mgr); err != nil {
		log.Fatal(err)
	}

	if err := (&controllers.StrategyOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: log,
	}).SetupWithManager(mgr); err != nil {
		log.Fatal(err)
	}

	if err := (&controllers.VirtualServiceOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: log,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err)
	}

	if err := mgr.Start(ctx); err != nil {
		log.Fatal(err)
	}

}
