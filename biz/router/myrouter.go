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
	CreateNewService(ctx context.Context, d *domain.Container) (string, error)
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
			ctrH.POST("/", handler.CreateContainer)
		}
	}
}

type createServiceReq struct {
	Name        string            `json:"name,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9_-]*$')"`
	Image       string            `json:"image,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9_:.-]*$')"`
	Labels      map[string]string `json:"labels,omitempty" vd:"range($, #k < 50 && #v < 50)"`
	Env         []string          `json:"env,omitempty" vd:"range($, regexp('^[A-Z0-9_]*$')) "`
	Limit       domain.Resource   `json:"limit,required"`
	Reservation domain.Resource   `json:"reservation"`
	Replica     uint64            `json:"replica,required" vd:"$<1000 && $>0"`
	Endpoint    []domain.Endpoint `json:"endpoint,required"`
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

	svcIdResp, err := m.svc.CreateNewService(ctx, &domain.Container{
		Name:        req.Name,
		CreatedTime: time.Now(),
		Image:       req.Image,
		Labels:      req.Labels,
		Env:         req.Env,
		Limit:       req.Limit,
		Reservation: req.Reservation,
		Replica:     req.Replica,
		Endpoint:    req.Endpoint,
	})
	if err != nil {

	}

	resp := &createContainerResp{
		Message: "Your Container created successfully",
		Container: domain.Container{
			CreatedTime: time.Now(),
			ServiceID:   svcIdResp,
			Name:        req.Name,
			Labels: map[string]string{
				"userID": "lintangpkk",
			},
			Replica: 3,
			Limit: domain.Resource{
				CPUs:   req.Limit.CPUs,
				Memory: req.Limit.Memory,
			},
			Image: req.Image,
			Env:   []string{"lalala"},
			Endpoint: []domain.Endpoint{{
				TargetPort:    req.Endpoint[0].TargetPort,
				PublishedPort: req.Endpoint[0].PublishedPort,
				Protocol:      req.Endpoint[0].Protocol,
			},
			},
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
