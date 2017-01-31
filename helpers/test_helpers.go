package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/routing-api"
	uaaclient "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	cfworkflow_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type RoutingConfig struct {
	config.Config
	RoutingApiUrl     string       `json:"-"` //"-" is used for ignoring field
	Addresses         []string     `json:"addresses"`
	OAuth             *OAuthConfig `json:"oauth"`
	IncludeHttpRoutes bool         `json:"include_http_routes"`
	TcpAppDomain      string       `json:"tcp_app_domain"`
	LBConfigured      bool         `json:"lb_configured"`
}

type OAuthConfig struct {
	TokenEndpoint string `json:"token_endpoint"`
	ClientName    string `json:"client_name"`
	ClientSecret  string `json:"client_secret"`
	Port          int    `json:"port"`
}

func LoadConfig() RoutingConfig {
	loadedConfig := loadConfigJsonFromPath()
	loadedConfig.Config = config.LoadConfig()

	if loadedConfig.OAuth == nil {
		panic("missing configuration oauth")
	}

	if len(loadedConfig.Addresses) == 0 {
		panic("missing configuration 'addresses'")
	}

	if loadedConfig.AppsDomain == "" {
		panic("missing configuration apps_domain")
	}

	if loadedConfig.ApiEndpoint == "" {
		panic("missing configuration api")
	}

	loadedConfig.RoutingApiUrl = fmt.Sprintf("https://%s", loadedConfig.ApiEndpoint)

	return loadedConfig
}

func NewUaaClient(routerApiConfig RoutingConfig, logger lager.Logger) uaaclient.Client {

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

func GetRouterGroupName(routingApiClient routing_api.Client, context cfworkflow_helpers.SuiteContext) string {
	os.Setenv("CF_TRACE", "true")
	var routerGroupName string
	cfworkflow_helpers.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
		routerGroupOutput := cf.Cf("router-groups").Wait(context.ShortTimeout()).Out.Contents()
		routerGroupName = GrabName(string(routerGroupOutput))
	})
	os.Setenv("CF_TRACE", "false")
	return routerGroupName
}

func GrabName(logLines string) string {
	defer GinkgoRecover()
	var re *regexp.Regexp

	re = regexp.MustCompile("name\":\"([a-zA-Z-]*)\"")

	matches := re.FindStringSubmatch(logLines)

	Expect(len(matches)).To(BeNumerically(">=", 2))
	// name
	return matches[1]
}

func UpdateOrgQuota(context cfworkflow_helpers.SuiteContext) {
	os.Setenv("CF_TRACE", "false")
	cfworkflow_helpers.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
		orgGuid := cf.Cf("org", context.RegularUserContext().Org, "--guid").Wait(context.ShortTimeout()).Out.Contents()
		quotaUrl, err := helpers.GetOrgQuotaDefinitionUrl(string(orgGuid), context.ShortTimeout())
		Expect(err).NotTo(HaveOccurred())

		cf.Cf("curl", quotaUrl, "-X", "PUT", "-d", "'{\"total_reserved_route_ports\":-1}'").Wait(context.ShortTimeout())
	})
	os.Setenv("CF_TRACE", "true")
}

func loadConfigJsonFromPath() RoutingConfig {
	var config RoutingConfig

	path := configPath()

	configFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	return config
}

func configPath() string {
	path := os.Getenv("CONFIG")
	if path == "" {
		panic("Must set $CONFIG to point to an integration config .json file.")
	}

	return path
}

func RandomName() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return guid.String()
}
