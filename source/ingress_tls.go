package source

// IngressTLS expresses information about existing IngressTLS.
// Controller requires some fileds of original Ingress struct.
// So, this definition masks unnecessary fields of https://godoc.org/k8s.io/api/extensions/v1beta1#IngressTLS
type IngressTLS struct {
	Endpoints  []*TLSEndpoint
	SecretName string
}
