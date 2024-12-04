package certretainpolicies

type CertificateRetainAll struct {
}


func (c *CertificateRetainAll) DeprecateCertificate(tx db.Querier, certificate *types.CertificateInfo) error {