package admission

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	indicatorDocumentDefaultsReviewRequested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_document_defaults_requested",
		Help: "The number of times the /defaults/indicatordocument handler was hit.",
	})
	indicatorDocumentDefaultsReviewErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_document_defaults_errored",
		Help: "The number of times the /defaults/indicatordocument handler was hit and errored.",
	})
	indicatorDefaultsReviewRequested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_defaults_requested",
		Help: "The number of times the /defaults/indicator handler was hit.",
	})
	indicatorDefaultsReviewErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_defaults_errored",
		Help: "The number of times the /defaults/indicator handler was hit and errored.",
	})
	indicatorDocumentValidationReviewRequested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_document_validation_requested",
		Help: "The number of times the /validation/indicatordocument handler was hit.",
	})
	indicatorDocumentValidationReviewErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_document_validation_errored",
		Help: "The number of times the /validation/indicatordocument handler was hit and errored.",
	})
	indicatorValidationReviewRequested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_validation_requested",
		Help: "The number of times the /validation/indicator handler was hit.",
	})
	indicatorValidationReviewErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "indicator_validation_errored",
		Help: "The number of times the /validation/indicator handler was hit and errored.",
	})
)

type ServerOpt func(*Server)

type Server struct {
	mu  sync.Mutex
	lis net.Listener
	srv *http.Server

	addr      string
	tlsConfig *tls.Config
}

func NewServer(addr string, options ...ServerOpt) *Server {
	s := &Server{
		addr: addr,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

func WithTLSConfig(tlsConfig *tls.Config) ServerOpt {
	return func(s *Server) {
		s.tlsConfig = tlsConfig
	}
}

func (s *Server) Run(blocking bool) {
	if blocking {
		s.run()
		return
	}
	go s.run()
}

func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	return s.lis.Close()
}

func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lis == nil {
		return ""
	}

	return s.lis.Addr().String()
}

func (s *Server) run() {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("Unable to start listener: %s", err)

	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/defaults/indicatordocument", indicatorDocumentDefaultHandler)
	mux.HandleFunc("/defaults/indicator", indicatorDefaultHandler)
	mux.HandleFunc("/validation/indicatordocument", indicatorDocumentValidationHandler)
	mux.HandleFunc("/validation/indicator", indicatorValidationHandler)

	s.mu.Lock()
	s.lis = lis
	s.srv = &http.Server{
		TLSConfig: s.tlsConfig,
		Handler:   mux,
	}
	s.mu.Unlock()

	if s.tlsConfig != nil {
		err = s.srv.ServeTLS(lis, "", "")
	} else {
		err = s.srv.Serve(lis)
	}

	if err != nil {
		log.Printf("Server shutdown: %s", err)
	}
}

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func indicatorDocumentDefaultHandler(responseWriter http.ResponseWriter, r *http.Request) {
	indicatorDocumentDefaultsReviewRequested.Inc()

	requestedAdmissionReview, httpErr := deserializeReview(r)
	if httpErr != nil {
		indicatorDocumentDefaultsReviewErrored.Inc()
		log.Printf("Error deserializing review: %s", httpErr.message)
		httpErr.Write(responseWriter)
		return
	}

	var doc v1alpha1.IndicatorDocument
	err := json.Unmarshal(requestedAdmissionReview.Request.Object.Raw, &doc)
	if err != nil {
		indicatorDocumentDefaultsReviewErrored.Inc()
		log.Printf("Error unmarshaling document: %s", err)
		errUnableToDeserialize.Write(responseWriter)
		return
	}

	patchOperations := getLayoutPatches(doc)
	for idx, i := range doc.Spec.Indicators {
		patchOperations = append(patchOperations,
			getAlertPatches(i.Alert, fmt.Sprintf("/spec/indicators/%d", idx))...)
		patchOperations = append(patchOperations,
			getPresentationPatches(i.Presentation, fmt.Sprintf("/spec/indicators/%d", idx))...)
	}

	patchBytes, err := marshalPatches(patchOperations)
	if err != nil {
		indicatorDocumentDefaultsReviewErrored.Inc()
		log.Printf("Error marshaling patches: %s", err)
		errInternal.Write(responseWriter)
		return
	}

	data, err := json.Marshal(&v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     requestedAdmissionReview.Request.UID,
			Allowed: true,
			Patch:   patchBytes,
		},
	})
	if err != nil {
		indicatorDocumentDefaultsReviewErrored.Inc()
		log.Printf("Error marshaling resp: %s", err)
		errInternal.Write(responseWriter)
		return
	}
	_, err = responseWriter.Write(data)
	if err != nil {
		indicatorDocumentDefaultsReviewErrored.Inc()
		log.Printf("Error writing resp: %s", err)
	}
}

func getLayoutPatches(doc v1alpha1.IndicatorDocument) []patch {
	var patchOperations []patch
	if (reflect.DeepEqual(doc.Spec.Layout, v1alpha1.Layout{})) {
		var names []string
		for _, i := range doc.Spec.Indicators {
			names = append(names, i.Name)
		}

		patchOperations = append(patchOperations, patch{
			Op:   "add",
			Path: "/spec/layout",
			Value: v1alpha1.Layout{
				Sections: []v1alpha1.Section{{
					Title:      "Metrics",
					Indicators: names,
				}},
			},
		})
	}
	return patchOperations
}

func indicatorDefaultHandler(responseWriter http.ResponseWriter, request *http.Request) {
	indicatorDefaultsReviewRequested.Inc()

	requestedAdmissionReview, httpErr := deserializeReview(request)
	if httpErr != nil {
		indicatorDefaultsReviewErrored.Inc()
		log.Printf("Error deserializing review: %s", httpErr.message)
		httpErr.Write(responseWriter)
		return
	}

	var k8sIndicator v1alpha1.Indicator
	err := json.Unmarshal(requestedAdmissionReview.Request.Object.Raw, &k8sIndicator)
	if err != nil {
		indicatorDefaultsReviewErrored.Inc()
		log.Printf("Error unmarshalling indicator: %s", err)
		errUnableToDeserialize.Write(responseWriter)
		return
	}

	var patchOperations []patch
	patchOperations = append(patchOperations, getAlertPatches(k8sIndicator.Spec.Alert, "/spec")...)
	patchOperations = append(patchOperations, getPresentationPatches(k8sIndicator.Spec.Presentation, "/spec")...)

	patchBytes, err := marshalPatches(patchOperations)
	if err != nil {
		indicatorDefaultsReviewErrored.Inc()
		log.Printf("Error marshaling patches: %s", err)
		errInternal.Write(responseWriter)
		return
	}

	data, err := json.Marshal(&v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     requestedAdmissionReview.Request.UID,
			Patch:   patchBytes,
			Allowed: true,
		},
	})
	if err != nil {
		indicatorDefaultsReviewErrored.Inc()
		log.Printf("Unable to marshal resp: %s", err)
		errInternal.Write(responseWriter)
		return
	}
	_, err = responseWriter.Write(data)
	if err != nil {
		indicatorDefaultsReviewErrored.Inc()
		log.Printf("Unable to write resp: %s", err)
	}
}

func indicatorDocumentValidationHandler(responseWriter http.ResponseWriter, request *http.Request) {
	indicatorDocumentValidationReviewRequested.Inc()
	requestedAdmissionReview, httpErr := deserializeReview(request)
	if httpErr != nil {
		indicatorDocumentValidationReviewErrored.Inc()
		log.Printf("Error deserializing review: %s", httpErr.message)
		httpErr.Write(responseWriter)
		return
	}

	var k8sIndicatorDoc v1alpha1.IndicatorDocument
	err := json.Unmarshal(requestedAdmissionReview.Request.Object.Raw, &k8sIndicatorDoc)
	if err != nil {
		indicatorDocumentValidationReviewErrored.Inc()
		log.Printf("Error unmarshalling indicator document: %s", err)
		errUnableToDeserialize.Write(responseWriter)
		return
	}

	id := domain.Map(&k8sIndicatorDoc)
	errors := indicator.ValidateForK8s(id)

	auditAnnotationMessage := createReviewAnnotationMap(errors)

	data, err := json.Marshal(&v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:              requestedAdmissionReview.Request.UID,
			Allowed:          len(errors) == 0,
			Result:  &metav1.Status{
				Message: auditAnnotationMessage,
			},
		},
	})
	if err != nil {
		indicatorDocumentValidationReviewErrored.Inc()
		log.Printf("Unable to marshal resp: %s", err)
		errInternal.Write(responseWriter)
		return
	}
	_, err = responseWriter.Write(data)
	if err != nil {
		indicatorDocumentValidationReviewErrored.Inc()
		log.Printf("Unable to write resp: %s", err)
	}
}

func indicatorValidationHandler(responseWriter http.ResponseWriter, request *http.Request) {
	indicatorValidationReviewRequested.Inc()

	requestedAdmissionReview, httpErr := deserializeReview(request)
	if httpErr != nil {
		indicatorValidationReviewErrored.Inc()
		log.Printf("Error deserializing review: %s", httpErr.message)
		httpErr.Write(responseWriter)
		return
	}

	var k8sIndicator v1alpha1.Indicator
	err := json.Unmarshal(requestedAdmissionReview.Request.Object.Raw, &k8sIndicator)
	if err != nil {
		indicatorValidationReviewErrored.Inc()
		log.Printf("Error unmarshalling indicator: %s", err)
		errUnableToDeserialize.Write(responseWriter)
		return
	}

	i := domain.ToDomainIndicator(k8sIndicator.Spec)
	errors := indicator.ValidateIndicator(i, 0)

	auditAnnotationMessage := createReviewAnnotationMap(errors)

	data, err := json.Marshal(&v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     requestedAdmissionReview.Request.UID,
			Allowed: len(errors) == 0,
			Result:  &metav1.Status{
				Message: auditAnnotationMessage,
			},
		},
	})
	if err != nil {
		indicatorValidationReviewErrored.Inc()
		log.Printf("Unable to marshal resp: %s", err)
		errInternal.Write(responseWriter)
		return
	}
	_, err = responseWriter.Write(data)
	if err != nil {
		indicatorValidationReviewErrored.Inc()
		log.Printf("Unable to write resp: %s", err)
	}
}

func createReviewAnnotationMap(errors []error)string {
	errorsString := ""
	for _, errorString := range errors {
		errorsString += errorString.Error() + "\n"
	}
	return errorsString
}

func marshalPatches(patchOperations []patch) ([]byte, error) {
	var err error
	var patchBytes []byte
	if len(patchOperations) > 0 {
		patchBytes, err = json.Marshal(patchOperations)
	}
	return patchBytes, err
}

func getPresentationPatches(presentation v1alpha1.Presentation, context string) []patch {
	if reflect.DeepEqual(presentation, v1alpha1.Presentation{}) {
		return []patch{
			{
				Op:   "add",
				Path: fmt.Sprintf("%s/presentation", context),
				Value: v1alpha1.Presentation{
					ChartType:    "step",
					CurrentValue: false,
					Frequency:    0,
					Labels:       []string{},
				},
			},
		}
	}

	var patchOperations []patch
	if presentation.ChartType == "" {
		patchOperations = append(patchOperations, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/presentation/chartType", context),
			Value: indicator.StepChart,
		})
	}
	if !presentation.CurrentValue {
		patchOperations = append(patchOperations, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/presentation/currentValue", context),
			Value: false,
		})
	}
	if presentation.Frequency == 0 {
		patchOperations = append(patchOperations, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/presentation/frequency", context),
			Value: 0,
		})
	}
	if len(presentation.Labels) == 0 {
		patchOperations = append(patchOperations, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/presentation/labels", context),
			Value: []string{},
		})
	}
	return patchOperations
}

func getAlertPatches(alert v1alpha1.Alert, context string) []patch {
	if alert.For == "" && alert.Step == "" {
		return []patch{{
			Op:   "add",
			Path: fmt.Sprintf("%s/alert", context),
			Value: v1alpha1.Alert{
				For:  "1m",
				Step: "1m",
			},
		}}
	}
	var patchOperations []patch
	if alert.For == "" {
		patchOperations = append(patchOperations, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/alert/for", context),
			Value: "1m",
		})
	}
	if alert.Step == "" {
		patchOperations = append(patchOperations, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/alert/step", context),
			Value: "1m",
		})
	}
	return patchOperations
}

func deserializeReview(r *http.Request) (*v1beta1.AdmissionReview, *httpError) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, errUnsupportedMedia
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errUnableToReadBody
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Error closing body: %v\n", err)
		}
	}()

	var requestedAdmissionReview v1beta1.AdmissionReview

	deserializer := codecs.UniversalDeserializer()
	_, _, err = deserializer.Decode(data, nil, &requestedAdmissionReview)
	if err != nil {
		return nil, errUnableToDeserialize
	}

	if !validRequest(requestedAdmissionReview) {
		return nil, errInvalidRequest
	}

	return &requestedAdmissionReview, nil
}

func validRequest(r v1beta1.AdmissionReview) bool {
	return r.Request != nil
}