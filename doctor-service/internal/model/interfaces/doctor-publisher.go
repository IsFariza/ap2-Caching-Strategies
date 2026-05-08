package interfaces

import "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"

type DoctorPublisher interface {
	PublishDoctorCreated(doc *model.Doctor) error
}
