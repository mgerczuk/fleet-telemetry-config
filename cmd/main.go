package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"github.com/mgerczuk/fleet-telemetry-config/api"
	"github.com/mgerczuk/fleet-telemetry-config/auth"
	"github.com/mgerczuk/fleet-telemetry-config/config"
	log "github.com/sirupsen/logrus"
)

func main() {

	var applicationConfig string
	var persistFile string

	flag.StringVar(&applicationConfig, "config", "config.json", "application configuration file")
	flag.StringVar(&persistFile, "persist", "persist.json", "application persistent data")
	flag.Parse()

	configData, err := config.LoadApplicationConfiguration(applicationConfig)
	if err != nil {
		panic(fmt.Sprintf("Error loading config data from '%s': %s", applicationConfig, err.Error()))
	}

	err = config.InitPersist(persistFile)
	if err != nil {
		panic(fmt.Sprintf("Error loading persistent data from '%s': %s", persistFile, err.Error()))
	}

	muxPublic := http.NewServeMux()
	muxPublic.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})
	muxPublic.HandleFunc("/auth/callback", auth.CodeCallback)
	muxPublic.HandleFunc("/.well-known/appspecific/com.tesla.3p.public-key.pem", api.GetPublicKey)

	fs := http.FileServer(http.Dir(configData.PrivateServer.WebRoot))

	muxPrivate := http.NewServeMux()
	muxPrivate.Handle("/", fs)
	api.HandleDataModel(muxPrivate, configData)
	muxPrivate.HandleFunc("POST /api/send_telemetry_config", api.SendTelemetryConfig(*configData))
	muxPrivate.HandleFunc("/api/vehicle_telemetry_config", api.VehicleTelemetryConfig(*configData))

	muxPrivate.HandleFunc("POST /api/register", api.Register(*configData))
	muxPrivate.HandleFunc("/auth/request", auth.StartAuth(*configData))
	muxPrivate.HandleFunc("POST /api/initial_token", api.GetInitialToken(*configData))
	muxPrivate.HandleFunc("POST /api/refresh_token", api.RefreshToken(*configData))

	publicServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", configData.PublicServer.Port),
		Handler: muxPublic,
		TLSConfig: &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				newCert, err := tls.LoadX509KeyPair(configData.PublicServer.Cert, configData.PublicServer.Key)
				if err != nil {
					return nil, err
				}
				return &newCert, nil
			},
			MinVersion: tls.VersionTLS12, // improves cert reputation score at https://www.ssllabs.com/ssltest/
		},
	}

	privateServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", configData.PrivateServer.Port),
		Handler: muxPrivate,
	}

	go func() {
		log.Infof("Public server started on port %v", configData.PublicServer.Port)
		panic(publicServer.ListenAndServeTLS("", ""))
	}()

	go func() {
		log.Infof("Private server started on port %v", configData.PrivateServer.Port)
		panic(privateServer.ListenAndServe())
	}()

	select {}
}
