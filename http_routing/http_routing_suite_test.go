package http_routing_test

import (
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers/assets"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"

	cf_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	cfworkflow_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

var (
	DEFAULT_TIMEOUT = 2 * time.Minute
	//	DEFAULT_POLLING_INTERVAL = 5 * time.Second
	CF_PUSH_TIMEOUT = 2 * time.Minute

	asset         assets.Assets
	adminContext  cfworkflow_helpers.UserContext
	routingConfig helpers.RoutingConfig
	environment   *cfworkflow_helpers.ReproducibleTestSuiteSetup
	logger        lager.Logger
)

func TestHttpRouting(t *testing.T) {
	RegisterFailHandler(Fail)
	routingConfig = helpers.LoadConfig()

	if routingConfig.DefaultTimeout > 0 {
		DEFAULT_TIMEOUT = time.Duration(routingConfig.DefaultTimeout) * time.Second
	}

	if routingConfig.CfPushTimeout > 0 {
		CF_PUSH_TIMEOUT = time.Duration(routingConfig.CfPushTimeout) * time.Second
	}

	componentName := "HttpRouting Suite"

	rs := []Reporter{}

	if routingConfig.ArtifactsDirectory != "" {
		cf_helpers.EnableCFTrace(routingConfig.Config, componentName)
		rs = append(rs, cf_helpers.NewJUnitReporter(routingConfig.Config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

var _ = BeforeSuite(func() {
	logger = lagertest.NewTestLogger("test")

	asset = assets.NewAssets()
	environment = cfworkflow_helpers.NewTestSuiteSetup(routingConfig.Config)
	adminContext = environment.AdminUserContext()
	regUser := environment.RegularUserContext()
	adminContext.TestSpace = regUser.TestSpace
	adminContext.Org = regUser.Org
	adminContext.Space = regUser.Space

	environment.Setup()
})

var _ = AfterSuite(func() {
	environment.Teardown()
	CleanupBuildArtifacts()
})
