package client

import (
	"context"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
	doctorpb "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type doctorClient struct {
	client doctorpb.DoctorServiceClient
}

func NewDoctorClient(conn *grpc.ClientConn) interfaces.DoctorClient {
	return &doctorClient{
		client: doctorpb.NewDoctorServiceClient(conn),
	}
}

func (dc *doctorClient) DoctorExists(ctx context.Context, doctorId string) (bool, error) {
	_, err := dc.client.GetDoctor(ctx, &doctorpb.GetDoctorRequest{Id: doctorId})

	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
