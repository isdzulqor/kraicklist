package handler

import "github.com/isdzulqor/kraicklist/helper/health"

type Root struct {
	Advertisement *Advertisement
	Health        *health.HealthHandler
}
