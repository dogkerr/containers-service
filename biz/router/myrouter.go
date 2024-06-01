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
	RecoverContainerAfterStoppedAccidentally(ctx context.Context) error
	ContainerDown(ctx context.Context, label CommonLabels) (string, error)
}

type ContainerHandler struct {
	svc ContainerService
}

func MyRouter(r *server.Hertz, c ContainerService) {

	handler := &ContainerHandler{
		svc: c,
	}
	// Access-Control-Allow-Origin

	// taruh dibawah di sebelum protected, seteleah protected cors masih gakbisa
	root := r.Group("/api/v1") // middleware.Cors()
	{
		ctrH := root.Group("/containers")
		{
			ctrH.POST("", append(middleware.Protected(), handler.CreateContainer)...)
			ctrH.POST("/upload", append(middleware.Protected(), handler.CreateContainerAndBuildImage)...)
			ctrH.GET("", append(middleware.Protected(), handler.GetUsersContainer)...)
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

			ctrH.POST("/alert/down", handler.UserContainerDown)

			// ctrH.POST("/cron/recoverContainer", handler.RecoverContainerAfterStoppedAccidentally)
			// ctrH.POST("/cron/recoverContainer")

			// load test
			// ctrH.GET("/loadtest", append(middleware.Protected(), handler.GetUserContainersLoadTest)...)

		}
	}
}

// ResponseError model info
// @Description error message
type ResponseError struct {
	Message string `json:"message"`
}

// createServiceReq model info
// @Description request body untuk membuat container
type createServiceReq struct {
	Name        string            `json:"name,required" vd:"len($)<100 && regexp('^[\\w\\-\\.]*$'); msg:'nama harus alphanumeric atau boleh juga simbol -,_,. dan tidak boleh ada spasi'" binding:"required"`
	Image       string            `json:"image,required" vd:"len($)<100 && regexp('^([\\w\\-\\.\\/]*|[\\w\\-\\.\\/]*:[\\w\\-\\.]+)$'); msg:'image harus alphanumeric atau simbol -,_,:,/ atau juga bisa dengan format <imagename>:<tag>'" binding:"required"`
	Labels      map[string]string `json:"labels,omitempty" vd:"range($, len(#k) < 50 && len(#v) < 50) ; msg:'label haruslah kurang dari 50 '" `
	Env         []string          `json:"env" vd:"  range($, regexp('^([A-Z0-9\\_]*)=([A-Za-z0-9\\_\\/\\:\\@\\?\\(\\)\\'\\.\\=]*)$', #v) ); msg:'env harus dalam format KEY=VALUE dengan semua huruf kapital'"`
	Limit       domain.Resource   `json:"limit,required; msg:'resource limit harus anda isi '" binding:"required"`
	Reservation domain.Resource   `json:"reservation,omitempty" `
	Replica     int64             `json:"replica,required" vd:"$<1000 && $>=0; msg:'replica harus diantara 0-1000'" binding:"required"`
	Endpoint    []domain.Endpoint `json:"endpoint,required" vd:"@:len($)>0; msg:'endpoint wajib diisi'" binding:"required"`
}

// createContainerResp model info
// @Description response body endpoint membuat container
type createContainerResp struct {
	Message   string           `json:"message"`
	Container domain.Container `json:"container"`
}

// CreateContainer
// @Summary  User Membuat swarm service lewat endpoint inieperti pada postman (bearer access token saja
// @Description  User Membuat swarm service lewat endpoint ini
// @Tags containers
// @Param body body createServiceReq true "request body membuat container"
// @Accept application/json
// @Produce application/json
// @Router /containers [post]
// @Security BearerAuth
// @Success 200 {object} createContainerResp
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
			Status:              domain.ServiceRun,
			ContainerPort:       int(req.Endpoint[0].TargetPort),
			PublicPort:          int(req.Endpoint[0].PublishedPort),
			ContainerLifecycles: []domain.ContainerLifecycle{*ctrLife},
		},
	}
	c.JSON(consts.StatusOK, resp)

}

// createServiceAndBuildImageReq model info
// @Description request body untuk membuat container dan upload file  source code tarfile
type createServiceAndBuildImageReq struct {
	Name        string                `form:"name,required" vd:"len($)<100 && regexp('^[\\w\\-\\.]*$'); msg:'nama harus alphanumeric atau boleh juga simbol -,_,. dan tidak boleh ada spasi'" binding:"required"`
	Labels      map[string]string     `form:"labels,omitempty" vd:"range($, len(#k) < 5UAdOZcjE3olCYbVtYQETU9Cyy01ac40k0U1OVtGX0 && len(#v) < 50) ; msg:'label haruslah kurang dari 50 '" swaggerignore:"true"`
	Env         []string              `form:"env" vd:"  range($, regexp('^([A-Z0-9\\_]*)=([A-Za-z0-9\\_\\/\\:\\@\\?\\(\\)\\'\\.\\=]*)$', #v) ); msg:'env harus dalam format KEY=VALUE dengan semua huruf kapital'" `
	Limit       domain.Resource       `form:"limit,required" binding:"required" swaggerignore:"true"`
	Reservation domain.Resource       `form:"reservation,omitempty"  swaggerignore:"true"`
	Replica     int64                 `form:"replica,required" vd:"$<1000 && $>=0; msg:'replica harus diantara 0-1000'" binding:"required"`
	Endpoint    []domain.Endpoint     `form:"endpoint,required" vd:"@:len($)>0; msg:'endpoint wajib diisi'" binding:"required"`
	ImageTar    *multipart.FileHeader `form:"image,required" vd:"msg:'endpoint wajib diisi'" binding:"required" swaggerignore:"true"`
	ImageName   string                `form:"imageName,required" vd:"len($)<100 && regexp('^([\\w\\-\\.\\/]*|[\\w\\-\\.\\/]*:[\\w\\-\\.]+)$');  msg:'image harus alphanumeric atau simbol -,_,:,/ atau juga bisa dengan format <imagename>:<tag>'" binding:"required"`
}

// CreateContainerAndUpload
// @Summary  User Membuat swarm service tetapi source code (tarfile) nya dia upload ,lewat endpoint inieperti pada postman (bearer access token saja
// @Description  User Membuat swarm service tetapi source code (tarfile) nya dia upload  lewat endpoint ini
// @Tags containers
// @Param body formData createServiceAndBuildImageReq true "request body membuat container dengan tarfile source code "
// @Accept multipart/form-data
// @Produce application/json
// @Router /containers/upload [post]
// @Security BearerAuth
// @Success 200 {object} createContainerResp
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
			Status:              domain.ServiceRun,
			ContainerPort:       int(req.Endpoint[0].TargetPort),
			PublicPort:          int(req.Endpoint[0].PublishedPort),
			ContainerLifecycles: []domain.ContainerLifecycle{*ctrLife},
		},
	}

	c.JSON(consts.StatusOK, resp)

}

type Pagination struct {
	Offset uint64 `query:"page"`
	Limit  uint64 `query:"limit" vd:"$<100; msg:'limit tidak boleh lebih dari 100'"`
}

// getUserContainersResp model info
// @Description response GetUsersContainer
type getUserContainersResp struct {
	Containers *[]domain.Container `json:"containers"`
}

// GetUsersContainer
// @Summary Mendapatkan semua swarm service milik user
// @Description  Mendapatkan semua swarm service milik user
// @Tags containers
// @Produce application/json
// @Router /containers [get]
// @Security BearerAuth
// @Success 200 {object} getUserContainersResp
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, getUserContainersResp{resp})

}

type getContainerReq struct {
	ID string `path:"id" vd:"len($)<400 regexp('^[\\w\\-]*$'); msg:'id hanya boleh alphanumeric dan simbol -'"`
}

// getContainerRes model info
// @Description mendapatkan container user berdasarkan id container
type getContainerRes struct {
	Container *domain.Container `json:"container"`
}

// GetContainer
// @Summary Mendapatkan swarm service user berdasarkan id
// @Description  Mendapatkan swarm service user berdasarkan id
// @Param id path string true "container id"
// @Tags containers
// @Produce application/json
// @Router /containers/{id} [get]
// @Security BearerAuth
// @Success 200 {object} getContainerRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, getContainerRes{resp})

}

// StartContainer
// @Summary run container user
// @Description run container user
// @Param id path string true "container id"
// @Tags containers
// @Produce application/json
// @Router /containers/{id}/start [post]
// @Security BearerAuth
// @Success 200 {object} getContainerRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, getContainerRes{resp})

}

// deleteRes model info
// @Description response body yg isinnya message success doang
type deleteRes struct {
	Message string `json:"message"`
}

// StopContainer
// @Summary stop container user
// @Description stop container user
// @Param id path string true "container id"
// @Tags containers
// @Produce application/json
// @Router /containers/{id}/stop [post]
// @Security BearerAuth
// @Success 200 {object} deleteRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, deleteRes{Message: fmt.Sprintf("container %s successfully stopped", req.ID)})

}

// DeleteContainer
// @Summary delete user swarm service
// @Description delete user swarm service
// @Param id path string true "container id"
// @Tags containers
// @Produce application/json
// @Router /containers/{id} [delete]
// @Security BearerAuth
// @Success 200 {object} deleteRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, deleteRes{Message: fmt.Sprintf("container %s successfully deleted", req.ID)})

}

// updateRes model info
// @Description response body isinya message success doang
type updateRes struct {
	Message string `json:"message"`
}

// UpdateContainer
// @Summary update swarm service user (bisa juga vertical scaling disini)
// @Description update swarm service user (bisa juga vertical scaling disini)
// @Param id path string true "container id"
// @Param body body createServiceReq true "request body update container"
// @Tags containers
// @Accept application/json
// @Produce application/json
// @Router /containers/{id} [put]
// @Security BearerAuth
// @Success 200 {object} updateRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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

	c.JSON(consts.StatusOK, updateRes{Message: fmt.Sprintf("container %s successfully updated", ctrID)})

}

// scaleReq model info
// @Description request body horizontal scaling
type scaleReq struct {
	Replica uint64 `json:"replica" vd:"$<=1000 && $>=0; msg:'replica harus di antara range 0-1000'"`
}

// ScaleX
// @Summary horizontal scaling container user
// @Description horizontal scaling container user
// @Param id path string true "container id"
// @Param body body scaleReq true "request body horizontal scaling"
// @Tags containers
// @Produce application/json
// @Router /containers/{id}/scale [put]
// @Security BearerAuth
// @Success 200 {object} updateRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, updateRes{Message: fmt.Sprintf("container %s successfully scaled", req.ID)})

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

	c.JSON(consts.StatusOK, "ok")

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
	c.JSON(consts.StatusOK, "container started")

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
	c.JSON(consts.StatusOK, "ok")

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

	c.JSON(consts.StatusOK, "ok")

}

// scheduleContainerReq model info
// @Description request body menjadwalkan start/stop/terminate container
type scheduleContainerReq struct {
	ID            string                 `path:"id" vd:"len($)<400 regexp('^[\\w\\-]*$'); msg:'id hanya boleh alphanumeric dan simbol -'" binding:"required"`
	Action        domain.ContainerAction `json:"action" vd:"in($ , 'START', 'STOP', 'TERMINATE'); msg:'action harus dari pilihan berikut=START, STOPPED, TERMINATE '"  binding:"required"`
	ScheduledTIme uint64                 `json:"scheduled_time" vd:"$<10000000 && $>0; msg:'scheduled_time harus lebih dari 0'"  binding:"required"`
	TimeFormat    domain.TimeFormat      `json:"time_format" vd:"in($, 'MONTH', 'DAY', 'HOUR', 'MINUTE', 'SECOND')"  binding:"required"`
}

// ScheduleContainer
// @Summary menjadwalkan start/stop/terminate container
// @Description menjadwalkan start/stop/terminate container
// @Param id path string true "container id"
// @Param body body scheduleContainerReq true "request body penjadwalan start/stop/terminate container"
// @Tags containers
// @Accept application/json
// @Produce application/json
// @Router /containers/{id}/schedule [post]
// @Security BearerAuth
// @Success 200 {object} deleteRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, deleteRes{fmt.Sprintf("action %s for container %s scheduled in %d %s", req.Action, req.ID, req.ScheduledTIme, req.TimeFormat)})

}

// scheduleCreateReq model info
// @Description request body penjadwalan pembuatan container
type scheduleCreateReq struct {
	Action        domain.ContainerAction   `json:"action" vd:"$=='CREATE'; msg:'action harus dari pilihan berikut=CREATE'" binding:"required"`
	ScheduledTIme uint64                   `json:"scheduled_time" vd:"$<10000000 && $>0; msg:'scheduled_time harus lebih dari 0'" binding:"required"`
	TimeFormat    domain.TimeFormat        `json:"time_format" vd:"in($, 'MONTH', 'DAY', 'HOUR', 'MINUTE', 'SECOND')" binding:"required"`
	ContainerReq  scheduleCreateServiceReq `json:"container" vd:" msg:'container wajib anda isi'" binding:"required"`
}

// CreateScheduledCreate
// @Summary menjadwalkan pembuatan container
// @Description menjadwalkan pembuatan container
// @Param body body scheduleCreateReq true "request body penjadwalan pembuatan container"
// @Tags containers
// @Accept application/json
// @Produce application/json
// @Router /containers/create/schedule [post]
// @Security BearerAuth
// @Success 200 {object} deleteRes
// @failure 400 {object} ResponseError
// @failure 500 {object} ResponseError
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
	c.JSON(consts.StatusOK, deleteRes{fmt.Sprintf("action %s scheduled in %d %s", req.Action, req.ScheduledTIme, req.TimeFormat)})

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
	c.JSON(consts.StatusOK, getUserContainersResp{resp})

}

// ----- user container down -----
type CommonLabels struct {
	Alertname                       string `json:"alertname"`
	ContainerSwarmServiceID         string `json:"container_label_com_docker_swarm_service_id"`
	ContainerDockerSwarmServiceName string `json:"container_label_com_docker_swarm_service_name"`
	ContainerLabelUserID            string `json:"container_label_user_id"`
}

type PromeWebhookReq struct {
	Receiver     string       `json:"receiver"`
	CommonLabels CommonLabels `json:"commonLabels"`
}

type promeWebhookRes struct {
	Message string `json:"message"`
}

func (m *ContainerHandler) UserContainerDown(ctx context.Context, c *app.RequestContext) {

	var req PromeWebhookReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
		return
	}

	msg, err := m.svc.ContainerDown(ctx, req.CommonLabels)

	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}
	c.JSON(consts.StatusOK, promeWebhookRes{msg})
}

type cronTerminatedAccidentallyReq struct {
	ServiceIDs []string `json:"service_ids"`
}

type messageRes struct {
	Message string `json:"message"`
}

// func (m *ContainerHandler) RecoverContainerAfterStoppedAccidentally(ctx context.Context, c *app.RequestContext) {
// 	err := m.svc.RecoverContainerAfterStoppedAccidentally(ctx)
// 	if err != nil {
// 		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
// 		return
// 	}
// 	c.JSON(consts.StatusOK, messageRes{"ok container recovered"})
// }

// // TerminatedAccidentally
// // @Desc ini dipanggil sama monitorservice setiap 4 detik ketika ada container mati > 2 detik
// // terus metrics dari ctr mati tsb di get dari monitor-service, terus metricsnya bakal diinsert
// // set container status terminated && container lifecycle stopped
// // kalau sebelumnya terminated di ctrnya, berarti gak usah di insert metricsnya lagi karena emang pernah diinsert pas deleteContainer
// // ke tabel container metrics , jadi container metrics itu nyimpen metrics container yang udah mati
// // karena kalo container mati
// func (m *ContainerHandler) TerminatedAccidentally(ctx context.Context, c *app.RequestContext) {
// 	var req cronTerminatedAccidentallyReq
// 	err := c.BindAndValidate(&req)
// 	if err != nil {
// 		c.String(consts.StatusBadRequest, err.Error())
// 		return
// 	}
// 	if len(req.ServiceIDs) != 0 {

// 		// kalau emang ada
// 	}
// 	c.JSON(consts.StatusOK, "ok")
// }

func (m *ContainerHandler) SayHello(ctx context.Context, c *app.RequestContext) {
	var req HelloReq

	err := c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(hello.HelloResp)
	resp.RespBody = "halo " + req.Name
	c.JSON(consts.StatusOK, resp)

}

func getStatusCode(err error) int {
	if err == nil {
		return consts.StatusOK
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
