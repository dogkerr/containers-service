package webapi

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"dogker/lintang/container-service/pb"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MonitorClient struct {
	service pb.MonitorServiceClient
}

func NewMonitorClient(cc *grpc.ClientConn) *MonitorClient {
	service := pb.NewMonitorServiceClient(cc)
	return &MonitorClient{service: service}
}

func (m *MonitorClient) SendMetricsStopTerminatedContainerToBillingService(ctx context.Context, metricMonitor domain.UserMetricsMessage) error {
	grpcCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &pb.SendMetricsStopTerminatedContainerToBillingServiceReq{
		ContainerId:         metricMonitor.ContainerID,
		UserId:              metricMonitor.UserID,
		CpuUsage:            metricMonitor.CpuUsage,
		MemoryUsage:         metricMonitor.MemoryUsage,
		NetworkIngressUsage: metricMonitor.NetworkIngressUsage,
		NetworkEgressUsage:  metricMonitor.NetworkEgressUsage,
	}

	msg, err := m.service.SendMetricsStopTerminatedContainerToBillingService(grpcCtx, req)
	if err != nil {
		zap.L().Error("m.service.SendMetricsStopTerminatedContainerToBillingService ", zap.Error(err))
		return err
	}

	zap.L().Info(msg.Message)
	return nil
}

func (m *MonitorClient) GetSpecificContainerMetrics(ctx context.Context, ctrID string, userID string, serviceStartTime time.Time) (*domain.Metric, error) {

	grpcCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &pb.GetSpecificContainerResourceUsageRequest{
		UserId:      userID,
		ContainerId: ctrID,
		FromTime:    timestamppb.New(serviceStartTime),
	}

	ctrMetrics, err := m.service.GetSpecificContainerResourceUsage(grpcCtx, req)
	if err != nil {
		zap.L().Error("GetSpecificContainerResourceUsage", zap.Error(err))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	res := &domain.Metric{
		ContainerID:         ctrMetrics.UserContainer.Id,
		CpuUsage:            ctrMetrics.UserContainer.CpuUsage,
		MemoryUsage:         ctrMetrics.UserContainer.MemoryUsage,
		NetworkIngressUsage: ctrMetrics.UserContainer.NetworkIngressUsage,
		NetworkEgressUsage:  ctrMetrics.UserContainer.NetworkEgressUsage,
	}
	return res, nil
}
