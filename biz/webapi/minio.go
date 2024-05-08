package webapi

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"dogker/lintang/container-service/config"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinioAPI struct {
	BaseURL         string
	AccessKeyID     string
	SecretAccessKey string
}

func NewMinioAPI(cfg *config.Config) *MinioAPI {
	return &MinioAPI{
		BaseURL:         cfg.Minio.BaseURL,
		AccessKeyID:     cfg.Minio.AccessKeyID,
		SecretAccessKey: cfg.Minio.SecretAccessKey,
	}
}

func (m *MinioAPI) UploadTarSourceCode(ctx context.Context, imageFile *multipart.FileHeader, imageName string) (*minio.UploadInfo, string, string, error) {
	minioClient, err := minio.New(m.BaseURL, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKeyID, m.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		zap.L().Error("new minio", zap.Error(err))
		return nil, "", "", domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}
	bucketName := "dogker-bucket"
	location := "us-east-1" // harus us-east-1
	newCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			zap.L().Debug(fmt.Sprintf("bucket %s already exists", bucketName))
		} else {
			zap.L().Error(fmt.Sprint("MakeBucket minio %s", bucketName))
			return nil, "", "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
		}
	} else {
		zap.L().Info(fmt.Sprintf("successfully created bucket %s", bucketName))
	}
	randomString := uuid.New().String()
	objectName := imageName + randomString + "_sc.tar.gz" // nama file di minio object
	// filePath := imageName
	// contentType := "application/octet-stream"
	// extennsionTar := ".tar.gz"
	fileObj, err := imageFile.Open()

	infoUpload, err := minioClient.PutObject(newCtx, bucketName, objectName, fileObj, imageFile.Size, minio.PutObjectOptions{})
	if err != nil {
		zap.L().Error("PutObject minioClient", zap.Error(err))
		return nil, "", "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	fileObj.Close()

	return &infoUpload, bucketName, objectName, nil

}

func (m *MinioAPI) GetObjectURL(ctx context.Context, objectName string) (string, error) {
	minioClient, err := minio.New(m.BaseURL, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKeyID, m.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		zap.L().Error("new minio", zap.Error(err))
		return "", domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}
	bucketName := "dogker-bucket"
	// location := "us-east-1" // harus us-east-1

	fileURl, err := minioClient.PresignedGetObject(ctx, bucketName, objectName, 10*time.Hour, url.Values{})
	return fileURl.Host + fileURl.Path, nil
}

func (m *MinioAPI) GetObject(ctx context.Context, bucketName string, objectName string) (*os.File, string, error) {
	minioClient, err := minio.New(m.BaseURL, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKeyID, m.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		zap.L().Error("new minio", zap.Error(err))
		return nil, "", domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}

	tarReader, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	defer tarReader.Close()
	randomString := uuid.New().String()
	myLocalFile, err := os.Create(bucketName + objectName + randomString)
	if err != nil {
		return nil, "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	defer myLocalFile.Close()

	stat, err := tarReader.Stat()
	if err != nil {
		zap.L().Error("tarReader.Stat", zap.Error(err))
		return nil, "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)

	}
	if _, err = io.CopyN(myLocalFile, tarReader, stat.Size); err != nil {
		zap.L().Error("CopyN io", zap.Error(err))
		return nil, "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	containerFile, err := os.Open(bucketName + objectName + randomString)
	if err != nil {
		zap.L().Error("os.Open()", zap.Error(err))
		return nil, "", domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	return containerFile, bucketName + objectName + randomString, nil
}
