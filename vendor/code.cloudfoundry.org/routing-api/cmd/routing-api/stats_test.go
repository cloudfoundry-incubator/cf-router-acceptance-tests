package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"time"

	"code.cloudfoundry.org/routing-api/cmd/routing-api/testrunner"
	"code.cloudfoundry.org/routing-api/models"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes API", func() {
	var (
		err               error
		route1            models.Route
		addr              *net.UDPAddr
		fakeStatsdServer  *net.UDPConn
		fakeStatsdChan    chan string
		routingAPIProcess ifrit.Process
	)

	BeforeEach(func() {
		routingAPIRunner := testrunner.New(routingAPIBinPath, routingAPIArgs)
		routingAPIProcess = ginkgomon.Invoke(routingAPIRunner)
		addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", 8125+GinkgoParallelNode()))
		Expect(err).ToNot(HaveOccurred())

		fakeStatsdServer, err = net.ListenUDP("udp", addr)
		Expect(err).ToNot(HaveOccurred())
		fakeStatsdServer.SetReadDeadline(time.Now().Add(15 * time.Second))
		fakeStatsdChan = make(chan string, 1)

		go func(statsChan chan string) {
			defer GinkgoRecover()
			for {
				buffer := make([]byte, 1000)
				_, err := fakeStatsdServer.Read(buffer)
				if err != nil {
					close(statsChan)
					return
				}
				scanner := bufio.NewScanner(bytes.NewBuffer(buffer))
				for scanner.Scan() {
					select {
					case statsChan <- scanner.Text():
					}
				}
			}
		}(fakeStatsdChan)

		time.Sleep(1000 * time.Millisecond)
	})

	AfterEach(func() {
		ginkgomon.Kill(routingAPIProcess)
		err := fakeStatsdServer.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Stats for event subscribers", func() {
		Context("Subscribe", func() {
			It("should increase subscriptions by 4", func() {

				eventStream1, err := client.SubscribeToEvents()
				Expect(err).NotTo(HaveOccurred())
				defer eventStream1.Close()

				eventStream2, err := client.SubscribeToEvents()
				Expect(err).NotTo(HaveOccurred())
				defer eventStream2.Close()

				eventStream3, err := client.SubscribeToEvents()
				Expect(err).NotTo(HaveOccurred())
				defer eventStream3.Close()

				eventStream4, err := client.SubscribeToEvents()
				Expect(err).NotTo(HaveOccurred())
				defer eventStream4.Close()

				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_subscriptions:+1|g")))
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_subscriptions:+1|g")))
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_subscriptions:+1|g")))
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_subscriptions:+1|g")))
			})
		})
	})

	Describe("Stats for total routes", func() {

		BeforeEach(func() {
			route1 = models.NewRoute("a.b.c", 33, "1.1.1.1", "potato", "", 55)
		})

		Context("periodically receives total routes", func() {
			It("Gets statsd messages for existing routes", func() {
				//The first time is because we get the event of adding the self route
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:1|g")))
				//Do it again to make sure it's not because of events
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:1|g")))
			})
		})

		Context("when creating and updating a new route", func() {
			It("Gets statsd messages for new routes", func() {
				client.UpsertRoutes([]models.Route{route1})

				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:+1|g")))
			})
		})

		Context("when deleting a route", func() {
			It("gets statsd messages for deleted routes", func() {
				client.UpsertRoutes([]models.Route{route1})

				client.DeleteRoutes([]models.Route{route1})

				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:+1|g")))
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:-1|g")))
			})
		})

		Context("when expiring a route", func() {
			It("gets statsd messages for expired routes", func() {
				routeExpire := models.NewRoute("z.a.k", 63, "42.42.42.42", "Tomato", "", 1)

				client.UpsertRoutes([]models.Route{routeExpire})

				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:+1|g")))
				Eventually(fakeStatsdChan).Should(Receive(Equal("routing_api.total_http_routes:-1|g")))
			})
		})
	})
})
