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
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

type ContainerService interface {
	Hello(context.Context) (string, error)
	CreateNewService(ctx context.Context, d *domain.Container) (string, time.Time, *domain.ContainerLifecycle, error)
	GetUserContainers(ctx context.Context, userID string, offset uint64, limit uint64) (*[]domain.Container, error)
	GetContainer(ctx context.Context, ctrID string, userID string) (*domain.Container, error)
	StartContainer(ctx context.Context, ctrID string, userID string) (*domain.Container, error)
	StopContainer(ctx context.Context, ctrID string, userID string) error
	DeleteContainer(ctx context.Context, ctrID string, userID string) error
	UpdateContainer(ctx context.Context, d *domain.Container, ctrID string, userID string) (string, error)
	ScaleX(ctx context.Context, userID string, ctrID string, replica uint64) error
	Schedule(ctx context.Context, userID string, ctrID string, scheduledTime uint64, timeFormat domain.TimeFormat, action domain.ContainerAction) error
	ScheduleCreate(ctx context.Context, userID string, scheduledTime uint64, timeFormat domain.TimeFormat, action domain.ContainerAction, ctr *domain.Container) error
	CreateNewServiceAndUpload(ctx context.Context, d *domain.Container, imageFile *multipart.FileHeader,
		imageName string) (string, time.Time, *domain.ContainerLifecycle, error)
	GetUserContainersLoadTest(ctx context.Context, userID string, offset uint64, limit uint64) (*[]domain.Container, error)
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
			ctrH.POST("/upload", append(middleware.Protected(), handler.CreateContainerAndBuildImage)...)
			ctrH.GET("/", append(middleware.Protected(), handler.GetUsersContainer)...)
			ctrH.GET("/:id", append(middleware.Protected(), handler.GetContainer)...)
			ctrH.POST("/:id/start", append(middleware.Protected(), handler.StartContainer)...)
			// todo: post /:id/stop, delete /:id, put /:id, put /:id/scaleX
			ctrH.POST("/:id/stop", append(middleware.Protected(), handler.StopContainer)...)
			ctrH.DELETE("/:id", append(middleware.Protected(), handler.DeleteContainer)...)
			ctrH.PUT("/:id", append(middleware.Protected(), handler.UpdateContainer)...)
			ctrH.PUT("/:id/scale", append(middleware.Protected(), handler.ScaleX)...)

			// scheduling related
			ctrH.POST("/:id/schedule", append(middleware.Protected(), handler.ScheduleContainer)...)
			ctrH.POST("/create/schedule", append(middleware.Protected(), handler.CreateScheduledCreate)...)
			ctrH.POST("/scheduler/:id/stop", handler.ScheduledStop)
			ctrH.POST("/scheduler/:id/start", handler.ScheduledStart)
			ctrH.POST("/scheduler/create", handler.ScheduleCreate)
			ctrH.POST("/scheduler/:id/terminate", handler.ScheduleTerminate)
			// ctrH.POST("/scheduler/:id/create", handler. ))

			// load test
			ctrH.GET("/loadtest", append(middleware.Protected(), handler.GetUserContainersLoadTest)...)

		}
	}
}

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

type createServiceReq struct {
	Name        string            `json:"name,required" vd:"len($)<100 && regexp('^[\\w\\-\\.]*$'); msg:'nama harus alphanumeric atau boleh juga simbol -,_,. dan tidak boleh ada spasi'"`
	Image       string            `json:"image,required" vd:"len($)<100 && regexp('^([\\w\\-\\.\\/]*|[\\w\\-\\.\\/]*:[\\w\\-\\.]+)$'); msg:'image harus alphanumeric atau simbol -,_,:,/ atau juga bisa dengan format <imagename>:<tag>'"`
	Labels      map[string]string `json:"labels,omitempty" vd:"range($, len(#k) < 50 && len(#v) < 50) ; msg:'label haruslah kurang dari 50 '"`
	Env         []string          `json:"env" vd:"  range($, regexp('^([A-Z0-9\\_]*)=([A-Za-z0-9\\_\\/\\:\\@\\?\\(\\)\\'\\.\\=]*)$', #v) ); msg:'env harus dalam format KEY=VALUE dengan semua huruf kapital'"`
	Limit       domain.Resource   `json:"limit,required; msg:'resource limit harus anda isi '"`
	Reservation domain.Resource   `json:"reservation,omitempty" `
	Replica     int64             `json:"replica,required" vd:"$<1000 && $>=0; msg:'replica harus diantara 0-1000'"`
	Endpoint    []domain.Endpoint `json:"endpoint,required" vd:"@:len($)>0; msg:'endpoint wajib diisi'"`
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
		Replica:     uint64(req.Replica),
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

type createServiceAndBuildImageReq struct {
	Name string `form:"name,required" vd:"len($)<100 && regexp('^[\\w\\-\\.]*$'); msg:'nama harus alphanumeric atau boleh juga simbol -,_,. dan tidak boleh ada spasi'"`
	// Image       string            `form:"image,required" vd:"len($)<100 && regexp('^[a-zA-Z0-9/_:-]*$'); msg:'image harus alphanumeric atau simbol -,_,:,/'"`
	Labels      map[string]string     `form:"labels,omitempty" vd:"range($, len(#k) < 50 && len(#v) < 50) ; msg:'label haruslah kurang dari 50 '"`
	Env    []string          `json:"env" vd:"  range($, regexp('^([A-Z0-9\\_]*)=([A-Za-z0-9\\_\\/\\:\\@\\?\\(\\)\\'\\.\\=]*)$', #v) ); msg:'env harus dalam format KEY=VALUE dengan semua huruf kapital'"`
	Limit       domain.Resource       `form:"limit,required; msg:'resource limit harus anda isi '"`
	Reservation domain.Resource       `form:"reservation,omitempty" `
	Replica     int64                 `form:"replica,required" vd:"$<1000 && $>=0; msg:'replica harus diantara 0-1000'"`
	Endpoint    []domain.Endpoint     `form:"endpoint,required" vd:"@:len($)>0; msg:'endpoint wajib diisi'""`
	ImageTar    *multipart.FileHeader `form:"image,required" vd:"msg:'endpoint wajib diisi'"`
	ImageName   string                `form:"imageName,required" vd:"len($)<100 && regexp('^([\\w\\-\\.\\/]*|[\\w\\-\\.\\/]*:[\\w\\-\\.]+)$');  msg:'image harus alphanumeric atau simbol -,_,:,/ atau juga bisa dengan format <imagename>:<tag>'"`
}

func (m *ContainerHandler) CreateContainerAndBuildImage(ctx context.Context, c *app.RequestContext) {
	userId, _ := c.Get("userID")
	var req createServiceAndBuildImageReq
	err := c.BindForm(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
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
	svcId, createdTime, ctrLife, err := m.svc.CreateNewServiceAndUpload(ctx, &domain.Container{
		Name:        req.Name,
		CreatedTime: time.Now(),
		Image:       req.ImageName,
		Labels:      req.Labels,
		Env:         req.Env,
		Limit:       domain.Resource(req.Limit),
		Reservation: domain.Resource(req.Reservation),
		Replica:     uint64(req.Replica),
		Endpoint:    dEndpoint,
		UserID:      userId.(string),
	}, req.ImageTar, req.ImageName)
	if err != nil {

		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	resp := &createContainerResp{
		Message: "Your Container created successfully",
		Container: domain.Container{
			CreatedTime: createdTime,
			ServiceID:   svcId,
			Name:        req.Name,
			Labels:      req.Labels,
			Replica:     3,
			Limit: domain.Resource{
				CPUs:   req.Limit.CPUs,
				Memory: req.Limit.Memory,
			},
			Image:               req.ImageName,
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
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	resp, err := m.svc.GetUserContainers(ctx, userId.(string), req.Offset, req.Limit)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, getUserContainersResp{resp})
}

type getContainerReq struct {
	ID string `path:"id" vd:"len($)<400 regexp('^[\\w\\-]*$'); msg:'id hanya boleh alphanumeric dan simbol -'"`
}

type getContainerRes struct {
	Container *domain.Container `json:"container"`
}

func (m *ContainerHandler) GetContainer(ctx context.Context, c *app.RequestContext) {
	userId, _ := c.Get("userID")
	var req getContainerReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	resp, err := m.svc.GetContainer(ctx, req.ID, userId.(string))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, getContainerRes{resp})
}

func (m *ContainerHandler) StartContainer(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")
	var req getContainerReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	resp, err := m.svc.StartContainer(ctx, req.ID, userID.(string))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, getContainerRes{resp})
}

type deleteRes struct {
	Message string `json:"message"`
}

func (m *ContainerHandler) StopContainer(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")
	var req getContainerReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	err = m.svc.StopContainer(ctx, req.ID, userID.(string))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, deleteRes{Message: fmt.Sprintf("container %s successfully stopped", req.ID)})
}

func (m *ContainerHandler) DeleteContainer(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")
	var req getContainerReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	err = m.svc.DeleteContainer(ctx, req.ID, userID.(string))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, deleteRes{Message: fmt.Sprintf("container %s successfully deleted", req.ID)})
}

type updateRes struct {
	Message string `json:"message"`
}

// UpdateContainer
// @Description update container , tapi cuma update field yg ada di createServiceReq
func (m *ContainerHandler) UpdateContainer(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")
	var req createServiceReq // req body

	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	var path getContainerReq
	err = c.BindAndValidate(&path)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	ctrID, err := m.svc.UpdateContainer(ctx, &domain.Container{

		Name:        req.Name,
		Image:       req.Image,
		Labels:      req.Labels,
		Env:         req.Env,
		Limit:       req.Limit,
		Reservation: req.Reservation,
		Replica:     uint64(req.Replica),
		Endpoint:    req.Endpoint,
	}, path.ID, userID.(string))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updateRes{Message: fmt.Sprintf("container %s successfully updated", ctrID)})
}

type scaleReq struct {
	Replica uint64 `json:"replica" vd:"$<=1000 && $>=0; msg:'replica harus di antara range 0-1000'"`
}

// ScaleX
// @Description horizontal scaling sawrm service/container
func (m *ContainerHandler) ScaleX(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")
	var req getContainerReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	var reqBody scaleReq
	err = c.BindAndValidate(&reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	err = m.svc.ScaleX(ctx, userID.(string), req.ID, reqBody.Replica)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, updateRes{Message: fmt.Sprintf("container %s successfully scaled", req.ID)})
}

type HelloReq struct {
	Name string `query:"name,required"`
}

type scheduledActionReq struct {
	ID     string `path:"id" vd:"len($)<400 regexp('^[\\w\\-]*$'); msg:'id hanya boleh alphanumeric dan simbol -'"`
	UserID string `json:"user_id"`
}

// ScheduledStop -.
// @Description yg di hit dkron buat stop container (scheduled job)
func (m *ContainerHandler) ScheduledStop(ctx context.Context, c *app.RequestContext) {
	var req scheduledActionReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	err = m.svc.StopContainer(ctx, req.ID, req.UserID)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, "ok")
}

func (m *ContainerHandler) ScheduledStart(ctx context.Context, c *app.RequestContext) {
	var req scheduledActionReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	_, err = m.svc.StartContainer(ctx, req.ID, req.UserID)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, "container started")
}

type scheduleCreateServiceReq struct {
	Name        string            `json:"name,required" vd:"len($)<100 && regexp('^[\\w\\-\\.]*$'); msg:'nama harus alphanumeric atau boleh juga simbol -,_,. dan tidak boleh ada spasi'"`
	Image       string            `json:"image,required" vd:"len($)<100 && regexp('^([\\w\\-\\.\\/]*|[\\w\\-\\.\\/]*:[\\w\\-\\.]+)$'); msg:'image harus alphanumeric atau simbol -,_,:,/ atau juga bisa dengan format <imagename>:<tag>'"`
	Labels      map[string]string `json:"labels,omitempty" vd:"range($, len(#k) < 50 && len(#v) < 50) ; msg:'label haruslah kurang dari 50 '"`
	Env         []string          `json:"env" vd:"  range($, regexp('^([A-Z0-9\\_]*)=([A-Za-z0-9\\_\\/\\:\\@\\?\\(\\)\\'\\.\\=]*)$', #v) ); msg:'env harus dalam format KEY=VALUE dengan semua huruf kapital'"`
	Limit       domain.Resource   `json:"limit,required" vd:" msg:'resource limit harus anda isi '" `
	Reservation domain.Resource   `json:"reservation,omitempty" `
	Replica     int64             `json:"replica,required" vd:"$<1000 && $>=0; msg:'replica harus diantara 0-1000'"`
	Endpoint    []domain.Endpoint `json:"endpoint,required" vd:"@:len($)>0; msg:'endpoint wajib diisi'"`
	UserID      string            `json:"user_id" vd:"len($)<100 && regexp('^[\\w\\-]*$'); msg:'user_id harus alphanumeric dan simbol -/_'"`
}

func (m *ContainerHandler) ScheduleCreate(ctx context.Context, c *app.RequestContext) {
	var req scheduleCreateServiceReq
	err := c.BindAndValidate(&req)
	if err != nil {
		zap.L().Error("BindAndValidate", zap.Error(err), zap.String("err", err.Error()))

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

	_, _, _, err = m.svc.CreateNewService(ctx, &domain.Container{
		Name:        req.Name,
		CreatedTime: time.Now(),
		Image:       req.Image,
		Labels:      req.Labels,
		Env:         req.Env,
		Limit:       domain.Resource(req.Limit),
		Reservation: domain.Resource(req.Reservation),
		Replica:     uint64(req.Replica),
		Endpoint:    dEndpoint,
		UserID:      req.UserID,
	})
	if err != nil {
		zap.L().Error("CreateNewService", zap.Error(err), zap.String("err", err.Error()))

		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (m *ContainerHandler) ScheduleTerminate(ctx context.Context, c *app.RequestContext) {
	var req scheduledActionReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	err = m.svc.DeleteContainer(ctx, req.ID, req.UserID)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, "ok")
}

type scheduleContainerReq struct {
	ID            string                 `path:"id" vd:"len($)<400 regexp('^[\\w\\-]*$'); msg:'id hanya boleh alphanumeric dan simbol -'"`
	Action        domain.ContainerAction `json:"action" vd:"in($ , 'START', 'STOP', 'TERMINATE'); msg:'action harus dari pilihan berikut=START, STOPPED, TERMINATE '"`
	ScheduledTIme uint64                 `json:"scheduled_time" vd:"$<10000000 && $>0; msg:'scheduled_time harus lebih dari 0'"`
	TimeFormat    domain.TimeFormat      `json:"time_format" vd:"in($, 'MONTH', 'DAY', 'HOUR', 'MINUTE', 'SECOND')"`
}

func (m *ContainerHandler) ScheduleContainer(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")

	var req scheduleContainerReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	err = m.svc.Schedule(ctx, userID.(string), req.ID, req.ScheduledTIme, req.TimeFormat, req.Action)

	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, deleteRes{fmt.Sprintf("action %s for container %s scheduled in %d %s", req.Action, req.ID, req.ScheduledTIme, req.TimeFormat)})
}

type scheduleCreateReq struct {
	Action        domain.ContainerAction   `json:"action" vd:"$=='CREATE'; msg:'action harus dari pilihan berikut=CREATE'"`
	ScheduledTIme uint64                   `json:"scheduled_time" vd:"$<10000000 && $>0; msg:'scheduled_time harus lebih dari 0'"`
	TimeFormat    domain.TimeFormat        `json:"time_format" vd:"in($, 'MONTH', 'DAY', 'HOUR', 'MINUTE', 'SECOND')"`
	ContainerReq  scheduleCreateServiceReq `json:"container" vd:" msg:'container wajib anda isi'"`
}

func (m *ContainerHandler) CreateScheduledCreate(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("userID")

	var req scheduleCreateReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	var dEndpoint []domain.Endpoint
	for _, endp := range req.ContainerReq.Endpoint {
		dEndpoint = append(dEndpoint, domain.Endpoint{
			TargetPort:    endp.TargetPort,
			PublishedPort: endp.PublishedPort,
			Protocol:      endp.Protocol,
		})
	}

	err = m.svc.ScheduleCreate(ctx, userID.(string), req.ScheduledTIme, req.TimeFormat, req.Action, &domain.Container{
		Name:        req.ContainerReq.Name,
		CreatedTime: time.Now(),
		Image:       req.ContainerReq.Image,
		Labels:      req.ContainerReq.Labels,
		Env:         req.ContainerReq.Env,
		Limit:       domain.Resource(req.ContainerReq.Limit),
		Reservation: domain.Resource(req.ContainerReq.Reservation),
		Replica:     uint64(req.ContainerReq.Replica),
		Endpoint:    dEndpoint,
		UserID:      req.ContainerReq.UserID,
	})
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, deleteRes{fmt.Sprintf("action %s scheduled in %d %s", req.Action, req.ScheduledTIme, req.TimeFormat)})
}

func (m *ContainerHandler) GetUserContainersLoadTest(ctx context.Context, c *app.RequestContext) {
	userId, _ := c.Get("userID")

	var req Pagination
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}
	resp, err := m.svc.GetUserContainers(ctx, userId.(string), req.Offset, req.Limit)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, getUserContainersResp{resp})
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
			return http.StatusInternalServerError
		}
	}

}
