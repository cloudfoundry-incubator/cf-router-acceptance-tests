package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	uaaclient "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	cfworkflow_helpers "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/gomega"
)

type RoutingConfig struct {
	*config.Config
	RoutingApiUrl     string       `json:"-"` //"-" is used for ignoring field
	Addresses         []string     `json:"addresses"`
	OAuth             *OAuthConfig `json:"oauth"`
	IncludeHttpRoutes bool         `json:"include_http_routes"`
	TcpAppDomain      string       `json:"tcp_apps_domain"`
	LBConfigured      bool         `json:"lb_configured"`
	TCPRouterGroup    string       `json:"tcp_router_group"`
	Xfcc              XFCC         `json:"xfcc"`
}

type OAuthConfig struct {
	TokenEndpoint string `json:"token_endpoint"`
	ClientName    string `json:"client_name"`
	ClientSecret  string `json:"client_secret"`
	Port          int    `json:"port"`
}

type XFCC struct {
	AlwaysForward []string `json:"always_forward"`
	Forward       []string `json:"forward"`
}

func loadDefaultTimeout(conf *RoutingConfig) {
	if conf.DefaultTimeout <= 0 {
		conf.DefaultTimeout = 120
	}

	if conf.CfPushTimeout <= 0 {
		conf.CfPushTimeout = 120
	}
}
func LoadConfig() RoutingConfig {
	loadedConfig := loadConfigJsonFromPath()

	loadedConfig.Config = config.LoadConfig()
	loadDefaultTimeout(&loadedConfig)

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

	if loadedConfig.TCPRouterGroup == "" {
		panic("missing configuration tcp_router_group")
	}

	loadedConfig.RoutingApiUrl = fmt.Sprintf("https://%s", loadedConfig.ApiEndpoint)

	return loadedConfig
}

func ValidateRouterGroupName(context cfworkflow_helpers.UserContext, tcpRouterGroup string) {
	var routerGroupOutput string
	cfworkflow_helpers.AsUser(context, context.Timeout, func() {
		routerGroupOutput = string(cf.Cf("router-groups").Wait(context.Timeout).Out.Contents())
	})

	Expect(routerGroupOutput).To(MatchRegexp(fmt.Sprintf("%s\\s+tcp", tcpRouterGroup)), fmt.Sprintf("Router group %s of type tcp doesn't exist", tcpRouterGroup))
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

func UpdateOrgQuota(context cfworkflow_helpers.UserContext) {
	os.Setenv("CF_TRACE", "false")
	cfworkflow_helpers.AsUser(context, context.Timeout, func() {
		orgGuid := cf.Cf("org", context.Org, "--guid").Wait(context.Timeout).Out.Contents()
		quotaUrl, err := helpers.GetOrgQuotaDefinitionUrl(string(orgGuid), context.Timeout)
		Expect(err).NotTo(HaveOccurred())

		cf.Cf("curl", quotaUrl, "-X", "PUT", "-d", "'{\"total_reserved_route_ports\":-1}'").Wait(context.Timeout)
	})
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

func CreateCertDER(cname string) (*rsa.PrivateKey, []byte) {
	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	Expect(err).ToNot(HaveOccurred())

	subject := pkix.Name{Organization: []string{"xyz, Inc."}}
	if cname != "" {
		subject.CommonName = cname
	}

	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour), // valid for an hour
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  true,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).ToNot(HaveOccurred())
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &privKey.PublicKey, privKey)
	Expect(err).ToNot(HaveOccurred())
	return privKey, certDER
}

func CreateKeyPair(cname string) (keyPEM, certPEM []byte) {
	privKey, certDER := CreateCertDER(cname)

	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	return
}
