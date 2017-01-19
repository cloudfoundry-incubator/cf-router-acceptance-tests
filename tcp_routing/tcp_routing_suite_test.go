package tcp_routing_test

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"

	routing_helpers "code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers"
	"code.cloudfoundry.org/routing-api"
	uaaclient "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
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

	context = cfworkflow_helpers.NewContext(routingConfig.Config)
	environment = cfworkflow_helpers.NewEnvironment(context)

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

	routingConfig    helpers.RoutingConfig
	routingApiClient routing_api.Client
	context          cfworkflow_helpers.SuiteContext
	environment      *cfworkflow_helpers.Environment
	logger           lager.Logger
)

var _ = BeforeSuite(func() {
	logger = lagertest.NewTestLogger("test")
	routingApiClient = routing_api.NewClient(routingConfig.RoutingApiUrl, routingConfig.SkipSSLValidation)

	uaaClient := newUaaClient(routingConfig, logger)
	token, err := uaaClient.FetchToken(true)
	Expect(err).ToNot(HaveOccurred())

	routingApiClient.SetToken(token.AccessToken)
	_, err = routingApiClient.Routes()
	Expect(err).ToNot(HaveOccurred(), "Routing API is unavailable")
	domainName = fmt.Sprintf("%s.%s", generator.PrefixedRandomName("TCP", "DOMAIN"), routingConfig.AppsDomain)
	cfworkflow_helpers.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
		routerGroupName := getRouterGroupName(routingApiClient)
		routing_helpers.CreateSharedDomain(domainName, routerGroupName, DEFAULT_TIMEOUT)
		Expect(routing_helpers.GetDomainGuid(domainName, DEFAULT_TIMEOUT)).NotTo(BeEmpty())
	})

	environment.Setup()
})

var _ = AfterSuite(func() {
	environment.Teardown()
	cfworkflow_helpers.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
		routing_helpers.DeleteSharedDomain(domainName, DEFAULT_TIMEOUT)
	})
	CleanupBuildArtifacts()
})

func newUaaClient(routerApiConfig helpers.RoutingConfig, logger lager.Logger) uaaclient.Client {

	tokenURL := fmt.Sprintf("%s:%d", routerApiConfig.OAuth.TokenEndpoint, routerApiConfig.OAuth.Port)

	cfg := &uaaconfig.Config{
		UaaEndpoint:           tokenURL,
		SkipVerification:      routerApiConfig.SkipSSLValidation,
		ClientName:            routerApiConfig.OAuth.ClientName,
		ClientSecret:          routerApiConfig.OAuth.ClientSecret,
		MaxNumberOfRetries:    3,
		RetryInterval:         500 * time.Millisecond,
		ExpirationBufferInSec: 30,
	}

	uaaClient, err := uaaclient.NewClient(logger, cfg, clock.NewClock())
	Expect(err).ToNot(HaveOccurred())

	_, err = uaaClient.FetchToken(true)
	Expect(err).ToNot(HaveOccurred())

	return uaaClient
}

func getRouterGroupName(routingApiClient routing_api.Client) string {
	os.Setenv("CF_TRACE", "true")
	var routerGroupName string
	cfworkflow_helpers.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
		routerGroupOutput := cf.Cf("router-groups").Wait(context.ShortTimeout()).Out.Contents()
		routerGroupName = grabName(string(routerGroupOutput))
	})
	os.Setenv("CF_TRACE", "false")
	return routerGroupName
}

func grabName(logLines string) string {
	defer GinkgoRecover()
	var re *regexp.Regexp

	re = regexp.MustCompile("name\":\"([a-zA-Z-]*)\"")

	matches := re.FindStringSubmatch(logLines)

	Expect(len(matches)).To(BeNumerically(">=", 2))
	// name
	return matches[1]
}
