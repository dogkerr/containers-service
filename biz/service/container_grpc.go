package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	pb "dogker/lintang/container-service/kitex_gen/container-service/pb"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ContainerGRPCServiceImpl implements the last service interface defined in the IDL.
type ContainerGRPCServiceImpl struct {
	containerRepo ContainerRepository
	dockerAPI     DockerEngineAPI
	dkronAPI      DkronAPI
	monitorClient MonitorClient
	minioAPI      MinioAPI
}

func NewContainerGRPCService(c ContainerRepository, d DockerEngineAPI, dkron DkronAPI, monitorSvc MonitorClient,
	minioAPI MinioAPI) *ContainerGRPCServiceImpl {
	return &ContainerGRPCServiceImpl{
		c, d, dkron, monitorSvc, minioAPI,
	}
}

// Hello implements the ContainerGRPCServiceImpl interface.
func (s *ContainerGRPCServiceImpl) Hello(ctx context.Context, req *pb.HelloReq) (resp *pb.HelloResp, err error) {
	// TODO: Your code here...
	return nil, nil
}

// ContainerTerminatedAccidentally implements the ContainerGRPCServiceImpl interface.
func (s *ContainerGRPCServiceImpl) ContainerTerminatedAccidentally(ctx context.Context, req *pb.ContainerTerminatedAccidentallyReq) (resp *pb.ContainerTerminatedAccidentallyRes, err error) {
	// TODO: Your code here...
	// get containers detail dari list of service Ids
	zap.L().Info(fmt.Sprintf("down ServiceIDs: %s", req.ServiceIDs), zap.Strings("serviceIDs", req.ServiceIDs))
	ctrsDB, err := s.containerRepo.GetContainersDetail(ctx, req.ServiceIDs)
	if err != nil {
		zap.L().Error("s.containerRepo.GetContainersDetail(ctx, serviceIDs) (TerminatedAccidentally) (ContainerService)", zap.Strings("serviceIDs", req.ServiceIDs))
		return nil, status.Errorf(getStatusCode(err), "containers not found: %v", err)
	}

	//  only filter container yang sebelumnya gak terminatted, karena yg sebelumnya terminated di db udah ada metrics nya di tabel metrics && status di tabel conatainer == terminated && status di tabel lifecycle == stopped
	for i := range ctrsDB {
		if ctrsDB[i].Status == domain.ServiceTerminated {
			// hapus cotnainer yang sebelumnya statusnya terminated, dari ctrsDB
			// delete inplace arraynya
			ctrsDB[i] = ctrsDB[len(ctrsDB)-1]
			ctrsDB = ctrsDB[:len(ctrsDB)-1]
		}
	}

	// get metrics dari setiap container dari monitor service (loop O(n))
	var ctrMetrics []domain.Metric
	for i, _ := range ctrsDB {
		metric, err := s.monitorClient.GetSpecificContainerMetrics(ctx, ctrsDB[i].ServiceID, ctrsDB[i].UserID, ctrsDB[i].CreatedTime)
		if err != nil {
			zap.L().Error(" s.monitorClient.GetSpecificContainerMetrics (ContainerTerminatedAccidentally) (containerGRPCService)", zap.Error(err))
			return nil, status.Errorf(getStatusCode(err), "%v", err)
		}
		ctrMetrics = append(ctrMetrics, *metric)

	}

	// batch insert container metrics untuk setiap container tadi
	err = s.containerRepo.BatchInsertContainerMetrics(ctx, ctrMetrics)
	if err != nil {
		zap.L().Error("s.containerRepo.BatchInsertContainerMetrics(ctx, ctrMetrics) (TerminatedAccidentally) (ContainerService)")
		return nil, status.Errorf(getStatusCode(err), "%v", err)
	}

	//  update batch  semua container tadi,  jadi stopped di tabel container
	err = s.containerRepo.BatchUpdateContainer(ctx, ctrsDB)
	if err != nil {
		zap.L().Error("s.containerRepo.BatchUpdateContainer(ctx, ctrsDB)")
		return nil, status.Errorf(getStatusCode(err), "%v", err)
	}

	// update lifecycle semua container tadi jadi stopped karena emang mati
	err = s.containerRepo.BatchUpdateContainerLifecycle(ctx, ctrsDB)
	if err != nil {
		zap.L().Error("s.containerRepo.BatchUpdateContainerLifecycle(ctx, ctrsDB)")
		return nil, status.Errorf(getStatusCode(err), "%v", err)
	}
	return &pb.ContainerTerminatedAccidentallyRes{Message: "ok"}, nil
}

func getStatusCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	var ierr *domain.Error
	if !errors.As(err, &ierr) {
		return codes.Internal
	} else {
		switch ierr.Code() {
		case domain.ErrInternalServerError:
			return codes.Internal
		case domain.ErrNotFound:
			return codes.NotFound
		case domain.ErrConflict:
			return codes.Internal
		case domain.ErrBadParamInput:
			return codes.InvalidArgument
		default:
			return codes.Internal
		}
	}

}
