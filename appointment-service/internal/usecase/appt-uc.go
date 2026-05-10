package usecase

import (
	"context"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
	"github.com/google/uuid"
)

type appointmentUsecase struct {
	repo         interfaces.AppointmentRepo
	doctorClient interfaces.DoctorClient
	publisher    interfaces.AppointmentPublisher
}

func NewAppointmentUsecase(repo interfaces.AppointmentRepo, dc interfaces.DoctorClient, pub interfaces.AppointmentPublisher) interfaces.AppointmentUsecase {
	return &appointmentUsecase{
		repo:         repo,
		doctorClient: dc,
		publisher:    pub,
	}
}

func (uc *appointmentUsecase) Create(ctx context.Context, appt *model.Appointment) error {
	if appt.Title == "" {
		return model.ErrTitleRequired
	}
	if appt.DoctorID == "" {
		return model.ErrDoctorIDRequired
	}

	exists, err := uc.doctorClient.DoctorExists(ctx, appt.DoctorID)
	if err != nil {
		return model.ErrDoctorServiceUnavailable
	}
	if !exists {
		return model.ErrDoctorNotFoundRemote
	}
	appt.Status = model.StatusNew
	appt.ID = uuid.New().String()
	ucErr := uc.repo.Create(ctx, appt)
	if ucErr != nil {
		return ucErr
	}
	uc.publisher.PublishCreated(appt)

	return nil
}

func (uc *appointmentUsecase) GetById(ctx context.Context, id string) (*model.Appointment, error) {
	if id == "" {
		return nil, model.ErrInvalidID
	}
	return uc.repo.GetById(ctx, id)
}

func (uc *appointmentUsecase) GetAll(ctx context.Context) ([]*model.Appointment, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *appointmentUsecase) Update(ctx context.Context, id string, newStatus model.Status) error {

	if id == "" {
		return model.ErrInvalidID
	}
	currentStatus, err := uc.repo.GetById(ctx, id)
	if err != nil {
		return err
	}

	if currentStatus.Status == model.StatusDone && newStatus == model.StatusNew {
		return model.ErrInvalidStatusTransition
	}
	if newStatus != model.StatusDone && newStatus != model.StatusInProgress &&
		newStatus != model.StatusNew {
		return model.ErrInvalidStatus
	}
	err = uc.repo.Update(ctx, id, newStatus)
	if err != nil {
		return err
	}

	oldStatus := currentStatus.Status
	currentStatus.Status = newStatus
	uc.publisher.PublishStatusUpdated(currentStatus, oldStatus, newStatus)
	return nil
}
