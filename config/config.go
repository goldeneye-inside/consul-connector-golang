package config

// Config is configuration of Consul connector
type Config struct {
	Address     string
	Token       string
	Certificate string
	// If Certificate property is empty, the Certificate value will be read from the file at CertificatePath
	CertificatePath string
}
