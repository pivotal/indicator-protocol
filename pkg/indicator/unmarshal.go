package indicator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"github.com/ghodss/yaml"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReadOpt func(options *readOptions)

func DocumentFromYAML(r io.ReadCloser, opts ...ReadOpt) (v1.IndicatorDocument, error) {
	docBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return v1.IndicatorDocument{}, err
	}

	readOptions := getReadOpts(opts)
	if readOptions.interpolate {
		docBytes, err = interpolateBytes(docBytes, readOptions.overrides)
		if err != nil {
			return v1.IndicatorDocument{}, err
		}
	}


	apiVersion, err := ApiVersionFromYAML(docBytes)
	if err != nil {
		return v1.IndicatorDocument{}, err
	}

	var doc v1.IndicatorDocument
	switch apiVersion {
	case api_versions.V0:
		log.Print("WARNING: apiVersion v0 will be deprecated in future releases")
		doc, err = v0documentFromBytes(docBytes)
	case api_versions.V1:
		// Validate documentBytes according to the OpenAPI Schema
		errs, ok := v1.ValidateBytesBySchema(docBytes, "IndicatorDocument")
		if !ok {
			errorString := "Unable to validate document, errors were:\n"
			for _, err := range errs {
				errorString += " - " + err.Error() + "\n"
			}
			return v1.IndicatorDocument{}, errors.New(errorString)
		}
		err = yaml.Unmarshal(docBytes, &doc)
	default:
		err = fmt.Errorf("invalid apiVersion, supported versions are: [v0, indicatorprotocol.io/v1]")
	}

	if err != nil {
		return v1.IndicatorDocument{}, err
	}

	v1.PopulateDefaults(&doc)

	return doc, nil
}

// Assuming the given bytes are yaml, reads  the mapping under `metadata.labels` and interpolates that values
// wherever the keys are written.
func interpolateBytes(docBytes []byte, overrides map[string]string) ([]byte, error) {
	var metadataContainer struct {
		Metadata struct {
			Labels map[string]string
		}
	}
	err := yaml.Unmarshal(docBytes, &metadataContainer)
	if err != nil {
		return []byte{}, err
	}

	for key, value := range overrides {
		metadataContainer.Metadata.Labels[key] = value
	}

	for key, value := range metadataContainer.Metadata.Labels {
		regString := fmt.Sprintf(`(\$%s)(\b|\_|$)|(\$\{%s\})`, key, key)
		regex := regexp.MustCompile(regString)
		docBytes = regex.ReplaceAll(docBytes, []byte(fmt.Sprintf("%s$2", value)))
	}
	return docBytes, nil
}

func v0documentFromBytes(yamlBytes []byte) (v1.IndicatorDocument, error) {
	var d v0yamlDocument

	err := yaml.Unmarshal(yamlBytes, &d)
	if err != nil {
		return v1.IndicatorDocument{}, fmt.Errorf("could not unmarshal indicator document")
	}

	var indicators []v1.IndicatorSpec
	for indicatorIndex, yamlIndicator := range d.Indicators {
		var thresholds []v1.Threshold
		for thresholdIndex, yamlThreshold := range yamlIndicator.Thresholds {
			threshold, err := v0thresholdFromYAML(yamlThreshold)
			if err != nil {
				return v1.IndicatorDocument{}, fmt.Errorf("could not unmarshal threshold %v in indicator %v", thresholdIndex, indicatorIndex)
			}

			thresholds = append(thresholds, threshold)
		}

		p := v0presentationFromYAML(yamlIndicator.Presentation)

		indicators = append(indicators, v1.IndicatorSpec{
			Name:          yamlIndicator.Name,
			Type:          v1.DefaultIndicator,
			PromQL:        yamlIndicator.Promql,
			Thresholds:    thresholds,
			Alert:         v0alertFromYAML(yamlIndicator.Alert),
			Presentation:  p,
			Documentation: yamlIndicator.Documentation,
		})
	}

	layout := getLayout(d.YAMLLayout)

	return v1.IndicatorDocument{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api_versions.V1,
			Kind:       "IndicatorDocument",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: d.Metadata,
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    d.Product.Name,
				Version: d.Product.Version,
			},
			Indicators: indicators,
			Layout:     layout,
		},
	}, nil
}

func getLayout(l *v0yamlLayout) v1.Layout {
	if l == nil {
		return v1.Layout{}
	}
	sections := make([]v1.Section, 0)

	for _, s := range l.Sections {
		sections = append(sections, v1.Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.IndicatorRefs,
		})
	}

	return v1.Layout{
		Title:       l.Title,
		Description: l.Description,
		Owner:       l.Owner,
		Sections:    sections,
	}
}

func v0thresholdFromYAML(threshold v0yamlThreshold) (v1.Threshold, error) {
	var operator v1.ThresholdOperator
	var value float64
	var err error

	switch {
	case threshold.LT != "":
		operator = v1.LessThan
		value, err = strconv.ParseFloat(threshold.LT, 64)
	case threshold.LTE != "":
		operator = v1.LessThanOrEqualTo
		value, err = strconv.ParseFloat(threshold.LTE, 64)
	case threshold.EQ != "":
		operator = v1.EqualTo
		value, err = strconv.ParseFloat(threshold.EQ, 64)
	case threshold.NEQ != "":
		operator = v1.NotEqualTo
		value, err = strconv.ParseFloat(threshold.NEQ, 64)
	case threshold.GTE != "":
		operator = v1.GreaterThanOrEqualTo
		value, err = strconv.ParseFloat(threshold.GTE, 64)
	case threshold.GT != "":
		operator = v1.GreaterThan
		value, err = strconv.ParseFloat(threshold.GT, 64)
	default:
		operator = v1.UndefinedOperator
	}

	if err != nil {
		return v1.Threshold{}, err
	}

	return v1.Threshold{
		Level:    threshold.Level,
		Operator: operator,
		Value:    value,
	}, nil
}

func v0presentationFromYAML(p *v0yamlPresentation) v1.Presentation {
	if p == nil {
		return v1.Presentation{
			ChartType:    v1.StepChart,
			CurrentValue: false,
			Frequency:    0,
			Labels:       []string{},
			Units:        "",
		}
	}

	chartType := p.ChartType
	if chartType == "" {
		chartType = v1.StepChart
	}

	return v1.Presentation{
		ChartType:    chartType,
		CurrentValue: p.CurrentValue,
		Frequency:    p.Frequency,
		Labels:       p.Labels,
		Units:        p.Units,
	}
}

func v0alertFromYAML(a v0yamlAlert) v1.Alert {
	alertFor, alertStep := a.For, a.Step
	if alertFor == "" {
		alertFor = "1m"
	}
	if alertStep == "" {
		alertStep = "1m"
	}

	return v1.Alert{
		For:  alertFor,
		Step: alertStep,
	}
}

type v0yamlDocument struct {
	APIVersion string            `json:"apiVersion"`
	Product    v0yamlProduct     `json:"product"`
	Metadata   map[string]string `json:"metadata"`
	Indicators []v0yamlIndicator `json:"indicators"`
	YAMLLayout *v0yamlLayout     `json:"layout"`
}

type v0yamlProduct struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type v0yamlLayout struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Sections    []v0yamlSection `json:"sections"`
	Owner       string          `json:"owner"`
}

type v0yamlSection struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	IndicatorRefs []string `json:"indicators"`
}

type v0yamlIndicator struct {
	Name          string              `json:"name"`
	Promql        string              `json:"promql"`
	Thresholds    []v0yamlThreshold   `json:"thresholds"`
	Alert         v0yamlAlert         `json:"alert"`
	Documentation map[string]string   `json:"documentation"`
	Presentation  *v0yamlPresentation `json:"presentation"`
}

type v0yamlAlert struct {
	For  string
	Step string
}

type v0yamlThreshold struct {
	Level string `json:"level"`
	LT    string `json:"lt"`
	LTE   string `json:"lte"`
	EQ    string `json:"eq"`
	NEQ   string `json:"neq"`
	GTE   string `json:"gte"`
	GT    string `json:"gt"`
}

type v0yamlPresentation struct {
	ChartType    v1.ChartType `json:"chartType"`
	CurrentValue bool         `json:"currentValue"`
	Frequency    int64        `json:"frequency"`
	Labels       []string     `json:"labels"`
	Units        string       `json:"units"`
}

func ApiVersionFromYAML(docBytes []byte) (string, error) {
	var d struct {
		ApiVersion string `yaml:"apiVersion"`
	}
	err := yaml.Unmarshal(docBytes, &d)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal apiVersion")
	}
	return d.ApiVersion, nil
}

func KindFromYAML(fileBytes []byte) (string, error) {
	var f struct{ Kind string }

	err := yaml.Unmarshal(fileBytes, &f)
	if err != nil {
		return "", err
	}

	return f.Kind, nil
}

func PatchFromYAML(reader io.ReadCloser) (Patch, error) {
	var yamlPatch yamlPatch
	patchBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return Patch{}, fmt.Errorf("could not read patch: %s", err)
	}
	err = yaml.Unmarshal(patchBytes, &yamlPatch)
	if err != nil {
		return Patch{}, fmt.Errorf("could not unmarshal patch: %s", err)
	}
	_ = reader.Close()

	return Patch{
		APIVersion: yamlPatch.APIVersion,
		Match: Match{
			Name:     yamlPatch.Match.Product.Name,
			Version:  yamlPatch.Match.Product.Version,
			Metadata: yamlPatch.Match.Metadata,
		},
		Operations: yamlPatch.Operations,
	}, nil
}

func ProductFromYAML(reader io.ReadCloser) (v1.Product, error) {
	docBytes, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return v1.Product{}, fmt.Errorf("could not read document")
	}

	apiVersion, err := ApiVersionFromYAML(docBytes)
	var product v1.Product
	switch apiVersion {
	case api_versions.V0:
		var d struct {
			Product v1.Product
		}
		err = yaml.Unmarshal(docBytes, &d)
		product = d.Product
	case api_versions.V1:
		var d struct {
			Spec struct {
				Product v1.Product
			}
		}
		err = yaml.Unmarshal(docBytes, &d)
		product = d.Spec.Product
	}

	if err != nil {
		return v1.Product{}, errors.New("could not unmarshal product information")
	}

	return product, nil
}

func MetadataFromYAML(reader io.ReadCloser) (map[string]string, error) {
	docBytes, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read document")
	}

	apiVersion, err := ApiVersionFromYAML(docBytes)
	var metadata map[string]string
	switch apiVersion {
	case api_versions.V0:
		var d struct {
			Metadata map[string]string
		}
		err = yaml.Unmarshal(docBytes, &d)
		metadata = d.Metadata
	case api_versions.V1:
		var d struct {
			Metadata struct {
				Labels map[string]string
			}
		}
		err = yaml.Unmarshal(docBytes, &d)
		metadata = d.Metadata.Labels
	}

	if err != nil {
		return map[string]string{}, fmt.Errorf("could not unmarshal metadata")
	}
	_ = reader.Close()

	return metadata, nil
}

type yamlPatch struct {
	APIVersion string               `yaml:"apiVersion"`
	Match      yamlMatch            `yaml:"match"`
	Operations []patch.OpDefinition `yaml:"operations"`
}

type yamlMatch struct {
	Product struct {
		Name    *string `yaml:"name,omitempty"`
		Version *string `yaml:"version,omitempty"`
	} `yaml:"product,omitempty"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

func SkipMetadataInterpolation(options *readOptions) {
	options.interpolate = false
}

func OverrideMetadata(overrideMetadata map[string]string) func(options *readOptions) {
	return func(options *readOptions) {
		for k, v := range overrideMetadata {
			options.overrides[k] = v
		}
	}
}

func ProcessDocument(patches []Patch, documentBytes []byte) (v1.IndicatorDocument, []error) {
	patchedDocBytes, err := ApplyPatches(patches, documentBytes)
	if err != nil {
		log.Print("failed to apply patches to document")
		return v1.IndicatorDocument{}, []error{err}
	}

	reader := ioutil.NopCloser(bytes.NewReader(patchedDocBytes))
	doc, err := DocumentFromYAML(reader)
	if err != nil {
		log.Print("failed to unmarshal document")
		return v1.IndicatorDocument{}, []error{err}
	}

	doc.Interpolate()

	errs := doc.Validate(api_versions.V0, api_versions.V1)
	if len(errs) > 0 {
		log.Print("document validation failed")
		for _, e := range errs {
			log.Printf("- %s \n", e.Error())
		}
		return v1.IndicatorDocument{}, errs
	}

	return doc, nil
}

func ParseMetadata(input string) map[string]string {
	metadata := map[string]string{}

	for _, pair := range strings.Split(input, ",") {
		v := strings.Split(pair, "=")
		if len(v) > 1 {
			metadata[v[0]] = v[1]
		}
	}

	return metadata
}

type readOptions struct {
	interpolate bool
	overrides   map[string]string
}
