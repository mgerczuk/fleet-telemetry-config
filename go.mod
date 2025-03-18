module github.com/mgerczuk/fleet-telemetry-config

go 1.23.4

replace github.com/teslamotors/vehicle-command => github.com/mgerczuk/vehicle-command-api v0.3.3-api

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/sirupsen/logrus v1.9.3
	github.com/teslamotors/vehicle-command v0.3.3
)

require (
	github.com/cronokirby/saferith v0.33.0 // indirect
	github.com/google/uuid v1.6.0
	golang.org/x/sys v0.31.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)
