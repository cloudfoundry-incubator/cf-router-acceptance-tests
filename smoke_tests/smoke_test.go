package smoke_test

import (
	"fmt"
	"net/http"
	"os"
	"time"

	routing_helpers "code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers/assets"
	"code.cloudfoundry.org/routing-api"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	cfworkflow_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	DEFAULT_TIMEOUT          = 2 * time.Minute
	DEFAULT_POLLING_INTERVAL = 5 * time.Second
	CF_PUSH_TIMEOUT          = 2 * time.Minute
	routingConfig            helpers.RoutingConfig
	domainName               string
	appName                  string
	tcpSampleGolang          = assets.NewAssets().TcpSampleGolang
	adminContext             cfworkflow_helpers.UserContext
	context                  cfworkflow_helpers.SuiteContext
	environment              *cfworkflow_helpers.Environment
)
var _ = Describe("SmokeTests", func() {
	Context("when router api is enabled", func() {
		BeforeEach(func() {
			routingConfig = helpers.LoadConfig()
			os.Setenv("CF_TRACE", "true")
			context = cfworkflow_helpers.NewContext(routingConfig.Config)
			environment = cfworkflow_helpers.NewEnvironment(context)

			logger := lagertest.NewTestLogger("test")
			routingApiClient := routing_api.NewClient(routingConfig.RoutingApiUrl, routingConfig.SkipSSLValidation)

			uaaClient := helpers.NewUaaClient(routingConfig, logger)
			token, err := uaaClient.FetchToken(true)
			Expect(err).ToNot(HaveOccurred())

			routingApiClient.SetToken(token.AccessToken)
			_, err = routingApiClient.Routes()
			Expect(err).ToNot(HaveOccurred(), "Routing API is unavailable")

			adminContext = context.AdminUserContext()
			regUser := context.RegularUserContext()
			adminContext.Org = regUser.Org
			adminContext.Space = regUser.Space

			environment.Setup()

			if routingConfig.TcpAppDomain != "" {
				domainName = routingConfig.TcpAppDomain
				cfworkflow_helpers.AsUser(adminContext, context.ShortTimeout(), func() {
					routing_helpers.VerifySharedDomain(routingConfig.TcpAppDomain, DEFAULT_TIMEOUT)
				})
			} else {
				domainName = fmt.Sprintf("%s.%s", generator.PrefixedRandomName("TCP", "DOMAIN"), routingConfig.AppsDomain)

				cfworkflow_helpers.AsUser(adminContext, context.ShortTimeout(), func() {
					routerGroupName := helpers.GetRouterGroupName(routingApiClient, context)
					routing_helpers.CreateSharedDomain(domainName, routerGroupName, DEFAULT_TIMEOUT)
					routing_helpers.VerifySharedDomain(domainName, DEFAULT_TIMEOUT)
				})
			}
			appName = routing_helpers.GenerateAppName()
			helpers.UpdateOrgQuota(context)
		})

		It("map tcp route to app successfully ", func() {
			routing_helpers.PushAppNoStart(appName, tcpSampleGolang, routingConfig.GoBuildpackName, domainName, CF_PUSH_TIMEOUT, "256M", "--no-route")
			routing_helpers.EnableDiego(appName, DEFAULT_TIMEOUT)
			routing_helpers.MapRandomTcpRouteToApp(appName, domainName, DEFAULT_TIMEOUT)
			routing_helpers.StartApp(appName, DEFAULT_TIMEOUT)
			port := routing_helpers.GetPortFromAppsInfo(appName, domainName, DEFAULT_TIMEOUT)

			var addr string
			if !routingConfig.LBConfigured {
				addr = routingConfig.Addresses[0]
			} else {
				addr = domainName
			}

			appUrl := fmt.Sprintf("http://%s:%s", addr, port)
			resp, err := http.Get(appUrl)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			routing_helpers.DeleteTcpRoute(domainName, port, DEFAULT_TIMEOUT)

			_, err = http.Get(appUrl)
			Expect(err).To(HaveOccurred())
		})
	})

	AfterEach(func() {
		routing_helpers.DeleteApp(appName, DEFAULT_TIMEOUT)
		if routingConfig.TcpAppDomain == "" {
			routing_helpers.DeleteSharedDomain(domainName, DEFAULT_TIMEOUT)
		}
		environment.Teardown()
	})
})
