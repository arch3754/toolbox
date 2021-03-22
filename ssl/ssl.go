package ssl

import (
	"crypto/tls"
	"crypto/x509"
)

func ConfigTls(Server, CA, ServerKey []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(Server, ServerKey)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(CA)
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            pool,
		InsecureSkipVerify: true,
	}, nil
}