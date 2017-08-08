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

var _ = FDescribe("XFCC", func() {

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

					certPEM := routingConfig.Xfcc.CertPEM
					keyPEM := routingConfig.Xfcc.KeyPEM
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
	FDescribe("sanitize_set", func() {
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

			FIt("replaces the xfcc header when xfcc header is provided", func() {
				for _, routerAddr := range routingConfig.Xfcc.SanitizeSet {
					By(fmt.Sprintf("testing router %s configured to 'sanitize_set'", routerAddr))
					routerURI := fmt.Sprintf("https://%s:443/headers", routerAddr)

					certPEM := routingConfig.Xfcc.CertPEM
					keyPEM := routingConfig.Xfcc.KeyPEM
					clientCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
					fmt.Printf("using some cert %v", clientCert)

					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{clientCert},
						ClientAuth:         tls.RequestClientCert,
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

					certPEM := routingConfig.Xfcc.CertPEM
					keyPEM := routingConfig.Xfcc.KeyPEM
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
