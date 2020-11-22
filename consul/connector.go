package consul

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"

	"gitlab.com/goldeneye-inside/consul/library/golang/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Connector invokes methods of Consul service
type Connector interface {
	RegisterService(hostName, serviceName, serviceIP string, servicePort uint32, healthCheckURL string) error
	GetServiceAddress(serviceName string) (string, error)
}

// NewConnector inits a new instance of Connector
func NewConnector(config *config.Config) Connector {
	return &connectorGRPC{
		config: config,
	}
}

type connectorGRPC struct {
	config *config.Config
}

func (svc *connectorGRPC) connect() (*grpc.ClientConn, error) {
	var (
		err   error
		creds credentials.TransportCredentials
	)
	if len(svc.config.Certificate) == 0 {
		creds, err = credentials.NewClientTLSFromFile(svc.config.CertificatePath, "")
		if err != nil {
			return nil, err
		}
	} else {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM([]byte(svc.config.Certificate)) {
			return nil, errors.New("failed to add server CA's certificate")
		}
		creds = credentials.NewClientTLSFromCert(certPool, "")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	return grpc.Dial(svc.config.Address, opts...)
}

func (svc *connectorGRPC) RegisterService(hostName, serviceName, serviceIP string, servicePort uint32, healthCheckURL string) error {
	dial, err := svc.connect()
	if err != nil {
		return err
	}
	defer dial.Close()

	_, err = NewConsulConnectorClient(dial).RegisterService(context.TODO(), &ReqRegisterService{
		ConsulToken:    svc.config.Token,
		ServiceName:    serviceName,
		ServiceId:      fmt.Sprintf("%s-%d", hostName, servicePort),
		ServiceIp:      serviceIP,
		ServicePort:    servicePort,
		HealthCheckUrl: healthCheckURL,
	})
	return err
}

func (svc *connectorGRPC) GetServiceAddress(serviceName string) (string, error) {
	dial, err := svc.connect()
	if err != nil {
		return "", err
	}
	defer dial.Close()

	resp, err := NewConsulConnectorClient(dial).GetServiceAddress(context.TODO(), &ReqGetServiceAddress{
		ConsulToken: svc.config.Token,
		ServiceName: serviceName,
	})
	if err != nil {
		return "", err
	}
	return resp.Address, nil
}
