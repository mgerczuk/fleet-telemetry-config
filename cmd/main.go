package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"github.com/mgerczuk/fleet-telemetry-config/api"
	"github.com/mgerczuk/fleet-telemetry-config/auth"
	"github.com/mgerczuk/fleet-telemetry-config/config"
	"github.com/mgerczuk/fleet-telemetry-config/teslalogger_api"
	"github.com/mgerczuk/fleet-telemetry-config/util"
	log "github.com/sirupsen/logrus"
)

var version = "local build"

func main() {

	var configFilename string
	var persistFilename string

	flag.StringVar(&configFilename, "config", "config.json", "application configuration file")
	flag.StringVar(&persistFilename, "persist", "persist.json", "application persistent data")
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	fmt.Printf("fleet-telemetry-config version %s\n", version)
	if *showVersion {
		return
	}

	configData, err := config.LoadApplicationConfiguration(configFilename)
	if err != nil {
		panic(fmt.Sprintf("Error loading config data from '%s': %s", configFilename, err.Error()))
	}

	err = config.InitPersist(persistFilename)
	if err != nil {
		panic(fmt.Sprintf("Error loading persistent data from '%s': %s", persistFilename, err.Error()))
	}

	err = api.CheckCertificate(configData)
	if err != nil {
		panic(fmt.Sprintf("Error checking certificates: %s", err.Error()))
	}

	muxPublic := http.NewServeMux()
	muxPublic.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})
	muxPublic.HandleFunc("/auth/callback", auth.CodeCallback)
	muxPublic.HandleFunc("GET /.well-known/appspecific/com.tesla.3p.public-key.pem", api.GetPublicKey)
	muxPublic.HandleFunc("GET /.well-known/appspecific/challenge", api.GetChallenge)

	fs := http.FileServer(http.Dir(configData.PrivateServer.WebRoot))

	muxPrivate := http.NewServeMux()
	muxPrivate.Handle("/", fs)
	api.HandleDataModel(muxPrivate, configData)
	muxPrivate.HandleFunc("POST /api/send_telemetry_config", api.SendTelemetryConfig(*configData))
	muxPrivate.HandleFunc("/api/vehicle_telemetry_config", api.VehicleTelemetryConfig(*configData))

	muxPrivate.HandleFunc("POST /api/register", api.Register(*configData))
	muxPrivate.HandleFunc("/auth/request", auth.StartAuth(*configData))
	muxPrivate.HandleFunc("POST /api/initial_token", api.GetInitialToken(*configData))
	muxPrivate.HandleFunc("POST /teslaredirect/refresh_token.php", teslalogger_api.RefreshToken)

	publicServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", configData.PublicServer.Port),
		Handler: util.HttpLogHandler(muxPublic),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // improves cert reputation score at https://www.ssllabs.com/ssltest/
		},
	}

	privateServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", configData.PrivateServer.Port),
		Handler: muxPrivate,
	}

	go func() {
		log.Infof("Public server started on port %v", configData.PublicServer.Port)
		panic(publicServer.ListenAndServeTLS(configData.PublicServer.Cert, configData.PublicServer.Key))
	}()

	go func() {
		log.Infof("Private server started on port %v", configData.PrivateServer.Port)
		panic(privateServer.ListenAndServe())
	}()

	select {}
}
