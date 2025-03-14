package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mgerczuk/fleet-telemetry-config/config"
	log "github.com/sirupsen/logrus"
)

func getChecksum(filename string) (string, error) {

	hasher := sha256.New()
	cert, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("cannot access cert %s: %s", filename, err.Error())
	}
	hasher.Write(cert)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func CheckCertificate(configData *config.Config) error {
	data := config.LockPersist()
	defer data.Unlock()

	cs, err := getChecksum(configData.PublicServer.Cert)
	if err != nil {
		return err
	}

	if cs == data.CAChecksum {
		return nil
	}

	log.Info("Refreshing telemetry config certificates...")
	for userId := range data.Users {
		err := RefreshTelemetryConfigCertificate(configData, data, userId)
		if err != nil {
			log.Fatalf("*** RefreshTelemetryConfig failed: %s", err.Error())
		}
	}
	log.Info("Refreshing telemetry config certificates - done")

	data.CAChecksum = cs
	config.PutPersist(data)

	return nil
}
