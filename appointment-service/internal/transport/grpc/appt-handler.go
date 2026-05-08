package grpc

import (
	"context"

	pb "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/appt_proto"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
)

type appointmentHandler struct {
	pb.UnimplementedAppointmentServiceServer
	uc interfaces.AppointmentUsecase
}

func NewAppointmentHandler(uc interfaces.AppointmentUsecase) *appointmentHandler {
	return &appointmentHandler{
		uc: uc,
	}
}

func (h *appointmentHandler) CreateAppointment(ctx context.Context, req *pb.CreateAppointmentRequest) (*pb.AppointmentResponse, error) {

	appt := toDomain(req)

	err := h.uc.Create(ctx, appt)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return toProto(appt), nil
}

func (h *appointmentHandler) GetAppointment(ctx context.Context, req *pb.GetAppointmentRequest) (*pb.AppointmentResponse, error) {
	appt, err := h.uc.GetById(ctx, req.Id)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return toProto(appt), nil
}

func (h *appointmentHandler) ListAppointments(ctx context.Context, req *pb.ListAppointmentsRequest) (*pb.ListAppointmentsResponse, error) {
	appts, err := h.uc.GetAll(ctx)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	var pbAppts []*pb.AppointmentResponse
	for _, appt := range appts {
		pbAppts = append(pbAppts, toProto(appt))
	}
	return &pb.ListAppointmentsResponse{
		Appointments: pbAppts,
	}, nil
}

func (h *appointmentHandler) UpdateAppointmentStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.AppointmentResponse, error) {
	err := h.uc.Update(ctx, req.Id, model.Status(req.Status))
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	updated, err := h.uc.GetById(ctx, req.Id)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}
	return toProto(updated), nil
}
