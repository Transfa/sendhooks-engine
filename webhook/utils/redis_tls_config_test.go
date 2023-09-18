package utils

/*
This is a test to ensure that given path of certificates required to run Redis  with SSL, the files are read and the
configuration can work. For this purpose, we use a  custom certificates generator for our tests.
*/

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	certificates "webhook/utils/tests"
)

func FormatErrorBody(errorType string) string {

	return fmt.Sprintf("error creating temp %s file", errorType)
}

func TestCreateTLSConfig(t *testing.T) {
	// Generate test certificates
	caCertPEM, certPEM, keyPEM, err := certificates.GenerateTestCertificates()
	assert.NoError(t, err, "Error generating test certificates")

	// Create temporary files to store these certificates and keys
	caCertFile, err := os.CreateTemp("", "caCert")
	assert.NoError(t, err, FormatErrorBody("CA cert"))
	defer os.Remove(caCertFile.Name())
	caCertFile.Write(caCertPEM)
	caCertFile.Close()

	clientCertFile, err := os.CreateTemp("", "clientCert")
	assert.NoError(t, err, FormatErrorBody("client cert"))

	defer os.Remove(clientCertFile.Name())
	clientCertFile.Write(certPEM)
	clientCertFile.Close()

	clientKeyFile, err := os.CreateTemp("", "clientKey")
	assert.NoError(t, err, FormatErrorBody("client key"))

	defer os.Remove(clientKeyFile.Name())
	clientKeyFile.Write(keyPEM)
	clientKeyFile.Close()

	// Optimist Test: valid CA certificate
	tlsConfig, err := CreateTLSConfig(caCertFile.Name(), "", "")
	assert.NoError(t, err, "Expected no error creating TLS config with only CA cert")

	//Subjects is deprecated for the moment, but it looks like there is no other alternatives. https://github.com/golang/go/issues/46287
	assert.NotEmpty(t, tlsConfig.RootCAs.Subjects(), "Expected CA certificates to be loaded")

	// Optimist Test: valid CA certificate with valid client certificate and key
	tlsConfig, err = CreateTLSConfig(caCertFile.Name(), clientCertFile.Name(), clientKeyFile.Name())
	assert.NoError(t, err, "Expected no error creating TLS config with client cert and key")
	assert.NotEmpty(t, tlsConfig.Certificates, "Expected client certificates to be loaded")

	// Pessimist Test: invalid file paths
	_, err = CreateTLSConfig("invalidPath", "", "")
	assert.Error(t, err, "Expected error creating TLS config with invalid CA cert path")

	_, err = CreateTLSConfig(caCertFile.Name(), "invalidPath", "invalidPath")
	assert.Error(t, err, "Expected error creating TLS config with invalid client cert and key paths")
}
