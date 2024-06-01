package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	pb "dogker/lintang/container-service/kitex_gen/container-service/pb"
	"errors"

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

// StopContainerCreditLimit implements the ContainerGRPCServiceImpl interface.
func (s *ContainerGRPCServiceImpl) StopContainerCreditLimit(ctx context.Context, req *pb.StopUserContainerCreditLimitReq) (resp *pb.StopUserContainerCreditLimitRes, err error) {
	// TODO: Your code here...
	userCtrs, err := s.containerRepo.GetAllUserContainers(ctx, req.UserID)
	if err != nil {
		zap.L().Error("s.containerRepo.GetAllUserContainers", zap.Error(err))
	}

	for _, ctr := range *userCtrs {
		// stop container di docker

		ctrID := ctr.ServiceID
		if ctr.Status == domain.ServiceStopped {
			continue
		}
		err = s.dockerAPI.Stop(ctx, ctrID, ctr.UserID, &ctr)
		if err != nil {
			zap.L().Error("s.dockerAPI.Stop (StopContainerCreditLimit) (ContainerGRPC)", zap.Error(err))
			continue // karena container udah  di terminated jadi ya lanjut ke container selanjutnya aja
		}

		lastLifeCycleID := qSortWaktu(ctr.ContainerLifecycles).ID
		err := s.containerRepo.UpdateContainerLifecycleStatus(ctx, domain.ContainerStatusSTOPPED, lastLifeCycleID)
		if err != nil {
			zap.L().Error("s.containerRepo.UpdateContainerLifecycleStatus ()StopContainerCreditLimit (ContainerGRPC)", zap.Error(err))
		}

	}

	var ctrs []*domain.Container
	for _, ctr := range *userCtrs {
		if ctr.Status == domain.ServiceStopped {
			continue
		}
		ctrs = append(ctrs, &domain.Container{
			ID:            ctr.ID,
			UserID:        ctr.UserID,
			Image:         ctr.Image,
			Status:        domain.ServiceStatus(ctr.Status),
			Name:          ctr.Name,
			ContainerPort: int(ctr.ContainerPort),
			ServiceID:     ctr.ServiceID,
		})
	}
	if len(ctrs) != 0 {
		err = s.containerRepo.BatchUpdateContainer(ctx, ctrs) // update status container jadi stop
		if err != nil {
			zap.L().Error("s.containerRepo.BatchUpdateContainer (StopContainerCreditLimit) (ContainerGRPC)", zap.Error(err))
			return nil, status.Errorf(getStatusCode(err), "%v", err)
		}
	}

	// err = s.containerRepo.BatchUpdateContainerLifecycle(ctx, ctrs) //  update status containerlifecycle jadi stopped
	// if err != nil {
	// 	zap.L().Error("s.containerRepo.BatchUpdateContainerLifecycle (StopContainerCreditLimit) (ContainerGRPC)", zap.Error(err))
	// 	return nil, status.Errorf(getStatusCode(err), "%v", err)
	// }

	res := &pb.StopUserContainerCreditLimitRes{
		Message: "user container succesfully stopped",
	}

	return res, nil
}

func (s *ContainerGRPCServiceImpl) GetContainerStatus(ctx context.Context, req *pb.GetContainerStatusReq) (resp *pb.GetContainerStatusRes, err error) {
	// TODO: Your code here...

	ctr, err := s.containerRepo.Get(ctx, req.ServiceID)
	if err != nil {
		zap.L().Error("s.containerRepo.Get(ctx, serviceIDs) (TerminatedAccidentally) (ContainerService)", zap.String("serviceIDs", req.ServiceID), zap.Error(err))
		return nil, status.Errorf(getStatusCode(err), "containers not found: %v", err)
	}
	ctrFromDockerAPI, err := s.dockerAPI.Get(ctx, req.ServiceID, ctr)
	if err != nil {
		zap.L().Error(".dockerAPI.Ge(ctx, serviceIDs) (TerminatedAccidentally) (ContainerService)", zap.String("serviceIDs", req.ServiceID), zap.Error(err))
		return nil, status.Errorf(getStatusCode(err), "containers not found: %v", err)
	}

	res := &pb.GetContainerStatusRes{
		Status: ctrFromDockerAPI.Status == domain.ServiceRun && ctr.Status == domain.ServiceRun,
	}

	return res, nil
}

// ContainerTerminatedAccidentally implements the ContainerGRPCServiceImpl interface.
// kalau latestCtrLife.Sttaus == RUN dan sebelumnya distop , tetep masuk disini
// misal stoped terus sampai 10 detik masih stopped berarti ya diinsert lifecycle barunya....
func (s *ContainerGRPCServiceImpl) ContainerTerminatedAccidentally(ctx context.Context, req *pb.ContainerTerminatedAccidentallyReq) (resp *pb.ContainerTerminatedAccidentallyRes, err error) {
	// TODO: Your code here...
	// get containers detail dari list of service Ids

	// yang iini kukomen semua karna error & gaperlu juga karena kalo service distop lewat cli bakal di start seendiri sama swarm

	// filter dulu pastiin container nya masih stopped di docker api
	// karna kalau coontainer distop terus dirun dengan sengaja oleh user , nanti status di docker api masinh run
	// for i, _ := range req.ServiceIDs {
	// 	ctr, err := s.containerRepo.Get(ctx, req.ServiceIDs[i])
	// 	if err != nil {
	// 		zap.L().Error("s.containerRepo.Get(ctx, serviceIDs) (TerminatedAccidentally) (ContainerService)", zap.Strings("serviceIDs", req.ServiceIDs), zap.Error(err))
	// 		return nil, status.Errorf(getStatusCode(err), "containers not found: %v", err)
	// 	}
	// 	ctrFromDockerAPI, err := s.dockerAPI.Get(ctx, req.ServiceIDs[i], ctr)
	// 	if err != nil {
	// 		zap.L().Error(".dockerAPI.Ge(ctx, serviceIDs) (TerminatedAccidentally) (ContainerService)", zap.Strings("serviceIDs", req.ServiceIDs), zap.Error(err))
	// 		return nil, status.Errorf(getStatusCode(err), "containers not found: %v", err)
	// 	}

	// 	if ctrFromDockerAPI.Status == domain.ServiceRun && ctr.Status == domain.ServiceRun {
	// 		req.ServiceIDs[i] = req.ServiceIDs[len(req.ServiceIDs)-1]
	// 		req.ServiceIDs = req.ServiceIDs[:len(req.ServiceIDs)-1]
	// 	}

	// }

	// zap.L().Info(fmt.Sprintf("down ServiceIDs: %s", req.ServiceIDs), zap.Strings("serviceIDs", req.ServiceIDs))
	// ctrsDB, err := s.containerRepo.GetContainersDetail(ctx, req.ServiceIDs)
	// if err != nil {
	// 	zap.L().Error("s.containerRepo.GetContainersDetail(ctx, serviceIDs) (TerminatedAccidentally) (ContainerService)", zap.Strings("serviceIDs", req.ServiceIDs), zap.Error(err))
	// 	return nil, status.Errorf(getStatusCode(err), "containers not found: %v", err)
	// }

	// //  only filter container yang sebelumnya gak terminatted, karena yg sebelumnya terminated di db udah ada metrics nya di tabel metrics && status di tabel conatainer == terminated && status di tabel lifecycle == stopped
	// for i := range ctrsDB {
	// 	if ctrsDB[i].Status == domain.ServiceTerminated {
	// 		// hapus cotnainer yang sebelumnya statusnya terminated, dari ctrsDB
	// 		// delete inplace arraynya
	// 		ctrsDB[i] = ctrsDB[len(ctrsDB)-1]
	// 		ctrsDB = ctrsDB[:len(ctrsDB)-1]
	// 	}
	// }

	// // get metrics dari setiap container dari monitor service (loop O(n))
	// var ctrMetrics []domain.Metric
	// for i, _ := range ctrsDB {
	// 	metric, err := s.monitorClient.GetSpecificContainerMetrics(ctx, ctrsDB[i].ServiceID, ctrsDB[i].UserID, ctrsDB[i].CreatedTime)
	// 	if err != nil {
	// 		zap.L().Error(" s.monitorClient.GetSpecificContainerMetrics (ContainerTerminatedAccidentally) (containerGRPCService)", zap.Error(err))
	// 		return nil, status.Errorf(getStatusCode(err), "%v", err)
	// 	}
	// 	ctrMetrics = append(ctrMetrics, *metric)
	// }

	// // batch insert container metrics untuk setiap container tadi
	// err = s.containerRepo.BatchInsertContainerMetrics(ctx, ctrMetrics)
	// if err != nil {
	// 	zap.L().Error("s.containerRepo.BatchInsertContainerMetrics(ctx, ctrMetrics) (TerminatedAccidentally) (ContainerService)", zap.Error(err))
	// 	return nil, status.Errorf(getStatusCode(err), "%v", err)
	// }

	// yang bawah ini gaperlu karena kalau distop secara tidak disengaja (docker stop ...) nanti otomatis start sendiri lagi
	// jadi gakperlu update status container &  gaperlu update status contianer lifecycle

	//  update batch  semua container tadi,  jadi stopped di tabel container
	// err = s.containerRepo.BatchUpdateContainer(ctx, ctrsDB)
	// if err != nil {
	// 	zap.L().Error("s.containerRepo.BatchUpdateContainer(ctx, ctrsDB)", zap.Error(err))
	// 	return nil, status.Errorf(getStatusCode(err), "%v", err)
	// }

	// for i, _ := range ctrsDB {
	// 	cLife := qSortWaktu(ctrsDB[i].ContainerLifecycles)                                                  // get latest containrelifecycle
	// 	err := s.containerRepo.UpdateContainerLifecycleStatus(ctx, domain.ContainerStatusSTOPPED, cLife.ID) // update containrelifecycle status = stopped
	// 	if err != nil {
	// 		zap.L().Error("s.containerRepo.UpdateContainerLifecycleStatus(ctx, ctrsDB)", zap.Error(err))
	// 		return nil, status.Errorf(getStatusCode(err), "%v", err)
	// 	}

	// 	stoppedCtrFromDockerAPI, err := s.containerRepo.Get(ctx, ctrsDB[i].ServiceID)
	// 	if err != nil {
	// 		zap.L().Error("s.containerRepo.Get (ContainerTerminatedAccidentally) (ContainerService)", zap.Error(err))
	// 		return nil, status.Errorf(getStatusCode(err), "%v", err)
	// 	}

	// 	// if stoppedCtrFromDockerAPI.Status == domain.ServiceRun {

	// 	// }

	// 	// insert new ctr lifecycle dengan status RUN
	// 	_, err = s.containerRepo.InsertLifecycle(ctx, &domain.ContainerLifecycle{
	// 		ContainerID: ctrsDB[i].ID,
	// 		StartTime:   time.Now(),
	// 		Status:      domain.ContainerStatusRUN,
	// 		Replica:     stoppedCtrFromDockerAPI.Replica,
	// 	}) // insert new ctr lifecycle dengan status RUN
	// 	if err != nil {
	// 		zap.L().Error("s.containerRepo.InsertLifecycle (RecoverContainerAfterStoppedAccidentally) (ContainerService)", zap.Error(err))
	// 	}

	// }

	// // update lifecycle semua container tadi jadi stopped karena emang mati
	// // err = s.containerRepo.BatchUpdateContainerLifecycle(ctx, ctrsDB)// ini salah cok

	// // karena terminated accidentally
	// err = s.containerRepo.BatchUpdateRunStatusContainer(ctx, ctrsDB) // update status container jd run buat recovered container
	// if err != nil {
	// 	zap.L().Error("s.containerRepo.BatchUpdateRunStatusContainer (StopContainerCreditLimit) (ContainerService)", zap.Error(err))
	// 	return nil, err
	// }
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
