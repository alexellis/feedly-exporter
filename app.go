package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpClient http.Client

//FeedResponse FeedResponse
type FeedResponse struct {
	Subscribers  int64   `json:"subscribers"`
	Velocity     float64 `json:"velocity"`
	ResponseTime float64
}

type executorCollector struct {
	URLs            []string
	subscribers     *prometheus.Desc
	velocity        *prometheus.Desc
	feedly_response *prometheus.Desc
}

func main() {
	var feedlyUrlsRaw string
	var feedlyURLs []string

	flag.StringVar(&feedlyUrlsRaw, "urls", "", "feedly subscription URL")

	flag.Parse()

	if len(feedlyUrlsRaw) == 0 {
		fmt.Println("The -urls flag is required. It can contain commas.")
		return
	}

	if strings.Contains(feedlyUrlsRaw, ",") {
		feedlyURLs = strings.Split(feedlyUrlsRaw, ",")
	} else {
		feedlyURLs = []string{
			feedlyUrlsRaw,
		}
	}

	collector := newExecutorCollector(feedlyURLs)
	prometheus.Register(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func getHosts(jenkinsHost string) []string {
	var hosts []string
	parts := strings.Split(jenkinsHost, ",")
	for _, part := range parts {
		hosts = append(hosts, strings.Trim(part, " "))
	}
	return hosts
}

// NewExecutorCollector creates new executorCollector
func newExecutorCollector(feedlyUrls []string) *executorCollector {
	c := executorCollector{
		URLs: feedlyUrls,
	}

	c.subscribers = prometheus.NewDesc("feedly_subscribers", "count of subscriptions", []string{"url"}, prometheus.Labels{})
	c.feedly_response = prometheus.NewDesc("feedly_response", "remote response duration", []string{"url"}, prometheus.Labels{})
	c.velocity = prometheus.NewDesc("feedly_velocity", "velocity of publishing", []string{"url"}, prometheus.Labels{})
	return &c
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (c *executorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.subscribers
	ch <- c.velocity
	ch <- c.feedly_response
}

// Collect is called by the Prometheus registry when collecting
// metrics.
func (c *executorCollector) Collect(ch chan<- prometheus.Metric) {

	// Note: there is no throttling here, so we will be opening N file descriptors.
	grp := sync.WaitGroup{}
	for _, URL := range c.URLs {
		grp.Add(1)
		go func(URL string) {
			result, err := getFeedResponse(URL)

			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("[%fs] %s\n", result.ResponseTime, URL)
			ch <- prometheus.MustNewConstMetric(c.subscribers, prometheus.GaugeValue, float64(result.Subscribers), URL)
			ch <- prometheus.MustNewConstMetric(c.velocity, prometheus.GaugeValue, float64(result.Velocity), URL)
			ch <- prometheus.MustNewConstMetric(c.feedly_response, prometheus.GaugeValue, float64(result.ResponseTime), URL)

			grp.Done()
		}(URL)
	}
	grp.Wait()
}

func getFeedResponse(URL string) (FeedResponse, error) {
	var response FeedResponse
	var err error

	start := time.Now()

	URLAddr := url.QueryEscape(fmt.Sprintf("feed/%s", URL))

	req, err := http.NewRequest("GET", fmt.Sprintf("https://feedly.com/v3/feeds/%s", URLAddr), nil)

	res, err := httpClient.Do(req)
	if err != nil {
		return response, err
	}

	if res.Body != nil {
		defer res.Body.Close()
		resBody, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			err = readErr
		} else {
			err = json.Unmarshal(resBody, &response)
			if err != nil {
				fmt.Println(string(resBody))
			}
		}
	}

	response.ResponseTime = time.Since(start).Seconds()

	return response, err
}
