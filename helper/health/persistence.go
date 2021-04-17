package health

// Persistence represents a generic struct model for persistence
// Name is free text. i.e: shipment
// type is free text. i.e: postgreSQL, redis, etc
type Persistence struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Status    string  `json:"status"`
	PingError *string `json:"ping_error,omitempty"`

	HealthPersistence HealthPersistence `json:"-"`
}

// NewPersistence to create new Instance for Persistence struct
func NewPersistence(name, persistenceType string, hp HealthPersistence) Persistence {
	return Persistence{
		Name:              name,
		Type:              persistenceType,
		HealthPersistence: hp,
	}
}

// Persistences represents array of Persistence
type Persistences []Persistence

// Ping will check each persintece status
// and assign status field for Persistence
// one persistence failed, will return false
// it indicates that service is unhealthy
func (p *Persistences) Ping() (ok bool) {
	ok = true
	for i, persistance := range *p {
		if err := persistance.HealthPersistence.Ping(); err != nil {
			errMsg := err.Error()
			persistance.Status = "FAILED"
			persistance.PingError = &errMsg
			ok = false
		} else {
			persistance.Status = "OK"
			persistance.PingError = nil
		}
		(*p)[i] = persistance
	}
	return
}

// HealthPersistence is interface contains methods those need to be implemented by the actual persistence
type HealthPersistence interface {
	Ping() error
}
