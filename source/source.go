package source

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Source struct defines abstruct client for Kubernetes API.
// Source uses ClientSet to call API endpoint of Kubernetes.
type Source struct {
	ClientSet kubernetes.Interface
}

// NewSource creates Source instance that defined Ingresses function.
func NewSource(clientSet kubernetes.Interface) *Source {
	return &Source{
		ClientSet: clientSet,
	}
}

// Ingresses returns list of Ingress that masked unnecessary fields
// Ingress struct is defined by ingress.go
func (s *Source) Ingresses() ([]*Ingress, error) {
	ingressList, err := s.ClientSet.ExtensionsV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ingresses := make([]*Ingress, len(ingressList.Items))
	for i, item := range ingressList.Items {

		ingressTLSs := make([]*IngressTLS, len(item.Spec.TLS))
		for j, tls := range item.Spec.TLS {

			endpoints := make([]*TLSEndpoint, len(tls.Hosts))
			for k, host := range tls.Hosts {
				// TODO: Support port numbers other than default
				endpoints[k] = NewTLSEndpoint(host, "")
			}

			ingressTLSs[j] = &IngressTLS{
				Endpoints:  endpoints,
				SecretName: tls.SecretName,
			}
		}

		ingresses[i] = &Ingress{
			ClusterName: item.ObjectMeta.ClusterName,
			Namespace:   item.ObjectMeta.Namespace,
			Name:        item.ObjectMeta.Name,
			TLS:         ingressTLSs,
		}
	}

	return ingresses, nil
}
