package source

import (
	"testing"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngresses(t *testing.T) {
	ingressList := v1.IngressList{
		Items: []v1.Ingress{
			v1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "ingress1",
					Namespace:   "namespace1",
					ClusterName: "clusterName",
					Labels: map[string]string{
						"protocol": "tls",
					},
				},
				Spec: v1.IngressSpec{
					TLS: []v1.IngressTLS{
						{
							Hosts:      []string{"1.example.com"},
							SecretName: "ingressSecret1",
						},
					},
				},
			},
			v1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "ingress2",
					Namespace:   "namespace2",
					ClusterName: "clusterName",
					Labels: map[string]string{
						"protocol": "http",
					},
				},
				Spec: v1.IngressSpec{
					TLS: []v1.IngressTLS{
						{
							Hosts:      []string{"2.example.com"},
							SecretName: "ingressSecret2",
						},
					},
				},
			},
		},
	}

	clientSet := fake.NewSimpleClientset(&ingressList)
	source := NewSource(clientSet)
	actualIngresses, _ := source.Ingresses()

	expectedNum := len(ingressList.Items)
	if len(actualIngresses) != expectedNum {
		t.Fatalf("Unexpected number of Ingresses: %d", len(actualIngresses))
	}

	for i, ingress := range actualIngresses {
		if ingress.Name != ingressList.Items[i].ObjectMeta.Name {
			t.Fatalf("Unmatch expected Name: %s", ingress.Name)
		}
		if ingress.Namespace != ingressList.Items[i].ObjectMeta.Namespace {
			t.Fatalf("Unmatch expected Namespace: %s", ingress.Namespace)
		}
		if ingress.ClusterName != ingressList.Items[i].ObjectMeta.ClusterName {
			t.Fatalf("Unmatch expected ClusterName: %s", ingress.ClusterName)
		}

		for j, tls := range ingress.TLS {
			if len(tls.Endpoints) != len(ingressList.Items[i].Spec.TLS[j].Hosts) {
				t.Fatalf("Unexpected number of TLS Hosts: %d", len(tls.Endpoints))
			}
			if tls.SecretName != ingressList.Items[i].Spec.TLS[j].SecretName {
				t.Fatalf("Unmatch expected TLS SecretName: %s", tls.SecretName)
			}
		}
	}
}
