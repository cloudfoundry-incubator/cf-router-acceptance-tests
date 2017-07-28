package http_routing_test

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	routing_helpers "code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"code.cloudfoundry.org/routing-acceptance-tests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: do NOT run the tests if not config
const certPEM = `-----BEGIN CERTIFICATE-----
MIIDNzCCAh+gAwIBAgIRALduDptTUbCSEXYYAdOoJNgwDQYJKoZIhvcNAQELBQAw
OTEMMAoGA1UEBhMDVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MREwDwYDVQQD
Ewhyb3V0ZXJDQTAeFw0xNzA3MjYyMzE5MTNaFw0xODA3MjYyMzE5MTNaMDQxDDAK
BgNVBAYTA1VTQTEWMBQGA1UEChMNQ2xvdWQgRm91bmRyeTEMMAoGA1UEAxMDcmF0
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0COlp3i9iDbwUeHSx03j
lN98pX4gzke3sgNVYENLfqia701PLjgMe2k0H51QxDf5BnHDA/bnAKzLMZtrhzy9
A6VrFHFHCElFen/l6t8Aq9zXT/otN+LsiM/hL8/OlJjEOlnjBr1nRfhS+GCR25sG
f2k7XA6sKXET+R5W6NLqb3RiJEHAgqjBMRQgph91vC0D9ygWAPSNvSGMsFbWkA1S
KANbv+pi4lk2Xt8t8M5PbI+aLX4mcDs2eIfjXuZUeNaTfvudvF8LDk7prshLfZr2
TlkI3oQmxfqx640d1PbUIrIfU/xK/zkTMEvmvypjrI8wrf1y/56c3EfBWPPmOADE
dQIDAQABoz8wPTAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwIG
CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcNAQELBQADggEBAB3JaY8e
AA7y5nwuBCWKL9jSLRZShwDOfdrAN49LCV2DRxRQsEV2ok80/LTM93Hb1JzbplOA
QvwKxrTy1P8i+F0YP8nPxPs3pezUuA7C4m08JJrTxsHpC+OGcgzVBqc7abppp9MR
gP3LlzWnUdsLBdLd7iWZHoJwQOZyL/5CaZdSQ/WcpzgzIL2ewasBNaQ8WQ7P2XEw
wRiWFDFw4Jjl8YYP7aO+JCKoKfE5USX/i/FAAyGqFJ4vDQxH437aR7tITkhRSdBo
9z4nnEVmHN+LKUsdJR4mnq2GN1SiJsBbIwRk8uIBqW8Su5x5ehKlpkVyluYoja+u
ZYH4aOvZ2+bCcZI=
-----END CERTIFICATE-----`

const keyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA0COlp3i9iDbwUeHSx03jlN98pX4gzke3sgNVYENLfqia701P
LjgMe2k0H51QxDf5BnHDA/bnAKzLMZtrhzy9A6VrFHFHCElFen/l6t8Aq9zXT/ot
N+LsiM/hL8/OlJjEOlnjBr1nRfhS+GCR25sGf2k7XA6sKXET+R5W6NLqb3RiJEHA
gqjBMRQgph91vC0D9ygWAPSNvSGMsFbWkA1SKANbv+pi4lk2Xt8t8M5PbI+aLX4m
cDs2eIfjXuZUeNaTfvudvF8LDk7prshLfZr2TlkI3oQmxfqx640d1PbUIrIfU/xK
/zkTMEvmvypjrI8wrf1y/56c3EfBWPPmOADEdQIDAQABAoIBAQCx99j78qR05Szl
hqb4naPbtqBYzRj16KKsRmdo8QGLYlVgCoWKqogZueHEqnnV3VpD5V/vct0gWZ9A
Ynk14HxpsrZ1e0pWTnhm/xczlcx4J1O/YdXqNFE1xjHw9MnZiyo2DoetSqQUUvl2
wPPWh56tsOf9ldolcTe3yfZcC4RDGQBlQCzMAc1uLqscZavzhiD8c7LsSZF9kLZz
x8HV1TJoaATxoC6lXMzQrR+kML2hHby0wG1XpwoMVMC1lcQv10KGvhEz+YDEstpO
mxZZXCn7p6vzHwUCfa2EIMXVtHudOuxCxe8KmEQGtGrnJurIzXYKWPPR6FDqbTI6
g9JyyBoBAoGBAPH7lxYLXIMcXrssNYIHOgauihNlV3kKORHBofwRdsz2PP0XCiPs
M5+KJ9+w43T+S/DgedQXz8FQ/ExjpG9cvxaT/FO6GfSJrDcy6FBUXZkvAoTbLh9M
DWQ5E8FfAQ1UNDP5V/ESEYaxTQC3IxYJlz3VEwXfhmOOg+m1VZ1vaPP1AoGBANwy
L0OGctDv5CzySIjlgaY9NO9ez3Utsmm5vnHAeTL6g3DienzpRoyo0rJHNPL5DjUJ
gkW1dCS/rvHLnHzRvSQzDVLRCsJ6Cd7CQjTpsy1Kv6HMzZal3guD92GFOL9+b+o4
9U3p4vzjY5jlTgueqfrz2w2EsL2zuOcPEABYFr6BAoGBALIkjMLe1Fl1bkwPLMkv
9sjqf53t0mq6Wu82hMDkPnh/osCT0JRHlG2UMOyd9aWwfEm1iBra+MiRjVvTUz/k
oIzHn1AoRmlfXRg58wsIQOu/zvPtw9OokodA+ck23rhoUBIfM1229o4ZQt4O9NaJ
cv1DOsDtIKt0RKquI3xGg5ZtAoGALfkgWxXQFQVw+11efY6FYiL3UV7XK5zt2hsY
wwEvjNA27zOp5TiDLUz2KJirWmtbZwFkPI+k/yMyMHOVaY4U0mECUB8rAu2d7+9Z
CVkdusAXgH2VEvXwhTD5TlgVQA3y6dEYjjrd1HTZT4vYnp5y2N1fB9SDXigO29cO
PTQnE4ECgYEAz4uQISBWz2UgWk4zEKqvMyye2rCq+9PM5XbImPvIr3RgA3R6B5ls
D+0QCpxBbfzlasE9j36uVImWSDS99VoPryTuHj5g9eSZBDdo3Zi7S6vToKvF+wqg
a1rUTOpL6POzvLvd0G7GNUWx+UeFCpQUIfLgP/xWdF4VEBNnYQYBE2A=
-----END RSA PRIVATE KEY-----`

var _ = Describe("XFCC", func() {

	Describe("always_forward", func() {
		var (
			appName      string
			golangHeader string
		)

		BeforeEach(func() {
			golangHeader = asset.TcpSampleGolang
			appName = routing_helpers.GenerateAppName()
			routing_helpers.PushAppNoStart(appName, golangHeader, routingConfig.GoBuildpackName, routingConfig.AppsDomain, CF_PUSH_TIMEOUT, "256M")
			routing_helpers.EnableDiego(appName, DEFAULT_TIMEOUT)
			routing_helpers.StartApp(appName, DEFAULT_TIMEOUT)
		})

		AfterEach(func() {
			routing_helpers.AppReport(appName, 2*time.Minute)
			routing_helpers.DeleteApp(appName, time.Duration(2)*time.Minute)
		})

		Context("non-MTLS connection", func() {

			It("does not remove the xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.AlwaysForward {
					By(fmt.Sprintf("testing router %s configured to 'always_forward'", routerAddr))
					routerURI := fmt.Sprintf("http://%s:80/headers", routerAddr)
					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Header.Add("X-Forwarded-Client-Cert", "fake-xfcc")
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{}
					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()
					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).To(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).To(ContainSubstring("fake-xfcc"))
				}
			})

			It("does not add the xfcc header when xfcc header is not provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.AlwaysForward {
					By(fmt.Sprintf("testing router %s configured to 'always_forward'", routerAddr))
					routerURI := fmt.Sprintf("http://%s:80/headers", routerAddr)
					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{}
					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()
					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
				}
			})
		})

		Context("MTLS connection", func() {

			It("does not remove the xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.AlwaysForward {
					By(fmt.Sprintf("testing router %s configured to 'always_forward'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					keyPEM, certPEM := helpers.CreateKeyPair("testclient.com")
					clientCert, err := tls.X509KeyPair(certPEM, keyPEM)

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
					}
					t := &http.Transport{
						TLSClientConfig: tlsConfig,
					}

					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Header.Add("X-Forwarded-Client-Cert", "fake-xfcc")
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{
						Transport: t,
					}

					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()

					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).To(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).To(ContainSubstring("fake-xfcc"))
				}
			})

			It("does not add the xfcc header when xfcc header is not provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.AlwaysForward {
					By(fmt.Sprintf("testing router %s configured to 'always_forward'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					keyPEM, certPEM := helpers.CreateKeyPair("testclient.com")
					clientCert, err := tls.X509KeyPair(certPEM, keyPEM)

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
					}
					t := &http.Transport{
						TLSClientConfig: tlsConfig,
					}

					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{
						Transport: t,
					}

					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()

					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
				}
			})
		})

	})

	Describe("forward", func() {
		var (
			appName      string
			golangHeader string
		)

		BeforeEach(func() {
			golangHeader = asset.TcpSampleGolang
			appName = routing_helpers.GenerateAppName()
			routing_helpers.PushAppNoStart(appName, golangHeader, routingConfig.GoBuildpackName, routingConfig.AppsDomain, CF_PUSH_TIMEOUT, "256M")
			routing_helpers.EnableDiego(appName, DEFAULT_TIMEOUT)
			routing_helpers.StartApp(appName, DEFAULT_TIMEOUT)
		})

		AfterEach(func() {
			routing_helpers.AppReport(appName, 2*time.Minute)
			routing_helpers.DeleteApp(appName, time.Duration(2)*time.Minute)
		})

		Context("non-MTLS connection", func() {

			It("removes the xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.Forward {
					By(fmt.Sprintf("testing router %s configured to 'forward'", routerAddr))
					routerURI := fmt.Sprintf("http://%s:80/headers", routerAddr)
					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Header.Add("X-Forwarded-Client-Cert", "fake-xfcc")
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{}
					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()
					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).ToNot(ContainSubstring("fake-xfcc"))
				}
			})

			It("does not add the xfcc header when xfcc header is not provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.Forward {
					By(fmt.Sprintf("testing router %s configured to 'forward'", routerAddr))
					routerURI := fmt.Sprintf("http://%s:80/headers", routerAddr)
					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{}
					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()
					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
				}
			})
		})

		Context("MTLS connection", func() {

			It("forwards xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.Forward {
					By(fmt.Sprintf("testing router %s configured to 'forward'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					//keyPEM, certPEM := helpers.CreateKeyPair("testclient.com")
					clientCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
					}
					t := &http.Transport{
						TLSClientConfig: tlsConfig,
					}

					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Header.Add("X-Forwarded-Client-Cert", "fake-xfcc")
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{
						Transport: t,
					}

					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()

					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).To(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).To(ContainSubstring("fake-xfcc"))
				}
			})

			It("does not add the xfcc header when xfcc header is not provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.Forward {
					By(fmt.Sprintf("testing router %s configured to 'forward'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					keyPEM, certPEM := helpers.CreateKeyPair("testclient.com")
					clientCert, err := tls.X509KeyPair(certPEM, keyPEM)

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
					}
					t := &http.Transport{
						TLSClientConfig: tlsConfig,
					}

					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{
						Transport: t,
					}

					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()

					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
				}
			})
		})

	})
	Describe("sanitize_set", func() {
		var (
			appName      string
			golangHeader string
		)

		BeforeEach(func() {
			golangHeader = asset.TcpSampleGolang
			appName = routing_helpers.GenerateAppName()
			routing_helpers.PushAppNoStart(appName, golangHeader, routingConfig.GoBuildpackName, routingConfig.AppsDomain, CF_PUSH_TIMEOUT, "256M")
			routing_helpers.EnableDiego(appName, DEFAULT_TIMEOUT)
			routing_helpers.StartApp(appName, DEFAULT_TIMEOUT)
		})

		AfterEach(func() {
			routing_helpers.AppReport(appName, 2*time.Minute)
			routing_helpers.DeleteApp(appName, time.Duration(2)*time.Minute)
		})

		Context("non-MTLS connection", func() {

			It("removes the xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.SanitizeSet {
					By(fmt.Sprintf("testing router %s configured to 'sanitize_set'", routerAddr))
					routerURI := fmt.Sprintf("http://%s:80/headers", routerAddr)
					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Header.Add("X-Forwarded-Client-Cert", "fake-xfcc")
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{}
					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()
					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).ToNot(ContainSubstring("fake-xfcc"))
				}
			})

			It("does not add the xfcc header when xfcc header is not provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.SanitizeSet {
					By(fmt.Sprintf("testing router %s configured to 'sanitize_set'", routerAddr))
					routerURI := fmt.Sprintf("http://%s:80/headers", routerAddr)
					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{}
					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()
					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).ToNot(ContainSubstring("X-Forwarded-Client-Cert"))
				}
			})
		})

		Context("MTLS connection", func() {

			It("replaces the xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.SanitizeSet {
					By(fmt.Sprintf("testing router %s configured to 'sanitize_set'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					//keyPEM, certPEM := helpers.CreateKeyPair("testclient.com")
					clientCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
					}
					t := &http.Transport{
						TLSClientConfig: tlsConfig,
					}

					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Header.Add("X-Forwarded-Client-Cert", "fake-xfcc")
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{
						Transport: t,
					}

					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()

					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).To(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).ToNot(ContainSubstring("fake-xfcc"))
					Expect(string(resBytes)).To(ContainSubstring(sanitize([]byte(certPEM))))
				}
			})

			It("adds the xfcc header when xfcc header is not provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.SanitizeSet {
					By(fmt.Sprintf("testing router %s configured to 'sanitize_set'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					//keyPEM, certPEM := helpers.CreateKeyPair("testclient.com")
					clientCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
					}
					t := &http.Transport{
						TLSClientConfig: tlsConfig,
					}

					req, err := http.NewRequest("GET", routerURI, nil)
					Expect(err).NotTo(HaveOccurred())
					req.Host = fmt.Sprintf("%s.%s", appName, routingConfig.AppsDomain)
					client := http.Client{
						Transport: t,
					}

					res, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.Body).NotTo(BeNil())
					defer func() {
						_ = res.Body.Close()
					}()

					resBytes, err := ioutil.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resBytes)).To(ContainSubstring("X-Forwarded-Client-Cert"))
					Expect(string(resBytes)).To(ContainSubstring(sanitize([]byte(certPEM))))
				}
			})
		})

	})

})

func sanitize(cert []byte) string {
	s := string(cert)
	r := strings.NewReplacer("-----BEGIN CERTIFICATE-----", "",
		"-----END CERTIFICATE-----", "",
		"\n", "")
	s = r.Replace(s)
	return base64.StdEncoding.EncodeToString([]byte(s))
}
