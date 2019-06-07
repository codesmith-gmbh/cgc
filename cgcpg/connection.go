package cgcpg

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/rdsutils"
	"github.com/gobuffalo/packr/v2"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"strconv"
)

const rootCertificateFile = "rds-ca-2015-root.pem"

var certPool = createCertPool()

func createCertPool() *x509.CertPool {
	certBox := packr.New("Certificates", "./certificates")
	certPool := x509.NewCertPool()
	pemCert, err := certBox.Find(rootCertificateFile)
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(pemCert)
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
