package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	reale2e "k8s.io/kubernetes/test/e2e"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/origin/test/extended/util"
	exutilcloud "github.com/openshift/origin/test/extended/util/cloud"

	// Initialize ovirt as a provider
	_ "github.com/openshift/origin/test/extended/util/ovirt"

	// these are loading important global flags that we need to get and set
	_ "k8s.io/kubernetes/test/e2e"
	_ "k8s.io/kubernetes/test/e2e/lifecycle"
)

type TestNameMatchesFunc func(name string) bool

func initializeTestFramework(context *e2e.TestContextType, config *exutilcloud.ClusterConfiguration, dryRun bool) (TestNameMatchesFunc, error) {
	// update context with loaded config
	context.Provider = config.ProviderName
	context.CloudConfig = e2e.CloudConfig{
		ProjectID:   config.ProjectID,
		Region:      config.Region,
		Zone:        config.Zone,
		NumNodes:    config.NumNodes,
		MultiMaster: config.MultiMaster,
		MultiZone:   config.MultiZone,
		ConfigFile:  config.ConfigFile,
	}
	context.AllowedNotReadyNodes = 100
	context.MaxNodesToGather = 0
	reale2e.SetViperConfig(os.Getenv("VIPERCONFIG"))

	// allow the CSI tests to access test data, but only briefly
	// TODO: ideally CSI would not use any of these test methods
	var err error
	exutil.WithCleanup(func() { err = initCSITests(dryRun) })
	if err != nil {
		return nil, err
	}

	if err := exutil.InitTest(dryRun); err != nil {
		return nil, err
	}
	gomega.RegisterFailHandler(ginkgo.Fail)

	e2e.AfterReadingAllFlags(context)
	context.DumpLogsOnFailure = true

	// given the configuration we have loaded, skip tests that our provider should exclude
	// or our network plugin should exclude
	var skips []string
	skips = append(skips, fmt.Sprintf("[Skipped:%s]", config.ProviderName))
	for _, id := range config.NetworkPluginIDs {
		skips = append(skips, fmt.Sprintf("[Skipped:Network/%s]", id))
	}
	matchFn := func(name string) bool {
		for _, skip := range skips {
			if strings.Contains(name, skip) {
				return false
			}
		}
		return true
	}
	return matchFn, nil
}

func decodeProvider(provider string, dryRun, discover bool) (*exutilcloud.ClusterConfiguration, error) {
	return &exutilcloud.ClusterConfiguration{ProviderName: "local"}, nil
}
