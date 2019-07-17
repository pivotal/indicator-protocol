package prometheus_alerts

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

type Rule struct {
	Alert       string
	Expr        string
	For         string
	Labels      map[string]string
	Annotations map[string]string
}

type Document struct {
	Groups []Group
}

type Group struct {
	Name  string
	Rules []Rule
}

func AlertDocumentFilename(documentBytes []byte, productName string) string {
	return fmt.Sprintf("%s_%x.yml", productName, sha1.Sum(documentBytes))
}

func AlertDocumentFrom(document v1alpha1.IndicatorDocument) Document {
	rules := make([]Rule, 0)

	for _, ind := range document.Spec.Indicators {
		for _, threshold := range ind.Thresholds {
			rules = append(rules, ruleFrom(document, ind, threshold))
		}
	}

	return Document{
		Groups: []Group{{
			Name:  document.Spec.Product.Name,
			Rules: rules,
		}},
	}
}

func ruleFrom(document v1alpha1.IndicatorDocument, i v1alpha1.IndicatorSpec, threshold v1alpha1.Threshold) Rule {
	labels := map[string]string{
		"product": document.Spec.Product.Name,
		"version": document.Spec.Product.Version,
		"level":   threshold.Level,
	}

	for k, v := range document.ObjectMeta.Labels {
		labels[k] = v
	}

	interpolatedPromQl := strings.Replace(i.PromQL, "$step", i.Alert.Step, -1)

	return Rule{
		Alert:       i.Name,
		Expr:        fmt.Sprintf("%s %s %+v", interpolatedPromQl, indicator.GetComparator(threshold.Operator), threshold.Value),
		For:         i.Alert.For,
		Labels:      labels,
		Annotations: i.Documentation,
	}
}
