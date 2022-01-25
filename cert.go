package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type CustomTLS struct {
	TLSCA              string
	TLSCert            string
	TLSKey             string
	InsecureSkipVerify bool
}

func (c CustomTLS) TLSConfig() (*tls.Config, error) {
	if c.TLSCA == "" && c.TLSCert == "" && c.TLSKey == "" {
		return nil, nil
	}
	cfg := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	if c.TLSCA != "" {
		pool, err := loadPool(c.TLSCA)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = pool
	}
	if c.TLSCert != "" && c.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
		cfg.Certificates = []tls.Certificate{cert}
	}

	return cfg, nil
}

func loadPool(ca string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	pem, err := os.ReadFile(ca)
	if err != nil {
		return nil, fmt.Errorf("could not read tls CA file: %v", err)
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to load CA file as x509 cert")
	}
	return pool, nil
}
