package issue

import "crypto/x509/pkix"

func newCSRSubject(commonName, country, province, organization, organizationalUnit string) pkix.Name {
	subj := pkix.Name{CommonName: commonName}
	if country != "" {
		subj.Country = []string{country}
	}
	if province != "" {
		subj.Province = []string{province}
	}
	if organization != "" {
		subj.Organization = []string{organization}
	}
	if organizationalUnit != "" {
		subj.OrganizationalUnit = []string{organizationalUnit}
	}
	return subj
}
