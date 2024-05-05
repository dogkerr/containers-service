package router

/*
 ini router yg dipake bukan yg di router.go

*/

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"dogker/lintang/container-service/biz/model/basic/hello"
	"dogker/lintang/container-service/biz/router/middleware"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type ContainerService interface {
	Hello(context.Context) (string, error)
	CreateNewService(ctx context.Context, d *domain.Container) (string, time.Time, *domain.ContainerLifecycle, error)
}

type ContainerHandler struct {
	svc ContainerService
}

func MyRouter(r *server.Hertz, c ContainerService) {

	handler := &ContainerHandler{
		svc: c,
	}

	root := r.Group("/api/v1")
	{
		root.GET("/lalala", append(middleware.Protected(), handler.SayHello)...)
		ctrH := root.Group("/containers")
		{
			ctrH.POST("/", append(middleware.Protected(), handler.CreateContainer)...)
		}
	}
}

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}


type endpoint struct {
	TargetPort    uint32 `json:"target_port,required" vd:"$<65555 && $>0"`
	PublishedPort uint64 `json:"published_port,required" vd:"$<65555 && $>0"`
	Protocol      string `json:"protocol" default:"tcp" vd:"in($, 'tcp','udp','sctp')" `
}

type createServiceReq struct {
	Name        string            `json:"name,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9_-]*$')"`
	Image       string            `json:"image,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9_:.-]*$')"`
	Labels      map[string]string `json:"labels,omitempty" vd:"range($, #k < 50 && #v < 50) || !$"`
	Env         []string          `json:"env,omitempty" vd:"range($, regexp('^[A-Z0-9_]*$')) || !$ "`
	Limit       domain.Resource          `json:"limit,required"`
	Reservation domain.Resource          `json:"reservation,omitempty"`
	Replica     uint64            `json:"replica,required" vd:"$<1000 && $>0"`
	Endpoint    []endpoint        `json:"endpoint,required"`
}

type createContainerResp struct {
	Message   string           `json:"message"`
	Container domain.Container `json:"container"`
}

func (m *ContainerHandler) CreateContainer(ctx context.Context, c *app.RequestContext) {
	var req createServiceReq

	err := c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	var dEndpoint []domain.Endpoint
	for _, endp := range req.Endpoint {
		dEndpoint = append(dEndpoint, domain.Endpoint{
			TargetPort:    endp.TargetPort,
			PublishedPort: endp.PublishedPort,
			Protocol:      endp.Protocol,
		})
	}
	userId, _ := c.Get("userID")

	svcIdResp, createdTime, ctrLife, err := m.svc.CreateNewService(ctx, &domain.Container{
		Name:        req.Name,
		CreatedTime: time.Now(),
		Image:       req.Image,
		Labels:      req.Labels,
		Env:         req.Env,
		Limit:       domain.Resource(req.Limit),
		Reservation: domain.Resource(req.Reservation),
		Replica:     req.Replica,
		Endpoint:    dEndpoint,
		UserID:      userId.(string),
	})
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	resp := &createContainerResp{
		Message: "Your Container created successfully",
		Container: domain.Container{
			CreatedTime: createdTime,
			ServiceID:   svcIdResp,
			Name:        req.Name,
			Labels: req.Labels,
			Replica: 3,
			Limit: domain.Resource{
				CPUs:   req.Limit.CPUs,
				Memory: req.Limit.Memory,
			},
			Image:               req.Image,
			Env:                 req.Env,
			Endpoint:            dEndpoint,
			UserID:              userId.(string),
			Status:              domain.ContainerStatusRUN,
			ContainerPort:       int(req.Endpoint[0].TargetPort),
			PublicPort:          int(req.Endpoint[0].PublishedPort),
			ContainerLifecycles: []domain.ContainerLifecycle{*ctrLife},
		},
	}
	c.JSON(http.StatusOK, resp)
}

type HelloReq struct {
	Name string `query:"name,required"`
}

func (m *ContainerHandler) SayHello(ctx context.Context, c *app.RequestContext) {
	var req HelloReq

	err := c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(hello.HelloResp)
	resp.RespBody = "halo " + req.Name
	c.JSON(http.StatusOK, resp)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	case domain.ErrBadParamInput:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
