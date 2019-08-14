// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/clientset/versioned/typed/indicatordocument/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeAppsV1 struct {
	*testing.Fake
}

func (c *FakeAppsV1) Indicators(namespace string) v1.IndicatorInterface {
	return &FakeIndicators{c, namespace}
}

func (c *FakeAppsV1) IndicatorDocuments(namespace string) v1.IndicatorDocumentInterface {
	return &FakeIndicatorDocuments{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAppsV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
