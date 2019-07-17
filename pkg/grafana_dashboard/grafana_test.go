package grafana_dashboard_test

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
)

func TestDocumentToDashboard(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)
		document := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:          "test_indicator",
						PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
						Documentation: map[string]string{"title": "Test Indicator Title"},
						Thresholds: []v1alpha1.Threshold{{
							Level:    "critical",
							Operator: v1alpha1.GreaterThan,
							Value:    1000,
						}, {
							Level:    "warning",
							Operator: v1alpha1.LessThanOrEqualTo,
							Value:    700,
						}},
					},
					{
						Name:       "second_test_indicator",
						PromQL:     `rate(gorouter_requests[1m])`,
						Thresholds: []v1alpha1.Threshold{},
					},
				},
				Layout: v1alpha1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(*dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
			Title: "Indicator Test Dashboard",
			Rows: []grafana_dashboard.GrafanaRow{{
				Title: "Test Section Title",
				Panels: []grafana_dashboard.GrafanaPanel{
					{
						Title: "Test Indicator Title",
						Type:  "graph",
						Targets: []grafana_dashboard.GrafanaTarget{{
							Expression: `sum_over_time(gorouter_latency_ms[30m])`,
						}},
						Thresholds: []grafana_dashboard.GrafanaThreshold{{
							Value:     1000,
							ColorMode: "critical",
							Op:        "gt",
							Fill:      true,
							Line:      true,
							Yaxis:     "left",
						}, {
							Value:     700,
							ColorMode: "warning",
							Op:        "lt",
							Fill:      true,
							Line:      true,
							Yaxis:     "left",
						}},
					},
				},
			}},
			Annotations: grafana_dashboard.GrafanaAnnotations{
				List: []grafana_dashboard.GrafanaAnnotation{
					{
						Enable:      true,
						Expr:        "ALERTS{product=\"\"}",
						TagKeys:     "level",
						TitleFormat: "{{alertname}} is {{alertstate}} in the {{level}} threshold",
						IconColor:   "#1f78c1",
					},
				},
			},
		}))
	})

	t.Run("uses the IP layout information to create distinct rows", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:          "test_indicator",
						PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
						Documentation: map[string]string{"title": "Test Indicator Title"},
					},
					{
						Name:   "second_test_indicator",
						PromQL: `rate(gorouter_requests[1m])`,
					},
				},
				Layout: v1alpha1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "foo",
							Indicators: []string{"second_test_indicator"},
						},
						{
							Title:      "bar",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Rows[0].Title).To(Equal("foo"))
		g.Expect(dashboard.Rows[0].Panels[0].Title).To(Equal("second_test_indicator"))
		g.Expect(dashboard.Rows[1].Title).To(Equal("bar"))
		g.Expect(dashboard.Rows[1].Panels[0].Title).To(Equal("Test Indicator Title"))
	})

	t.Run("falls back to product name/version when layout title is missing", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		indicators := []v1alpha1.IndicatorSpec{
			{
				Name:   "test_indicator",
				PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
			},
		}
		document := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "test product",
					Version: "v0.9",
				},
				Indicators: indicators,
				Layout: v1alpha1.Layout{
					Sections: []v1alpha1.Section{
						{
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Title).To(BeEquivalentTo("test product - v0.9"))
	})

	t.Run("replaces $step with $__interval", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		indicators := []v1alpha1.IndicatorSpec{
			{
				Name:   "test_indicator",
				PromQL: `rate(sum_over_time(gorouter_latency_ms[$step])[$step])`,
			},
		}
		document := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: indicators,
				Layout: v1alpha1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Rows[0].Panels[0].Targets[0].Expression).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$__interval])[$__interval])`))
	})

	t.Run("replaces only $step with $__interval", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		indicators := []v1alpha1.IndicatorSpec{
			{
				Name:   "test_indicator",
				PromQL: `rate(sum_over_time(gorouter_latency_ms[$steper])[$STEP])`,
			},
			{
				Name:   "another_indicator",
				PromQL: `avg_over_time(demo_latency{source_id="$step"}[5m])`,
			},
		}
		document := v1alpha1.IndicatorDocument{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"ste": "123"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{

				Indicators: indicators,
				Layout: v1alpha1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator", "another_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(dashboard.Rows[0].Panels[0].Targets[0].Expression).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$steper])[$__interval])`))
		g.Expect(dashboard.Rows[0].Panels[1].Targets[0].Expression).ToNot(BeEquivalentTo(`avg_over_time(demo_latency{source_id="123p"}[5m])`))
	})

	t.Run("creates a filename based on product name and contents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.LessThan,
						Value:    5,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Test Indicator Title"},
				}},
				Layout: v1alpha1.Layout{
					Title: "Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		docBytes, err := json.Marshal(document)
		g.Expect(err).ToNot(HaveOccurred())
		filename := grafana_dashboard.DashboardFilename(docBytes, "test_product")
		// Should have a SHA1 in the middle, but don't want to specify the SHA
		g.Expect(filename).To(MatchRegexp("test_product_[a-f0-9]{40}\\.json"))
	})

	t.Run("includes annotations based on product & metadata alerts", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.LessThan,
						Value:    5,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Test Indicator Title"},
				}, {
					Name:   "second_test_indicator",
					PromQL: "second_test_query",
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.GreaterThan,
						Value:    10,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Second Test Indicator Title"},
				}},
				Layout: v1alpha1.Layout{
					Title: "Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator", "second_test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Annotations.List).To(ConsistOf(grafana_dashboard.GrafanaAnnotation{
			Enable:      true,
			Expr:        "ALERTS{product=\"test_product\",deployment=\"test_deployment\"}",
			TagKeys:     "level",
			TitleFormat: "{{alertname}} is {{alertstate}} in the {{level}} threshold",
			IconColor:   "#1f78c1",
		}))
	})
}
