package router

/*
 ini router yg dipake bukan yg di router.go

*/

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"dogker/lintang/container-service/biz/model/basic/hello"
	"dogker/lintang/container-service/biz/router/middleware"
	"errors"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type ContainerService interface {
	Hello(context.Context) (string, error)
	CreateNewService(ctx context.Context, d *domain.Container) (string, time.Time, *domain.ContainerLifecycle, error)
	GetUserContainers(ctx context.Context, userID string, offset uint64, limit uint64) (*[]domain.Container, error)
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
			ctrH.GET("/", append(middleware.Protected(), handler.GetUsersContainer)...)
		}
	}
}

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

type createServiceReq struct {
	Name        string            `json:"name,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9_-]*$'); msg:'nama harus alphanumeric atau boleh juga simbol -,_ dan tidak boleh ada spasi'"`
	Image       string            `json:"image,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9_:-]*$'); msg:'image harus alphanumeric atau simbol -,_,:'"`
	Labels      map[string]string `json:"labels,omitempty" vd:"range($, len(#k) < 50 && len(#v) < 50) || !$; msg:'label haruslah kurang dari 50 '"`
	Env         []string          `json:"env,omitempty" vd:"range($, regexp('^[A-Z0-9_]*$')) || !$; msg:'env harus alphanumeric atau symbol _'"`
	Limit       domain.Resource   `json:"limit,required; msg:'resource limit harus anda isi '"`
	Reservation domain.Resource   `json:"reservation,omitempty" `
	Replica     uint64            `json:"replica,required" vd:"$<1000 && $>0; msg:'replica harus diantara 0-1000'"`
	Endpoint    []domain.Endpoint `json:"endpoint,required; msg:'endpoint wajib diisi'"`
}

type createContainerResp struct {
	Message   string           `json:"message"`
	Container domain.Container `json:"container"`
}

func (m *ContainerHandler) CreateContainer(ctx context.Context, c *app.RequestContext) {
	var req createServiceReq

	err := c.Bind(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	err = c.Validate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ResponseError{Message: err.Error()})
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
			Labels:      req.Labels,
			Replica:     3,
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

type Pagination struct {
	Offset uint64 `query:"page"`
	Limit  uint64 `query:"limit" vd:"$<100; msg:'limit tidak boleh lebih dari 100'"`
}

type getUserContainersResp struct {
	Containers *[]domain.Container `json:"containers"`
}

func (m *ContainerHandler) GetUsersContainer(ctx context.Context, c *app.RequestContext) {
	userId, _ := c.Get("userID")

	var req Pagination
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	resp, err := m.svc.GetUserContainers(ctx, userId.(string), req.Offset, req.Limit)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, getUserContainersResp{resp})
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
	var ierr *domain.Error
	if !errors.As(err, &ierr) {
		return http.StatusInternalServerError
	} else {
		switch ierr.Code() {
		case domain.ErrInternalServerError:
			return http.StatusInternalServerError
		case domain.ErrNotFound:
			return http.StatusNotFound
		case domain.ErrConflict:
			return http.StatusConflict
		case domain.ErrBadParamInput:
			return http.StatusBadRequest
		default:
			return http.StatusBadRequest
		}
	}

}
