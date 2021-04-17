package handler

import "kraicklist/helper/health"

type Root struct {
	Advertisement *Advertisement
	Health        *health.HealthHandler
}
