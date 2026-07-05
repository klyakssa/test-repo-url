package router

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klyakssa/test-repo-url/internal/config"
)

type MyRouter struct {
	*gin.Engine
	Config *config.Config
	server *http.Server
}

func NewMyRouter(cfg *config.Config) *MyRouter {
	engine := gin.Default()
	return &MyRouter{
		Engine: engine,
		Config: cfg,
		server: &http.Server{
			Addr:         cfg.WebConfig.HostPort,
			Handler:      engine,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (r *MyRouter) Run(ctx context.Context, webConfig *config.WebConfig) error {
	errChan := make(chan error, 1)

	go func() {
		if webConfig.EnableHTTPS {
			if err := generateTLSCertificates(webConfig.CertFile, webConfig.KeyFile); err != nil {
				errChan <- err
			}
			if err := r.server.ListenAndServeTLS(webConfig.CertFile, webConfig.KeyFile); err != nil {
				errChan <- err
			}
			return
		}
		if err := r.server.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

func (r *MyRouter) SGET(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.Engine.GET(pattern, func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})
}

func (r *MyRouter) SPOST(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.Engine.POST(pattern, func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})
}

func (r *MyRouter) Middleware(middleware ...gin.HandlerFunc) {
	r.Engine.Use(middleware...)
}

func (r *MyRouter) Group(grp string) *RouterGroup {
	return NewGroup(r.Engine.Group(grp))
}

func generateTLSCertificates(certFile, keyFile string) error {

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil
		}
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certFileHandle, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer func() {
		err = certFileHandle.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if err := pem.Encode(certFileHandle, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	keyFileHandle, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() {
		err = keyFileHandle.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if err := pem.Encode(keyFileHandle, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}
