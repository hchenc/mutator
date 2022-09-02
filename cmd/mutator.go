/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	application "github.com/hchenc/application/pkg/apis/app/v1beta1"
	servicemeshv1alpha2 "github.com/hchenc/mutator/pkg/apis/servicemesh/v1alpha2"

	"github.com/hchenc/mutator/pkg/constants"
	"github.com/hchenc/mutator/pkg/controllers"
	logger "github.com/hchenc/mutator/pkg/utils/logger"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"os/user"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	leaderElection          bool
	leaderElectionNamespace string
	leaderElectionID        string
	log                     = logger.GetLogger()
)

func NewMutatorCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mutator",
		Short: "A patch tool to mutate Kubernetes resource as needed",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	rootCmd.PersistentFlags().BoolVar(&leaderElection, "leader-election", false, "leader election")
	rootCmd.PersistentFlags().StringVar(&leaderElectionNamespace, "leader-election-namespace", constants.DevopsNamespace, "leader election namespace")
	rootCmd.PersistentFlags().StringVar(&leaderElectionID, "leader-election-id", "resource-mutator-leader-election", "leader election id")

	rootCmd.AddCommand(newRunCommand())
	return rootCmd
}

type KubernetesOptions struct {
	// kubeconfig
	KubeConfig *rest.Config

	// kubeconfig path, if not specified, will use
	// in cluster way to create clientset
	KubeConfigPath string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes apiserver public address, used to generate kubeconfig
	// for downloading, default to host defined in kubeconfig
	// +optional
	Master string `json:"master,omitempty" yaml:"master"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitempty" yaml:"qps"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst"`
}

func NewKubernetesConfig() (option *KubernetesOptions) {
	option = &KubernetesOptions{
		QPS:   1e6,
		Burst: 1e6,
	}

	// make it be easier for those who wants to run api-server locally
	homePath := homedir.HomeDir()
	if homePath == "" {
		// try os/user.HomeDir when $HOME is unset.
		if u, err := user.Current(); err == nil {
			homePath = u.HomeDir
		}
	}

	userHomeConfig := path.Join(homePath, ".kube/config")
	if _, err := os.Stat(userHomeConfig); err == nil {
		option.KubeConfigPath = userHomeConfig
	}
	return
}

func newRunCommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run kubernetes and kubesphere resource operator to patch target data",
		Run: func(cmd *cobra.Command, args []string) {
			var c = NewKubernetesConfig()
			if config, err := clientcmd.BuildConfigFromFlags("", c.KubeConfigPath); err == nil {
				c.KubeConfig = config
			}
			run(c.KubeConfig, signals.SetupSignalHandler())
		},
	}
	return runCmd
}

func run(kubeConfig *rest.Config, ctx context.Context) {
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
	corev1.AddToScheme(mgr.GetScheme())
	appsv1.AddToScheme(mgr.GetScheme())
	servicemeshv1alpha2.AddToScheme(mgr.GetScheme())
	application.AddToScheme(mgr.GetScheme())
	v1.AddToScheme(mgr.GetScheme())
	if err := (&controllers.IngressOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: log,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err)
	}
	if err := mgr.Start(ctx); err != nil {
		log.Error(err)

	}

}

func main() {
	cmd := NewMutatorCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

}
