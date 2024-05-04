package main

import (
	"context"
	pb "dogker/lintang/container-service/kitex_gen/container-service/pb"
)

// ContainerServiceImpl implements the last service interface defined in the IDL.
type ContainerServiceImpl struct{}

func NewContainerService() ContainerServiceImpl {
	return ContainerServiceImpl{}
}

// Hello implements the ContainerServiceImpl interface.
func (s *ContainerServiceImpl) Hello(ctx context.Context, req *pb.HelloReq) (resp *pb.HelloResp, err error) {
	// TODO: Your code here...
	return
}
