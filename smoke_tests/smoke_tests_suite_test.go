package smoke_test

import (
	"os"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers"
	"code.cloudfoundry.org/routing-api"
	cf_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	cfworkflow_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
)

var (
	DEFAULT_TIMEOUT          = 2 * time.Minute
	DEFAULT_POLLING_INTERVAL = 5 * time.Second
	CF_PUSH_TIMEOUT          = 2 * time.Minute
	routingConfig            helpers.RoutingConfig
	environment              *cfworkflow_helpers.ReproducibleTestSuiteSetup
)

func TestSmokeTests(t *testing.T) {
	routingConfig = helpers.LoadConfig()
	RegisterFailHandler(Fail)
	componentName := "SmokeTests Suite"
	rs := []Reporter{}
	if routingConfig.ArtifactsDirectory != "" {
		cf_helpers.EnableCFTrace(routingConfig.Config, componentName)
		rs = append(rs, cf_helpers.NewJUnitReporter(routingConfig.Config, componentName))
	}
	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)

}

var _ = BeforeSuite(func() {
	if routingConfig.DefaultTimeout > 0 {
		DEFAULT_TIMEOUT = time.Duration(routingConfig.DefaultTimeout) * time.Second
	}

	if routingConfig.CfPushTimeout > 0 {
		CF_PUSH_TIMEOUT = time.Duration(routingConfig.CfPushTimeout) * time.Second
	}

	os.Setenv("CF_TRACE", "true")
	environment = cfworkflow_helpers.NewTestSuiteSetup(routingConfig)

	logger := lagertest.NewTestLogger("test")
	routingApiClient := routing_api.NewClient(routingConfig.RoutingApiUrl, routingConfig.SkipSSLValidation)

	uaaClient := helpers.NewUaaClient(routingConfig, logger)
	token, err := uaaClient.FetchToken(true)
	Expect(err).ToNot(HaveOccurred())

	routingApiClient.SetToken(token.AccessToken)
	_, err = routingApiClient.Routes()
	Expect(err).ToNot(HaveOccurred(), "Routing API is unavailable")
})

var _ = AfterSuite(func() {
	environment.Teardown()
	CleanupBuildArtifacts()
})
