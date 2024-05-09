package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type ContainerRepository interface {
	Get(ctx context.Context, serviceID string) (*domain.Container, error)
	GetAllUserContainers(ctx context.Context, userID string) (*[]domain.Container, error)
	Insert(ctx context.Context, c *domain.Container) (*domain.Container, error)
	Update(ctx context.Context, c *domain.Container) error
	Delete(ctx context.Context, serviceID string) error
	InsertLifecycle(ctx context.Context, c *domain.ContainerLifecycle) (*domain.ContainerLifecycle, error)
	GetLifecycle(ctx context.Context, lifeId string) (*domain.ContainerLifecycle, error)
	UpdateLifecycle(ctx context.Context, lifeId string, stopTime time.Time, status domain.ContainerStatus, replica uint32) error
	UpdateCtrLifeCycleWithoutStopTime(ctx context.Context, replica uint64, lifeID string) error
	InsertContainerMetrics(ctx context.Context, metrics domain.Metric) error
}

type DockerEngineAPI interface {
	CreateService(ctx context.Context, c *domain.Container) (string, error)
	GetAllUserContainers(ctx context.Context, userID string, cDB []domain.Container) (*[]domain.Container, error)
	Get(ctx context.Context, ctrID string, cDB *domain.Container) (*domain.Container, error)
	Start(ctx context.Context, ctrID string, lastReplicaFromDB uint64, userID string, cDB *domain.Container) (*domain.Container, error)
	Stop(ctx context.Context, ctrID string, userID string, cDB *domain.Container) error
	Delete(ctx context.Context, ctrID string) error
	Update(ctx context.Context, ctrID string, c *domain.Container, userID string) (err error)
	ScaleX(ctx context.Context, ctrID string, replica uint64, userID string) error
	IsPublicPortAndNameAvailable(ctx context.Context, wantedPorts []uint32, name string) error
	BuildImageFromFile(ctx context.Context, file *os.File, imageName string) (types.ImageBuildResponse, error)
}

type DkronAPI interface {
	AddJob(ctx context.Context, schedule uint64, ctrID string, action domain.ContainerAction, userID string) error
	AddCreateJob(ctx context.Context, schedule uint64, action domain.ContainerAction, userID string, ctr *domain.Container) error
}

type MonitorClient interface {
	GetSpecificContainerMetrics(ctx context.Context, ctrID string, userID string, serviceStartTime time.Time) (*domain.Metric, error)
}

type MinioAPI interface {
	UploadTarSourceCode(ctx context.Context, imageFile *multipart.FileHeader, imageName string) (*minio.UploadInfo, string, string, error)
	GetObject(ctx context.Context, bucketName string, objectName string) (*os.File, string, error)
}

type ContainerService struct {
	containerRepo ContainerRepository
	dockerAPI     DockerEngineAPI
	dkronAPI      DkronAPI
	monitorClient MonitorClient
	minioAPI      MinioAPI
}

func NewContainerService(c ContainerRepository, d DockerEngineAPI, dkron DkronAPI, monitorSvc MonitorClient,
	minioAPI MinioAPI) *ContainerService {
	return &ContainerService{
		containerRepo: c,
		dockerAPI:     d,
		dkronAPI:      dkron,
		monitorClient: monitorSvc,
		minioAPI:      minioAPI,
	}
}

func (s *ContainerService) Hello(ctx context.Context) (string, error) {
	return "hello", nil
}

func (s *ContainerService) CreateNewService(ctx context.Context, d *domain.Container) (string, time.Time, *domain.ContainerLifecycle, error) {
	serviceId, err := s.dockerAPI.CreateService(ctx, d)
	if err != nil {
		zap.L().Error("s.dockerAPI.CreateService", zap.Error(err))
		return "", time.Now(), nil, err
	}

	d.ServiceID = serviceId
	d.CreatedTime = time.Now()

	ctrRowId, err := s.containerRepo.Insert(ctx, d)
	if err != nil {
		return "", time.Now(), nil, err
	}

	ctrLife, err := s.containerRepo.InsertLifecycle(ctx, &domain.ContainerLifecycle{
		ContainerID: ctrRowId.ID,
		StartTime:   d.CreatedTime,
		Status:      domain.ContainerStatusRUN,
		Replica:     d.Replica,
	})
	if err != nil {
		return "", time.Now(), nil, err
	}
	return serviceId, d.CreatedTime, ctrLife, nil
}

func (s *ContainerService) CreateNewServiceAndUpload(ctx context.Context, d *domain.Container, imageFile *multipart.FileHeader,
	imageName string) (string, time.Time, *domain.ContainerLifecycle, error) {

	// upload image ke minio && build image

	_, bucketName, objectName, err := s.minioAPI.UploadTarSourceCode(ctx, imageFile, imageName)
	if err != nil {
		zap.L().Error("UploadTarSourceCode minio", zap.Error(err))
		return "", time.Now(), nil, err
	}

	imageFileMinio, fileName, err := s.minioAPI.GetObject(ctx, bucketName, objectName)
	if err != nil {
		return "", time.Now(), nil, err
	}

	buildRes, err := s.dockerAPI.BuildImageFromFile(ctx, imageFileMinio, imageName)
	if err != nil {
		return "", time.Now(), nil, err
	}
	defer imageFileMinio.Close()
	defer buildRes.Body.Close()

	os.Remove(fileName)

	// setelah build image , create swarm service
	d.Image = imageName
	serviceId, err := s.dockerAPI.CreateService(ctx, d)
	if err != nil {
		zap.L().Error("s.dockerAPI.CreateService", zap.Error(err))
		return "", time.Now(), nil, err
	}

	d.ServiceID = serviceId
	d.CreatedTime = time.Now()

	ctrRowId, err := s.containerRepo.Insert(ctx, d)
	if err != nil {
		return "", time.Now(), nil, err
	}

	ctrLife, err := s.containerRepo.InsertLifecycle(ctx, &domain.ContainerLifecycle{
		ContainerID: ctrRowId.ID,
		StartTime:   d.CreatedTime,
		Status:      domain.ContainerStatusRUN,
		Replica:     d.Replica,
	})
	if err != nil {
		return "", time.Now(), nil, err
	}

	return serviceId, d.CreatedTime, ctrLife, nil
}

// GetUserContainers -.
// @Description get semua container milik user , tapi ini yg masih run sebagai swarm service doang jadi masih salah
func (s *ContainerService) GetUserContainers(ctx context.Context, userID string, offset uint64, limit uint64) (*[]domain.Container, error) {
	// get all user container di repo
	userCtrsDb, err := s.containerRepo.GetAllUserContainers(ctx, userID)
	if err != nil {
		return nil, err
	}

	// mendapatkan semua container milik user di docker api
	ctrsDocker, err := s.dockerAPI.GetAllUserContainers(ctx, userID, *userCtrsDb)
	if err != nil {
		return nil, err
	}

	// append ctr yg status nya terminated, karena di dockerapi ga kedeteck sama sekali
	var ctrDockerSet map[string]struct{} = make(map[string]struct{}) // set ctr list yg dikasih dockerapi
	for _, ctr := range *ctrsDocker {
		ctrDockerSet[ctr.ServiceID] = struct{}{}
	}

	for _, ctrDB := range *userCtrsDb {
		if _, ok := ctrDockerSet[ctrDB.ServiceID]; !ok {
			// kalo ctr terminated gak ada di list ctr docker api
			*ctrsDocker = append(*ctrsDocker, ctrDB)
		}
	}

	return ctrsDocker, nil
}

// GetUserContainersLoadTest -.
// @Description ini cuma buat coba load testing doang hehe
func (s *ContainerService) GetUserContainersLoadTest(ctx context.Context, userID string, offset uint64, limit uint64) (*[]domain.Container, error) {
	// get all user container di repo
	userCtrsDb, err := s.containerRepo.GetAllUserContainers(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return userCtrsDb, nil
}

// GetContainer -.
// @Description get container by id, tapi ini yg masih run sebagai swarm service doang jadi masih salah
func (s *ContainerService) GetContainer(ctx context.Context, ctrID string, userID string) (*domain.Container, error) {
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return nil, err
	}
	if ctrDB.UserID != userID {
		return nil, domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}
	ctrDocker, err := s.dockerAPI.Get(ctx, ctrID, ctrDB)
	if err != nil {
		return nil, err
	}

	if ctrDocker == nil {
		// kalo status container terminated, di docker api gak kedeteck , jadi ambil dari repo
		ctrDocker = ctrDB // tapi fieldnya gak lengkapp
	}
	return ctrDocker, nil
}

func (s *ContainerService) StartContainer(ctx context.Context, ctrID string, userID string) (*domain.Container, error) {
	// get ctr dari db
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return nil, err
	}
	if ctrDB.UserID != userID {
		return nil, domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}

	// get lastReplica dari tabel ctrlifecycles
	lifecycles := ctrDB.ContainerLifecycles
	lastReplicaFromDB := qSortWaktu(lifecycles).Replica

	// start container
	ctr, err := s.dockerAPI.Start(ctx, ctrID, lastReplicaFromDB, userID, ctrDB)
	if err != nil {
		zap.L().Error("Start s.dockerAPI", zap.Error(err), zap.String("ctrID", ctrID), zap.String("userID", userID))
		return nil, err
	}

	// save to tabel ctrlifecycles
	newLife, err := s.containerRepo.InsertLifecycle(ctx, &domain.ContainerLifecycle{
		ContainerID: ctrDB.ID,
		StartTime:   time.Now(),
		Replica:     ctr.Replica,
		Status:      domain.ContainerStatusRUN,
	})
	if err != nil {
		return nil, err
	}

	// update container status to run
	ctrDB.Status = domain.ContainerStatusRUN
	err = s.containerRepo.Update(ctx, ctrDB)
	if err != nil {
		return nil, err
	}

	ctr.ContainerLifecycles = append(ctr.ContainerLifecycles, *newLife)

	return ctr, nil
}

func (s *ContainerService) StopContainer(ctx context.Context, ctrID string, userID string) error {
	// get ctr dari db
	// cek apakah user yg punya containernya
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return err
	}
	if ctrDB.UserID != userID {
		return domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}

	// catat metrics sebelum di stop, biar di prometheus gak fetch metrics cpu & memorynya = 0 & metrics cpu dan  memory =  0 gak dikirim ke billing-service
	// get last container metrics from monitoservice
	// ini harus dilakuin sebelum stop docker service, biar cpunya gak kedeteck 0
	metric, err := s.monitorClient.GetSpecificContainerMetrics(ctx, ctrID, userID, ctrDB.CreatedTime)
	if err != nil {
		return err
	}

	// insert last metrics ino container_metrics table
	metric.ContainerID = ctrDB.ID
	err = s.containerRepo.InsertContainerMetrics(ctx, *metric)
	if err != nil {
		return err
	}

	// stop container
	err = s.dockerAPI.Stop(ctx, ctrID, userID, ctrDB)
	if err != nil {
		return err
	}

	// get lastLifecycleID dari tabel ctrlifecycles
	lifecycles := ctrDB.ContainerLifecycles
	lastLifecycleID := qSortWaktu(lifecycles).ID

	// update current lifecycles
	err = s.containerRepo.UpdateLifecycle(ctx, lastLifecycleID, time.Now(), domain.ContainerStatusSTOPPED, uint32(ctrDB.Replica))
	if err != nil {
		return err
	}

	// update row di table container jd stop statusnya
	ctrDB.Status = domain.ContainerStatusSTOPPED
	err = s.containerRepo.Update(ctx, ctrDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *ContainerService) DeleteContainer(ctx context.Context, ctrID string, userID string) error {
	// get ctr dari db
	// cek apakah user yg punya containernya
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return err
	}
	if ctrDB.UserID != userID {
		return domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}

	// get last lifecycleID , berdasaarkan starttime terbaru
	lifecycles := ctrDB.ContainerLifecycles
	lifeCycleStatus := qSortWaktu(lifecycles).Status

	// insert metrics jika status container sebelumya tidak stop, biar si billing service gak dapet metrics 0 cpu & memory
	if lifeCycleStatus != domain.ContainerStatusSTOPPED {
		// get last container metrics from monitoservice
		metric, err := s.monitorClient.GetSpecificContainerMetrics(ctx, ctrID, userID, ctrDB.CreatedTime)
		if err != nil {
			return err
		}

		// insert last metrics ino container_metrics table
		metric.ContainerID = ctrDB.ID
		err = s.containerRepo.InsertContainerMetrics(ctx, *metric)
		if err != nil {
			return err
		}
	}

	// delete container
	err = s.dockerAPI.Delete(ctx, ctrID)
	if err != nil {
		return err
	}

	// update terminatedTime di tabel containers
	ctrDB.TerminatedTime = time.Now()
	ctrDB.Status = domain.ContainerStatusSTOPPED
	err = s.containerRepo.Update(ctx, ctrDB)
	if err != nil {
		return err
	}

	// update status ctrlifecycle jadi stop
	lifeCycleID := qSortWaktu(ctrDB.ContainerLifecycles).ID

	err = s.containerRepo.UpdateLifecycle(ctx, lifeCycleID, time.Now(), domain.ContainerStatusSTOPPED, uint32(ctrDB.Replica))
	if err != nil {
		return err
	}

	// update status container jd stop di tabel container
	ctrDB.Status = domain.ContainerStatusSTOPPED
	err = s.containerRepo.Update(ctx, ctrDB)
	return nil
}

func (s *ContainerService) UpdateContainer(ctx context.Context, d *domain.Container, ctrID string, userID string) (string, error) {
	// get ctr dari db
	// cek apakah user yg punya containernya
	// sekalian cek apakah container dg ctrID ada
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return "", err
	}
	if ctrDB.UserID != userID {
		return "", domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}

	// cek apakah public port dan nama container baru yang diinginkan user tersedia
	var endpointDBs map[uint64]struct{} = make(map[uint64]struct{})
	for _, endpointDB := range ctrDB.Endpoint {
		endpointDBs[endpointDB.PublishedPort] = struct{}{}
	}

	var wantedPorts []uint32
	for _, endpoint := range d.Endpoint {
		if _, ok := endpointDBs[endpoint.PublishedPort]; !ok {
			wantedPorts = append(wantedPorts, uint32(endpoint.PublishedPort)) // hanya append publicport baru yang beda sama public port container sebelumnya
		}
	}

	err = s.dockerAPI.IsPublicPortAndNameAvailable(ctx, wantedPorts, d.Name)
	if err != nil {
		return "", err
	}

	// update container di docker
	err = s.dockerAPI.Update(ctx, ctrID, d, userID)
	if err != nil {
		return "", err
	}

	// update container di db
	d.Status = ctrDB.Status // make sure status ctr gak diubah sama user
	d.CreatedTime = ctrDB.CreatedTime
	d.ContainerPort = int(d.Endpoint[0].TargetPort)
	d.PublicPort = int(d.Endpoint[0].PublishedPort)
	err = s.containerRepo.Update(ctx, d)
	if err != nil {
		return "", err
	}
	return ctrDB.ID, nil
}

// ScaleX -.
// @Description horizontal scaling
func (s *ContainerService) ScaleX(ctx context.Context, userID string, ctrID string, replica uint64) error {
	// get ctr dari db
	// cek apakah user yg punya containernya
	// sekalian cek apakah container dg ctrID ada
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return err
	}
	if ctrDB.UserID != userID {
		return domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}

	// horizontal scaling swarm service
	err = s.dockerAPI.ScaleX(ctx, ctrID, replica, userID)
	if err != nil {
		return err
	}

	// update field replica di tabel lifecycle
	// get last lifecycleID , berdasaarkan starttime terbaru
	lifecycles := ctrDB.ContainerLifecycles
	lifeCycleID := qSortWaktu(lifecycles).ID
	err = s.containerRepo.UpdateCtrLifeCycleWithoutStopTime(ctx, replica, lifeCycleID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ContainerService) Schedule(ctx context.Context, userID string, ctrID string, scheduledTime uint64, timeFormat domain.TimeFormat, action domain.ContainerAction) error {
	// get ctr dari db
	// cek apakah user yg punya containernya
	// sekalian cek apakah container dg ctrID ada
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return err
	}
	if ctrDB.UserID != userID {
		zap.L().Debug("user bukan pemilik container")
		return domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container %s bukan milik anda", ctrID))
	}

	// convert scheduledTime to second format
	var scheduledTimeSecond uint64

	if timeFormat == domain.Day {
		scheduledTimeSecond = 86400 * scheduledTime
	} else if timeFormat == domain.Hour {
		scheduledTimeSecond = 3600 * scheduledTime
	} else if timeFormat == domain.Second {
		scheduledTimeSecond = scheduledTime
	} else if timeFormat == domain.Month {
		// 86400 * 30 = detik adalam 1 bulan
		scheduledTimeSecond = 86400 * 30 * scheduledTime
	}

	// create cron job di dkron
	err = s.dkronAPI.AddJob(ctx, scheduledTimeSecond, ctrID, action, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *ContainerService) ScheduleCreate(ctx context.Context, userID string, scheduledTime uint64, timeFormat domain.TimeFormat, action domain.ContainerAction, ctr *domain.Container) error {
	// convert scheduledTime to second format
	var scheduledTimeSecond uint64

	if timeFormat == domain.Day {
		scheduledTimeSecond = 86400 * scheduledTime
	} else if timeFormat == domain.Hour {
		scheduledTimeSecond = 3600 * scheduledTime
	} else if timeFormat == domain.Second {
		scheduledTimeSecond = scheduledTime
	} else if timeFormat == domain.Month {
		// 86400 * 30 = detik adalam 1 bulan
		scheduledTimeSecond = 86400 * 30 * scheduledTime
	}

	// cek apakah public port yang diinginkan user tersedia
	var wantedPorts []uint32
	for _, endpoint := range ctr.Endpoint {
		wantedPorts = append(wantedPorts, uint32(endpoint.PublishedPort))
	}
	err := s.dockerAPI.IsPublicPortAndNameAvailable(ctx, wantedPorts, ctr.Name)
	if err != nil {
		return err
	}

	// bikin cron job di dkron
	err = s.dkronAPI.AddCreateJob(ctx, scheduledTimeSecond, action, userID, ctr)
	if err != nil {
		zap.L().Error("AddCreateJob dkron", zap.String("ctrID", ctr.ID), zap.String("userID", userID))
		return err
	}
	return nil
}

func qSortWaktu(arr []domain.ContainerLifecycle) domain.ContainerLifecycle {
	var recurse func(left int, right int)
	var partition func(left int, right int, pivot int) int

	partition = func(left int, right int, pivot int) int {
		v := arr[pivot]
		right--
		arr[pivot], arr[right] = arr[right], arr[pivot]

		for i := left; i < right; i++ {
			if arr[i].StartTime.Unix() <= v.StartTime.Unix() {
				arr[i], arr[left] = arr[left], arr[i]
				left++
			}
		}

		arr[left], arr[right] = arr[right], arr[left]
		return left
	}

	recurse = func(left int, right int) {
		if left < right {
			pivot := (right + left) / 2
			pivot = partition(left, right, pivot)
			recurse(left, pivot)
			recurse(pivot+1, right)
		}
	}

	recurse(0, len(arr))
	return arr[len(arr)-1]
}
