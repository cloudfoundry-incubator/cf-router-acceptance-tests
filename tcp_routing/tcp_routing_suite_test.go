package tcp_routing_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"

	routing_helpers "code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers"
	"code.cloudfoundry.org/routing-api"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	cf_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	cfworkflow_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

func TestTcpRouting(t *testing.T) {
	RegisterFailHandler(Fail)

	routingConfig = helpers.LoadConfig()
	if routingConfig.DefaultTimeout > 0 {
		DEFAULT_TIMEOUT = routingConfig.DefaultTimeout * time.Second
	}

	if routingConfig.CfPushTimeout > 0 {
		CF_PUSH_TIMEOUT = routingConfig.CfPushTimeout * time.Second
	}
	componentName := "TCP Routing"

	rs := []Reporter{}

	if routingConfig.ArtifactsDirectory != "" {
		cf_helpers.EnableCFTrace(routingConfig.Config, componentName)
		rs = append(rs, cf_helpers.NewJUnitReporter(routingConfig.Config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

const preallocatedExternalPorts = 100

var (
	DEFAULT_TIMEOUT          = 2 * time.Minute
	DEFAULT_POLLING_INTERVAL = 5 * time.Second
	CF_PUSH_TIMEOUT          = 2 * time.Minute
	domainName               string

	adminContext     cfworkflow_helpers.UserContext
	routingConfig    helpers.RoutingConfig
	routingApiClient routing_api.Client
	context          cfworkflow_helpers.SuiteContext
	environment      *cfworkflow_helpers.Environment
	logger           lager.Logger
)

var _ = BeforeSuite(func() {
	context = cfworkflow_helpers.NewContext(routingConfig.Config)
	environment = cfworkflow_helpers.NewEnvironment(context)

	logger = lagertest.NewTestLogger("test")
	routingApiClient = routing_api.NewClient(routingConfig.RoutingApiUrl, routingConfig.SkipSSLValidation)

	uaaClient := helpers.NewUaaClient(routingConfig, logger)
	token, err := uaaClient.FetchToken(true)
	Expect(err).ToNot(HaveOccurred())

	routingApiClient.SetToken(token.AccessToken)
	_, err = routingApiClient.Routes()
	Expect(err).ToNot(HaveOccurred(), "Routing API is unavailable")
	domainName = fmt.Sprintf("%s.%s", generator.PrefixedRandomName("TCP", "DOMAIN"), routingConfig.AppsDomain)

	adminContext = context.AdminUserContext()
	regUser := context.RegularUserContext()
	adminContext.Org = regUser.Org
	adminContext.Space = regUser.Space

	environment.Setup()

	cfworkflow_helpers.AsUser(adminContext, context.ShortTimeout(), func() {
		routerGroupName := helpers.GetRouterGroupName(routingApiClient, context)
		routing_helpers.CreateSharedDomain(domainName, routerGroupName, DEFAULT_TIMEOUT)
		routing_helpers.VerifySharedDomain(domainName, DEFAULT_TIMEOUT)
	})

})

var _ = AfterSuite(func() {
	cfworkflow_helpers.AsUser(adminContext, context.ShortTimeout(), func() {
		routing_helpers.DeleteSharedDomain(domainName, DEFAULT_TIMEOUT)
	})
	environment.Teardown()
	CleanupBuildArtifacts()
})
