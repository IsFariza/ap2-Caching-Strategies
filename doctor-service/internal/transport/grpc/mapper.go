package grpc

import (
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
	doctorpb "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/proto"
)

func toDomain(req *doctorpb.CreateDoctorRequest) *model.Doctor {
	return &model.Doctor{
		FullName:       req.GetFullName(),
		Specialization: req.GetSpecialization(),
		Email:          req.GetEmail(),
	}
}
func toProto(doc *model.Doctor) *doctorpb.DoctorResponse {
	return &doctorpb.DoctorResponse{
		Id:             doc.ID,
		FullName:       doc.FullName,
		Specialization: doc.Specialization,
		Email:          doc.Email,
	}
}
