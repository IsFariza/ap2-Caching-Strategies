package interfaces

import "context"

type DoctorClient interface {
	DoctorExists(ctx context.Context, doctorId string) (bool, error)
}
