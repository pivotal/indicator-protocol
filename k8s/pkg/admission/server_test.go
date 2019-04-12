package admission_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/admission"
	"k8s.io/api/admission/v1beta1"
)

func TestServer(t *testing.T) {
	t.Run("it returns 200 for metrics endpoint without TLS", func(t *testing.T) {
		g := NewGomegaWithT(t)
		server := startServer(g)
		defer func() {
			_ = server.Close()
		}()

		resp, err := http.Get("http://" + server.Addr() + "/metrics")
		if err != nil {
			t.Fatal(err)
		}

		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("it returns 200 for metrics endpoint with TLS", func(t *testing.T) {
		g := NewGomegaWithT(t)
		cert, err := tls.X509KeyPair(FakeLocalhostCert, FakeLocalhostKey)
		if err != nil {
			log.Fatalf("Unable to load certs: %s", err)
		}
		tlsConf := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server := admission.NewServer("127.0.0.1:0", admission.WithTLSConfig(tlsConf))
		server.Run(false)
		defer func() {
			_ = server.Close()
		}()

		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		var resp *http.Response
		for i := 0; i < 100; i++ {
			resp, err = client.Get("https://" + server.Addr() + "/metrics")
			if err == nil {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if err != nil {
			t.Fatal(err)
		}

		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("it allows blocking on server.Run", func(t *testing.T) {
		g := NewGomegaWithT(t)

		server := admission.NewServer("127.0.0.1:0")

		done := make(chan struct{})
		go func() {
			defer close(done)
			server.Run(true)
		}()

		g.Consistently(done).ShouldNot(BeClosed())

		defer func() {
			_ = server.Close()
		}()
		var (
			err  error
			resp *http.Response
		)
		for i := 0; i < 100; i++ {
			resp, err = http.Get("http://" + server.Addr() + "/metrics")
			if err == nil {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if err != nil {
			t.Fatal(err)
		}

		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("it expects a content type of application/json", func(t *testing.T) {
		server := admission.NewServer("127.0.0.1:0")
		server.Run(false)
		defer func() {
			_ = server.Close()
		}()

		for _, endpoint := range []string{"indicatordocument", "indicator"} {
			t.Run(endpoint, func(t *testing.T) {
				g := NewGomegaWithT(t)
				startServer(g)
				var (
					err  error
					resp *http.Response
				)
				for i := 0; i < 100; i++ {
					resp, err = http.Post(
						fmt.Sprintf("http://%s/defaults/%s", server.Addr(), endpoint),
						"text/plain",
						strings.NewReader(`{}`),
					)
					if err == nil {
						break
					}
					time.Sleep(5 * time.Millisecond)
				}
				g.Expect(err).To(BeNil())

				body, err := ioutil.ReadAll(resp.Body)
				g.Expect(err).To(BeNil())
				g.Expect(string(body)).To(ContainSubstring("Expected a Content-Type of application/json"))
				g.Expect(resp.StatusCode).To(Equal(http.StatusUnsupportedMediaType))
			})
		}
	})
}

func TestValidators(t *testing.T) {

	t.Run("indicator document", func(t *testing.T) {
		t.Run("allows request when valid", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorDocumentRequest("CREATE", `{
							"product": {"name":"uaa", "version":"v1.2.3"},
							"indicators": [{
						    	"name": "latency",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": { "step" : "30m", "for": "5m" },
								"presentation": { 
									"chartType" : "step", 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							}],
							"layout": {
								"sections":[{
									"title": "Metrics",
									"indicators": ["latency"],
									"description": ""
								}]
							}
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/validation/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)

			g.Expect(actualResp.Response.Allowed).To(BeTrue())
		})
		t.Run("does not allow request when metadata contains a `step` key", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorDocumentRequest("CREATE", `{
							"product": {"name":"uaa", "version":"v1.2.3"},
							"apiVersion": "v0",
							"metadata": {"step": "12m"},
							"indicators": [{
						    	"name": "latency",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": { "step" : "30m", "for": "5m" },
								"presentation": { 
									"chartType" : "step", 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							}],
							"layout": {
								"sections":[{
									"title": "Metrics",
									"indicators": ["latency"],
									"description": ""
								}]
							}
						  }`, `{"step": "10m"}`)
			resp, err := http.Post(fmt.Sprintf("http://%s/validation/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)

			g.Expect(actualResp.Response.Allowed).To(BeFalse())
			g.Expect(actualResp.Response.Result.Message).To(ContainSubstring("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"))
		})
		t.Run("return UUID in patch response", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()
			reqBody := newIndicatorDocumentRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60"
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/validation/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			g.Expect(err).To(BeNil())
			g.Expect(actualResp.Response.UID).To(Equal(types.UID("f70772c9-572a-11e9-904e-42010a80018e")))
		})
	})

	t.Run("indicator", func(t *testing.T) {
		t.Run("allows request when valid", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := strings.NewReader(`
			{
			  "kind": "AdmissionReview",
			  "apiVersion": "admission.k8s.io/v1beta1",
			  "request": {
				"uid": "f70772c9-572a-11e9-904e-42010a80018d",
				"kind": {
				  "group": "apps.pivotal.io",
				  "version": "v1alpha1",
				  "kind": "Indicator"
				},
				"resource": {
				  "group": "apps.pivotal.io",
				  "version": "v1alpha1",
				  "resource": "indicators"
				},
				"namespace": "monitoring-indicator-protocol",
				"operation": "CREATE",
				"object": {
				  "apiVersion": "apps.pivotal.io/v1alpha1",
				  "kind": "Indicator",
				  "metadata": {
					"name": "test-indicator",
					"namespace": "monitoring-indicator-protocol"
				  },
				  "spec": {
				    "name": "latency",
				    "promql": "rate(apiserver_request_count[5m]) * 60",
				    "thresholds": [{
					  "gt": 375,
					  "level": "critical"
				    }],
					"presentation": {
						"currentValue": true,
						"frequency": 10,
						"labels": ["moo"],
						"chartType": "step"
					}
				  }
				},
				"oldObject": null
			  }
			}
		`)
			resp, err := http.Post(fmt.Sprintf("http://%s/validation/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			g.Expect(actualResp.Response.Allowed).To(BeTrue())
		})

		t.Run("does not allow request when missing threshold operator", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := strings.NewReader(`
			{
			  "kind": "AdmissionReview",
			  "apiVersion": "admission.k8s.io/v1beta1",
			  "request": {
				"uid": "f70772c9-572a-11e9-904e-42010a80018d",
				"kind": {
				  "group": "apps.pivotal.io",
				  "version": "v1alpha1",
				  "kind": "Indicator"
				},
				"resource": {
				  "group": "apps.pivotal.io",
				  "version": "v1alpha1",
				  "resource": "indicators"
				},
				"namespace": "monitoring-indicator-protocol",
				"operation": "CREATE",
				"object": {
				  "apiVersion": "apps.pivotal.io/v1alpha1",
				  "kind": "Indicator",
				  "metadata": {
					"name": "test-indicator",
					"namespace": "monitoring-indicator-protocol"
				  },
				  "spec": {
				    "name": "latency",
				    "promql": "rate(apiserver_request_count[5m]) * 60",
				    "thresholds": [{
					  "level": "critical"
				    }],
					"presentation": {
					  "chartType": "step"
					}
				  }
				},
				"oldObject": null
			  }
			}
		`)
			resp, err := http.Post(fmt.Sprintf("http://%s/validation/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			g.Expect(actualResp.Response.Allowed).To(BeFalse())
			g.Expect(actualResp.Response.Result.Message).To(ContainSubstring("one of [lt, lte, eq, neq, gte, gt] must be provided as a float"))
		})

		t.Run("return UUID in patch response", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60"
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/validation/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			g.Expect(err).To(BeNil())
			g.Expect(actualResp.Response.UID).To(Equal(types.UID("f70772c9-572a-11e9-904e-42010a80018d")))
		})
	})
}

func TestDefaultValues(t *testing.T) {

	t.Run("indicator", func(t *testing.T) {
		t.Run("patches default alert `for`", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "step" : "30m" },
							"presentation": { 
								"chartType" : "step", 
								"currentValue" : true,
								"frequency": 10,
								"labels": ["pod"]
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/alert/for","value":"1m"}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches default alert `step`", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("UPDATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "for" : "30m" },
							"presentation": { 
								"chartType" : "step", 
								"currentValue" : true,
								"frequency": 10,
								"labels": ["pod"]
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/alert/step","value":"1m"}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches default alert `for` and `step`", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"presentation": { 
								"chartType" : "step", 
								"currentValue" : true,
								"frequency": 10,
								"labels": ["pod"]
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/alert","value":{"for":"1m","step":"1m"}}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches default presentation", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "step" : "30m", "for": "5m" }
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/presentation","value":{ 
								"chartType" : "step", 
								"currentValue" : false,
								"frequency": 0,
								"labels": []
							}}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches default presentation `chartType`", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("UPDATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "step" : "30m", "for": "5m" },
							"presentation": { 
								"currentValue" : true,
								"frequency": 10,
								"labels": ["moo"]
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/presentation/chartType","value": "step"}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches default presentation `currentValue`, `frequency`, and `labels`", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "step" : "30m", "for": "5m" },
							"presentation": { 
								"chartType" : "bar"
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[
{"op":"add","path":"/spec/presentation/currentValue","value": false},
{"op":"add","path":"/spec/presentation/frequency","value": 0},
{"op":"add","path":"/spec/presentation/labels","value": []}
]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches both 'presentation' and 'alert'", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			reqBody := newIndicatorRequest("UPDATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "step" : "30m" },
							"presentation": { 
								"currentValue" : true,
								"frequency": 10,
								"labels": ["pod"]
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}
			patch := []byte(`[
{"op":"add","path":"/spec/alert/for","value": "1m"},
{"op":"add","path":"/spec/presentation/chartType","value": "step"}
]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))

		})

		t.Run("does not patch in noop case", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60",
							"alert": { "for" : "1m", "step" : "1m" },
							"presentation": { 
								"chartType" : "step",
								"currentValue" : true,
								"frequency": 10,
								"labels": ["pod"]
							}
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			g.Expect(err).To(BeNil())
			g.Expect(actualResp.Response.Patch).To(BeNil())
		})

		t.Run("return UUID in patch response", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()
			reqBody := newIndicatorRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60"
						  }`)
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicator", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			g.Expect(err).To(BeNil())
			g.Expect(actualResp.Response.UID).To(Equal(types.UID("f70772c9-572a-11e9-904e-42010a80018d")))
		})
	})

	t.Run("indicator document", func(t *testing.T) {
		t.Run("patches default layout", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorDocumentRequest("CREATE", `{
							"product": {"name":"uaa", "version":"v1.2.3"},
							"indicators": [{
						    	"name": "latency",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": { "step" : "30m", "for": "5m" },
								"presentation": { 
									"chartType" : "step", 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							}]
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/layout","value":{
				"owner": "",
				"title": "",
				"description": "",
				"sections":[{
					"title": "Metrics",
					"indicators": ["latency"],
					"description": ""
				}]
			}}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches indicator alert and layout", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorDocumentRequest("UPDATE", `{
							"product": {"name":"uaa", "version":"v1.2.3"},
							"indicators": [{
						    	"name": "latency",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": {"step": "1m"},
								"presentation": { 
									"chartType" : "step", 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							}]
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[{"op":"add","path":"/spec/layout","value":{
				"owner": "",
				"title": "",
				"description": "",
				"sections":[{
					"title": "Metrics",
					"indicators": ["latency"],
					"description": ""
				}]
			}},{"op":"add","path":"/spec/indicators/0/alert/for","value":"1m"}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("patches multiple indicators", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorDocumentRequest("UPDATE", `{
							"product": {"name":"uaa", "version":"v1.2.3"},
							"layout": {
								"owner": "",
								"title": "",
								"description": "",
								"sections":[{
									"title": "Metrics",
									"indicators": ["throughput", "latency"],
									"description": ""
								}]
							},
							"indicators": [{
						    	"name": "throughput",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": {"step": "10m", "for": "5m"},
								"presentation": { 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							},{
						    	"name": "latency",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": {"step": "1m"},
								"presentation": { 
									"chartType" : "step", 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							}]
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}

			patch := []byte(`[
{"op":"add","path":"/spec/indicators/0/presentation/chartType","value":"step"},
{"op":"add","path":"/spec/indicators/1/alert/for","value":"1m"}]`)
			g.Expect(actualResp.Response.Patch).NotTo(BeNil())
			g.Expect(actualResp.Response.Patch).To(MatchJSON(patch))
		})

		t.Run("does not patch noop", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()

			//v1beta1.AdmissionReview{}
			reqBody := newIndicatorDocumentRequest("UPDATE", `{
							"product": {"name":"uaa", "version":"v1.2.3"},
							"indicators": [{
						    	"name": "latency",
						    	"promql": "rate(apiserver_request_count[5m]) * 60",
								"alert": { "step" : "30m", "for": "5m" },
								"presentation": { 
									"chartType" : "step", 
									"currentValue" : true,
									"frequency": 10,
									"labels": ["pod"]
								}
							}],
							"layout": {
								"owner": "Foo",
								"title": "Bar",
								"description": "why not",
								"sections":[{
									"title": "Metrics",
									"indicators": ["latency"],
									"description": "again"
								}]
							}
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			if err != nil {
				t.Errorf("unable to decode resp body: %s", err)
			}
			g.Expect(actualResp.Response.Patch).To(BeNil())
		})

		t.Run("return UUID in patch response", func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := startServer(g)
			defer func() {
				_ = server.Close()
			}()
			reqBody := newIndicatorDocumentRequest("CREATE", `{
						    "name": "latency",
						    "promql": "rate(apiserver_request_count[5m]) * 60"
						  }`, "{}")
			resp, err := http.Post(fmt.Sprintf("http://%s/defaults/indicatordocument", server.Addr()), "application/json", reqBody)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(200))

			var actualResp v1beta1.AdmissionReview
			err = json.NewDecoder(resp.Body).Decode(&actualResp)
			g.Expect(err).To(BeNil())
			g.Expect(actualResp.Response.UID).To(Equal(types.UID("f70772c9-572a-11e9-904e-42010a80018e")))
		})
	})
}

func startServer(g *GomegaWithT) *admission.Server {
	server := admission.NewServer("127.0.0.1:0")
	server.Run(false)
	var err error
	for i := 0; i < 100; i++ {
		_, err = http.Get(fmt.Sprintf("http://%s/metrics", server.Addr()))
		if err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	g.Expect(err).To(BeNil())
	return server
}

func newIndicatorRequest(operation string, indicatorSpec string) *strings.Reader {
	return strings.NewReader(fmt.Sprintf(`
					{
					  "kind": "AdmissionReview",
					  "apiVersion": "admission.k8s.io/v1beta1",
					  "request": {
						"uid": "f70772c9-572a-11e9-904e-42010a80018d",
						"kind": {
						  "group": "apps.pivotal.io",
						  "version": "v1alpha1",
						  "kind": "Indicator"
						},
						"resource": {
						  "group": "apps.pivotal.io",
						  "version": "v1alpha1",
						  "resource": "indicators"
						},
						"namespace": "monitoring-indicator-protocol",
						"operation": "%s",
						"object": {
						  "apiVersion": "apps.pivotal.io/v1alpha1",
						  "kind": "Indicator",
						  "metadata": {
							"name": "test-indicator",
							"namespace": "monitoring-indicator-protocol"
						  },
						  "spec": %s
						},
						"oldObject": null
					  }
					}
				`, operation, indicatorSpec))
}

func newIndicatorDocumentRequest(operation string, indicatorDocumentSpec string, metadata string) *strings.Reader {
	return strings.NewReader(fmt.Sprintf(`
						{
						  "kind": "AdmissionReview",
						  "apiVersion": "admission.k8s.io/v1beta1",
						  "request": {
							"uid": "f70772c9-572a-11e9-904e-42010a80018e",
							"kind": {
							  "group": "apps.pivotal.io",
							  "version": "v1alpha1",
							  "kind": "IndicatorDocument"
							},
							"resource": {
							  "group": "apps.pivotal.io",
							  "version": "v1alpha1",
							  "resource": "indicatordocuments"
							},
							"namespace": "monitoring-indicator-protocol",
							"operation": "%s",
							"object": {
							  "apiVersion": "apps.pivotal.io/v1alpha1",
							  "kind": "IndicatorDocument",
							  "metadata": {
								"labels": %s
							  },
							  "spec": %s
							},
							"oldObject": null
						  }
						}
					`, operation, metadata, indicatorDocumentSpec))
}