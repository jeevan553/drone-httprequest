package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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
	isClientCert := p.SslCertPath != ""
	isProxy := p.Proxy != ""

	fmt.Println("p.keystore", p.KeystorePath)
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
		p.httpClient, err = setupSslWithClientCertNoProxy(p.SslCertPath, p.KeystorePath, p.Password)
		if err != nil {
			return err
		}

	// SSL required, client cert provided, proxy enabled
	case !isIgnoreSsl && isClientCert && isProxy:
		p.httpClient, err = setupSslWithClientCertWithProxy(p.SslCertPath, p.Proxy, p.KeystorePath, p.Password)
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
		p.httpClient, err = setupNoSslWithClientCertNoProxy(p.SslCertPath, p.KeystorePath, p.Password)
		if err != nil {
			return err
		}

	// No SSL, client cert provided, proxy enabled
	case isIgnoreSsl && isClientCert && isProxy:
		p.httpClient, err = setupNoSslWithClientCertWithProxy(p.SslCertPath, p.Proxy, p.KeystorePath, p.Password)
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
func setupSslWithClientCertNoProxy(certPath string, KeystorePath string, Password string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, false, KeystorePath, Password)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// SSL required, client cert provided, proxy enabled
func setupSslWithClientCertWithProxy(certPath string, proxy string, KeystorePath string, Password string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, false, KeystorePath, Password)
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
func setupNoSslWithClientCertNoProxy(certPath string, KeystorePath string, Password string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, true, KeystorePath, Password)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// no SSL, client cert provided, proxy enabled
func setupNoSslWithClientCertWithProxy(certPath string, proxy string, KeystorePath string, Password string) (*http.Client, error) {
	tlsConfig, err := createTlsConfigWithClientCert(certPath, true, KeystorePath, Password)
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
func createTlsConfigWithClientCert(certPath string, ignoreSsl bool, KeystorePath string, Password string) (*tls.Config, error) {

	certs, err := generateCleanCerts(certPath, KeystorePath, Password)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	fmt.Println("Generated clean-certs.pem:\n", certs)
	log.Println("Attempting to load certificate from:", certPath)

	// Create a new certificate pool and append the custom CA certificate
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(certs)) {
		log.Println("Failed to append custom CA certificate:")
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	log.Println("Successfully loaded custom CA certificate")

	// Return TLS configuration with only the custom CA
	return &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: ignoreSsl,
	}, nil
}

func generateCleanCerts(certPath, keystorePath, password string) (string, error) {
	fmt.Println("certPath", certPath)
	fmt.Println("keystorePath", keystorePath)
	fmt.Println("password", password)

	exec.Command("keytool", "-delete", "-alias", certPath, "-keystore", keystorePath, "-storepass", password).Run()

	cmd1 := exec.Command("keytool", "-import", "-trustcacerts",
		"-alias", certPath,
		"-file", certPath,
		"-keystore", keystorePath,
		"-storepass", password,
		"-noprompt",
	)
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	err := cmd1.Run()
	if err != nil {
		return "", fmt.Errorf("error importing certificate: %v", err)
	}
	// Step 2: List the contents of the keystore (for debugging, optional)
	cmd2 := exec.Command("keytool", "-list", "-v", "-keystore", keystorePath, "-storepass", password)
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	err = cmd2.Run()
	if err != nil {
		return "", fmt.Errorf("error listing keystore: %v", err)
	}

	// Step 3: Convert JKS to PKCS12 format
	cmd3 := exec.Command("keytool", "-importkeystore",
		"-srckeystore", keystorePath,
		"-srcstoretype", "pkcs12",
		"-srcstorepass", password,
		"-destkeystore", "truststore.p12",
		"-deststoretype", "pkcs12",
		"-deststorepass", password,
		"-noprompt",
	)
	cmd3.Stdout = os.Stdout
	cmd3.Stderr = os.Stderr
	err = cmd3.Run()
	if err != nil {
		return "", fmt.Errorf("error converting keystore to PKCS12: %v", err)
	}

	// Step 4: Extract certificates into certs.pem
	cmd4 := exec.Command("openssl", "pkcs12", "-in", "truststore.p12", "-out", "certs.pem", "-nokeys", "-passin", fmt.Sprintf("pass:%s", password))
	cmd4.Stdout = os.Stdout
	cmd4.Stderr = os.Stderr
	err = cmd4.Run()
	if err != nil {
		return "", fmt.Errorf("error extracting certs.pem: %v", err)
	}

	// Step 5: Filter certificates into clean-certs.pem
	cmd5 := exec.Command("awk", "/BEGIN CERTIFICATE/,/END CERTIFICATE/", "certs.pem")
	output, err := cmd5.Output()
	if err != nil {
		return "", fmt.Errorf("error filtering clean-certs.pem: %v", err)
	}
	err = ioutil.WriteFile("clean-certs.pem", output, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing clean-certs.pem: %v", err)

	}
	// Read and return the clean-certs.pem contents
	cleanCerts, err := ioutil.ReadFile("clean-certs.pem")
	if err != nil {
		return "", fmt.Errorf("error reading clean-certs.pem: %v", err)
	}

	return string(cleanCerts), nil
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
