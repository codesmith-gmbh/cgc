package cgcpg

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/rdsutils"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"strconv"
)

var rdsCa2015Root = []byte(`-----BEGIN CERTIFICATE-----
MIID9DCCAtygAwIBAgIBQjANBgkqhkiG9w0BAQUFADCBijELMAkGA1UEBhMCVVMx
EzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUxIjAgBgNVBAoM
GUFtYXpvbiBXZWIgU2VydmljZXMsIEluYy4xEzARBgNVBAsMCkFtYXpvbiBSRFMx
GzAZBgNVBAMMEkFtYXpvbiBSRFMgUm9vdCBDQTAeFw0xNTAyMDUwOTExMzFaFw0y
MDAzMDUwOTExMzFaMIGKMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3Rv
bjEQMA4GA1UEBwwHU2VhdHRsZTEiMCAGA1UECgwZQW1hem9uIFdlYiBTZXJ2aWNl
cywgSW5jLjETMBEGA1UECwwKQW1hem9uIFJEUzEbMBkGA1UEAwwSQW1hem9uIFJE
UyBSb290IENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuD8nrZ8V
u+VA8yVlUipCZIKPTDcOILYpUe8Tct0YeQQr0uyl018StdBsa3CjBgvwpDRq1HgF
Ji2N3+39+shCNspQeE6aYU+BHXhKhIIStt3r7gl/4NqYiDDMWKHxHq0nsGDFfArf
AOcjZdJagOMqb3fF46flc8k2E7THTm9Sz4L7RY1WdABMuurpICLFE3oHcGdapOb9
T53pQR+xpHW9atkcf3pf7gbO0rlKVSIoUenBlZipUlp1VZl/OD/E+TtRhDDNdI2J
P/DSMM3aEsq6ZQkfbz/Ilml+Lx3tJYXUDmp+ZjzMPLk/+3beT8EhrwtcG3VPpvwp
BIOqsqVVTvw/CwIDAQABo2MwYTAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUw
AwEB/zAdBgNVHQ4EFgQUTgLurD72FchM7Sz1BcGPnIQISYMwHwYDVR0jBBgwFoAU
TgLurD72FchM7Sz1BcGPnIQISYMwDQYJKoZIhvcNAQEFBQADggEBAHZcgIio8pAm
MjHD5cl6wKjXxScXKtXygWH2BoDMYBJF9yfyKO2jEFxYKbHePpnXB1R04zJSWAw5
2EUuDI1pSBh9BA82/5PkuNlNeSTB3dXDD2PEPdzVWbSKvUB8ZdooV+2vngL0Zm4r
47QPyd18yPHrRIbtBtHR/6CwKevLZ394zgExqhnekYKIqqEX41xsUV0Gm6x4vpjf
2u6O/+YE2U+qyyxHE5Wd5oqde0oo9UUpFETJPVb6Q2cEeQib8PBAyi0i6KnF+kIV
A9dY7IHSubtCK/i8wxMVqfd5GtbA8mmpeJFwnDvm9rBEsHybl08qlax9syEwsUYr
/40NawZfTUU=
-----END CERTIFICATE-----
`)

var certPool = createCertPool()

func createCertPool() *x509.CertPool {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(rdsCa2015Root)
	return certPool
}

type ConnConfigOption interface {
	Configure(config *pgx.ConnConfig) error
}

type RdsIamConfigOption struct {
	Host, Username, Database string
	Port                     int
	Config                   aws.Config
}

func (r RdsIamConfigOption) Configure(config *pgx.ConnConfig) error {
	endpoint := r.Host + ":" + strconv.Itoa(r.Port)
	token, err := rdsutils.BuildAuthToken(endpoint, r.Config.Region, r.Username, r.Config.Credentials)
	if err != nil {
		return errors.Wrapf(err, "could not get an auth token for the host %s, port %d and username %s", r.Host, r.Port, r.Username)
	}
	config.Host = r.Host
	config.Port = uint16(r.Port)
	config.User = r.Username
	config.Database = r.Database
	config.Password = token
	return nil
}

type TlsConfigOption struct {
	Host string
}

func (t TlsConfigOption) Configure(config *pgx.ConnConfig) error {
	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		ServerName: t.Host,
	}
	config.TLSConfig = tlsConfig
	return nil
}

func NewConfig(options ...ConnConfigOption) (pgx.ConnConfig, error) {
	config := pgx.ConnConfig{}
	for _, o := range options {
		if err := o.Configure(&config); err != nil {
			return pgx.ConnConfig{}, err
		}
	}
	return config, nil
}

func NewRdsSslConnection(cfg aws.Config, host string, port int, database, username string) (*pgx.Conn, error) {
	config, err := NewConfig(
		RdsIamConfigOption{Host: host, Port: port, Database: database, Username: username, Config: cfg},
		TlsConfigOption{Host: host},
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the connection configuration")
	}
	conn, err := pgx.Connect(config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to the database %s on host %s with user %s", database, host, username)
	}
	return conn, nil
}
