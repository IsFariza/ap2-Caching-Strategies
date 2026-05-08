package grpc

import (
	"context"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model/interfaces"
	doctorpb "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/proto"
)

type doctorHandler struct {
	doctorpb.UnimplementedDoctorServiceServer
	uc interfaces.DoctorUseCase
}

func NewDoctorHandler(uc interfaces.DoctorUseCase) doctorpb.DoctorServiceServer {
	return &doctorHandler{uc: uc}
}

func (h *doctorHandler) CreateDoctor(ctx context.Context, req *doctorpb.CreateDoctorRequest) (*doctorpb.DoctorResponse, error) {
	newDoc := toDomain(req)

	savedDoc, err := h.uc.Create(ctx, newDoc)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}
	return toProto(savedDoc), nil
}

func (h *doctorHandler) GetDoctor(ctx context.Context, req *doctorpb.GetDoctorRequest) (*doctorpb.DoctorResponse, error) {
	doc, err := h.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}
	return toProto(doc), nil
}

func (h *doctorHandler) ListDoctors(ctx context.Context, req *doctorpb.ListDoctorsRequest) (*doctorpb.ListDoctorsResponse, error) {
	docs, err := h.uc.List(ctx)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	var pbDoctors []*doctorpb.DoctorResponse
	for _, d := range docs {
		pbDoctors = append(pbDoctors, toProto(d))
	}

	return &doctorpb.ListDoctorsResponse{
		Doctors: pbDoctors,
	}, nil
}
