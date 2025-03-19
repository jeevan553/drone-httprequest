package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

/*
	SSL Cert is for the client to state who the client is
	SSL no ignore is to have a secure connection with the server
	SSL is proxy is to go through a proxy server

	All these are independent of each other and 8 combinations are possible

	isIgnoreSsl,isClientCert,isProxy,Combination Description
	false,false,false,SSL required, no client cert, no proxy
	false,false,true,SSL required, no client cert, proxy enabled
	false,true,false,SSL required, client cert provided, no proxy
	false,true,true,SSL required, client cert provided, proxy enabled
	true,false,false,No SSL, no client cert, no proxy
	true,false,true,No SSL skipping, no client cert, proxy enabled
	true,true,false,No SSL skipping, client cert provided, no proxy
	true,true,true,No SSL skipping, client cert provided, proxy enabled

*/

func (p *Plugin) SetHttpConnectionParameters() error {

	isIgnoreSsl := p.IgnoreSsl
	// isClientCert := p.SslCertPath != ""
	isClientCert := true
	certPath := p.RootCertPaths // ✅ Assign RootCertPaths to certPath
	fmt.Println("🔍 Using certPath:", certPath)
	isProxy := p.Proxy != ""

	LogPrintf(p, "Configuration Ignore SSL: %t, Client Cert: %t, Proxy: %t\n", isIgnoreSsl, isClientCert, isProxy)

	var err error

	switch {
	// SSL required, no client cert, no proxy
	case !isIgnoreSsl && !isClientCert && !isProxy:
		p.httpClient, err = setupSslNoClientCertNoProxy()
		if err != nil {
			return err
		}

	// SSL required, no client cert, proxy enabled
	case !isIgnoreSsl && !isClientCert && isProxy:
		p.httpClient, err = setupSslNoClientCertWithProxy(p.Proxy)
		if err != nil {
			return err
		}

	// SSL required, client cert provided, no proxy
	case !isIgnoreSsl && isClientCert && !isProxy:
		p.httpClient, err = setupSslWithClientCertNoProxy(p.RootCertPaths)
		if err != nil {
			return err
		}

	// SSL required, client cert provided, proxy enabled
	case !isIgnoreSsl && isClientCert && isProxy:
		p.httpClient, err = setupSslWithClientCertWithProxy(p.SslCertPath, p.Proxy)
		if err != nil {
			return err
		}

	// No SSL, no client cert, no proxy
	case isIgnoreSsl && !isClientCert && !isProxy:
		p.httpClient, err = setupNoSslNoClientCertNoProxy()
		if err != nil {
			return err
		}

	// No SSL, no client cert, proxy enabled
	case isIgnoreSsl && !isClientCert && isProxy:
		p.httpClient, err = setupNoSslNoClientCertWithProxy(p.Proxy)
		if err != nil {
			return err
		}

	// No SSL, client cert provided, no proxy
	case isIgnoreSsl && isClientCert && !isProxy:
		p.httpClient, err = setupNoSslWithClientCertNoProxy(p.SslCertPath)
		if err != nil {
			return err
		}

	// No SSL, client cert provided, proxy enabled
	case isIgnoreSsl && isClientCert && isProxy:
		p.httpClient, err = setupNoSslWithClientCertWithProxy(p.SslCertPath, p.Proxy)
		if err != nil {
			return err
		}
	}

	return nil
}

// SSL required, no client cert, no proxy
func setupSslNoClientCertNoProxy() (*http.Client, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: false}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// SSL required, no client cert, proxy enabled
func setupSslNoClientCertWithProxy(proxy string) (*http.Client, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: false}
	transport, err := createTransportWithProxy(proxy, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

// SSL required, client cert provided, no proxy
func setupSslWithClientCertNoProxy(certPath string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, false)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// SSL required, client cert provided, proxy enabled
func setupSslWithClientCertWithProxy(certPath string, proxy string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, false)
	if err != nil {
		return nil, err
	}
	transport, err := createTransportWithProxy(proxy, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

// no SSL, no client cert, no proxy
func setupNoSslNoClientCertNoProxy() (*http.Client, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// no SSL, no client cert, proxy enabled
func setupNoSslNoClientCertWithProxy(proxy string) (*http.Client, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport, err := createTransportWithProxy(proxy, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

// no SSL, client cert provided, no proxy
func setupNoSslWithClientCertNoProxy(certPath string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, true)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// no SSL, client cert provided, proxy enabled
func setupNoSslWithClientCertWithProxy(certPath string, proxy string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, true)
	if err != nil {
		return nil, err
	}
	transport, err := createTransportWithProxy(proxy, tlsConfig)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

// Function to create TLS configuration with client certificate
func createTlsConfigWithClientCert(certPath string, ignoreSsl bool) (*tls.Config, error) {
	fmt.Println("🔍 Attempting to load certificate from:", certPath)

	// Read the custom CA certificate
	caCert, err := ioutil.ReadFile(certPath)
	if err != nil {
		fmt.Println("❌ Error loading certificate:", certPath, err)
		return nil, err
	}

	// Create a new certificate pool and append the custom CA certificate
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		fmt.Println("❌ Failed to append custom CA certificate:", certPath)
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	fmt.Println("✅ Successfully loaded custom CA certificate")

	// Return TLS configuration with only the custom CA
	return &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: ignoreSsl,
	}, nil
}

// Function to create HTTP transport with proxy and TLS configuration
func createTransportWithProxy(proxyUrl string, tlsConfig *tls.Config) (*http.Transport, error) {
	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, errors.New("invalid proxy URL: " + err.Error())
	}
	return &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: tlsConfig,
	}, nil
}
