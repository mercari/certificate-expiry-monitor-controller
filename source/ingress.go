package source

// Ingress expresses information about existing Ingress.
// Controller requires some fileds of original Ingress struct.
// So, this definition masks unnecessary fields of https://godoc.org/k8s.io/api/extensions/v1beta1#Ingress
type Ingress struct {
	ClusterName string
	Namespace   string
	Name        string
	TLS         []*IngressTLS
}
