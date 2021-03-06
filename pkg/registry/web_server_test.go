package registry_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
)

func TestServingMetrics(t *testing.T) {
	g := NewGomegaWithT(t)
	addr, stop := newWebServer(19876)
	defer stop()

	var resp *http.Response
	f := func() error {
		var err error
		resp, err = http.Get(fmt.Sprintf("http://%s/metrics", addr))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("invalid status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
	g.Eventually(f).ShouldNot(HaveOccurred())
}

func TestRegisterAndServeDocuments(t *testing.T) {
	g := NewGomegaWithT(t)
	addr, stop := newWebServer(32897)
	defer stop()

	var resp *http.Response
	f := func() error {
		var err error
		file, err := os.Open("test_fixtures/doc.yml")
		if err != nil {
			return err
		}
		resp, err = http.Post(
			fmt.Sprintf("http://%s/v1/register", addr),
			"application/yml",
			file,
		)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("invalid status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}

	g.Eventually(f).ShouldNot(HaveOccurred())

	resp, err := http.Get(fmt.Sprintf("http://%s/v1/indicator-documents", addr))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, err := ioutil.ReadAll(resp.Body)
	g.Expect(err).ToNot(HaveOccurred())
	expectedJSON, err := ioutil.ReadFile("test_fixtures/example_response4.json")
	g.Expect(body).To(MatchJSON(expectedJSON))
}

func TestWritingAndReadingStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	addr, stop := newWebServer(32422)
	defer stop()
	var resp *http.Response
	f := func() error {
		var err error
		file, err := os.Open("test_fixtures/doc.yml")
		if err != nil {
			return err
		}
		resp, err = http.Post(
			fmt.Sprintf("http://%s/v1/register", addr),
			"application/yml",
			file,
		)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("invalid status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
	g.Eventually(f).ShouldNot(HaveOccurred())

	// make our status update request
	const documentUID = `my-product-a-a902332065d69c1787f419e235a1f1843d98c884`
	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/indicator-documents/%s/bulk_status", addr, documentUID),
		"application/json",
		bytes.NewReader([]byte(statusRequest)),
	)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	// get document
	resp, err = http.Get(fmt.Sprintf("http://%s/v1/indicator-documents", addr))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, err := ioutil.ReadAll(resp.Body)
	g.Expect(err).ToNot(HaveOccurred())

	var documents []registry.APIDocumentResponse
	err = json.Unmarshal(body, &documents)
	g.Expect(err).ToNot(HaveOccurred())

	var (
		indie1Status *string
		indie2Status *string
	)
	for _, doc := range documents {
		if doc.UID == documentUID {
			for _, ind := range doc.Spec.Indicators {
				switch ind.Name {
				case "indie1":
					indie1Status = ind.Status.Value
				case "indie2":
					indie2Status = ind.Status.Value
				}
			}
		}
	}

	g.Expect(indie1Status).To(Equal(strPtr("critical")))
	g.Expect(indie2Status).To(Equal(strPtr("warning")))
}

const (
	statusRequest = `[
			{
				"name": "indie1",
				"status": "critical"
			},
			{
				"name": "indie2",
				"status": "warning"
			}
		]`
)

func TestRoutesAllExist(t *testing.T) {
	g := NewGomegaWithT(t)
	addr, stop := newWebServer(34555)
	defer stop()

	routes := []string{
		"http://%s/v1/indicator-documents",
		"http://%s/v1/indicator-documents/",
		"http://%s/v1/register",
		"http://%s/v1/register/",
	}
	for _, route := range routes {
		completedRoute := fmt.Sprintf(route, addr)
		g.Eventually(func() (int, error) {
			resp, err := http.Get(completedRoute)
			if err != nil {
				return -1, err
			}
			return resp.StatusCode, err
		}).ShouldNot(Or(Equal(http.StatusNotFound), Equal(-1)),
			fmt.Sprintf("Could not reach route %s", completedRoute))

	}
}

func newWebServer(port int) (string, func() error) {
	conf := registry.WebServerConfig{
		Address:       "localhost:" + strconv.Itoa(port),
		DocumentStore: registry.NewDocumentStore(time.Second, time.Now),
		StatusStore:   status_store.New(time.Now),
	}

	start, stop := registry.NewWebServer(conf)
	go func() {
		err := start()
		if err != nil {
			log.Printf("web server failed: %s", err)
		}
	}()

	return conf.Address, stop
}
