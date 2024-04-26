package tests

/*
This utils is used to generate dummy certificates for SSL connection to Redis.
*/

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// GenerateTestCertificates generates a dummy CA certificate and a client certificate
// signed by that CA for testing purposes. It returns the PEM-encoded certificates and key.
func GenerateTestCertificates() (caCertPEM, certPEM, keyPEM []byte, err error) {
	// Generate CA certificate:

	// Create a new private key using elliptic curve cryptography
	caKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	// Define the template for the CA certificate
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),                                         // Unique serial number for the certificate
		Subject:      pkix.Name{Organization: []string{"Test CA"}},          // Define details of the entity the certificate represents
		NotBefore:    time.Now(),                                            // Certificate validity start time
		NotAfter:     time.Now().Add(24 * time.Hour),                        // Certificate expiry time (24 hours from now)
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, // Define how the certificate can be used
		IsCA:         true,                                                  // Indicate that this is a CA certificate
	}

	// Create the CA certificate using the template and private key
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}
	// Convert the DER-encoded certificate into PEM format
	caCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	// Generate client certificate:

	// Create a new private key for the client certificate
	clientKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	// Define the template for the client certificate
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{Organization: []string{"Test Client"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// Create the client certificate using the template, CA certificate, and private key
	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caTemplate, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}
	// Convert the DER-encoded client certificate into PEM format
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})

	// Convert the private key into DER format
	keyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return nil, nil, nil, err
	}
	// Convert the DER-encoded key into PEM format
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return
}
