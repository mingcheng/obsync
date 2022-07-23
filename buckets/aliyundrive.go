package buckets

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/mingcheng/aliyundrive"
	"github.com/mingcheng/aliyundrive/store"
	"github.com/mingcheng/obsync"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"sync"
)

var filePutLock sync.Mutex

type AliyunDrive struct {
	Config         *obsync.BucketConfig
	client         *aliyundrive.AliyunDrive
	DefaultDriveID string
}

// refreshToken to refresh the storage token from specified keys
func (t *AliyunDrive) refreshToken(ctx context.Context) error {
	resp, err := t.client.RefreshToken(ctx, &aliyundrive.RefreshTokenReq{
		RefreshToken: t.Config.Key,
	})

	if err != nil {
		return err
	}

	log.Tracef("mark default drive id is %s", resp.DefaultDriveID)
	t.DefaultDriveID = resp.DefaultDriveID

	return nil
}

func (t *AliyunDrive) Info(ctx context.Context) (interface{}, error) {
	return "This is a test buckets", nil
}

func (t *AliyunDrive) Exists(ctx context.Context, path string) bool {

	//info, err := t.client.Get(context.TODO(), &aliyundrive.GetFileReq{
	//	DriveID: aliyundrive.DefaultDriveID,
	//	FileID:  result.FileID,
	//})

	// @TODO
	return false
}

func (t *AliyunDrive) Put(ctx context.Context, localFile, key string) error {
	filePutLock.Lock()
	defer filePutLock.Unlock()

	createFolderReq := &aliyundrive.CreateFolderReq{
		DriveID:       t.DefaultDriveID,
		ParentFileID:  aliyundrive.RootFileID,
		CheckNameMode: aliyundrive.ModeRefuse,
		Type:          aliyundrive.TypeFolder,
		Name:          path.Dir(key),
	}

	resp, err := t.client.CreateFolder(ctx, createFolderReq)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Tracef("create remote folder %s is successful, file id is %s", resp.FileName, resp.FileID)

	log.Tracef("start upload local file %s to %s", localFile, key)
	_, err = t.client.UploadFile(context.Background(), &aliyundrive.UploadFileReq{
		DriveID:       t.DefaultDriveID,
		ParentID:      resp.FileID,
		FilePath:      localFile,
		CheckNameMode: aliyundrive.ModeRefuse,
		Name:          path.Base(key),
	})

	if err != nil {
		log.Error(err)
		return err
	}

	log.Tracef("upload file %s to %s is finished, bye~", localFile, key)
	return nil
}

func init() {
	// instance a redis connection from environment variable
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	status := redisClient.Ping(context.Background())
	if status.Err() != nil {
		log.Warnf("ping redis error: %v", status.Err())
	}

	_ = obsync.RegisterBucketClientFunc("aliyundrive",
		func(config obsync.BucketConfig) (obsync.BucketClient, error) {
			storage := store.RedisStore{
				Client: redisClient,
			}

			cmd := redisClient.Get(context.Background(), aliyundrive.KeyRefreshToken)
			if err, refreshToken := cmd.Err(), cmd.Val(); err == nil && refreshToken != "" {
				log.Debugf("get refresh token from redis: %v", refreshToken)
				config.Key = refreshToken
			}

			if config.Key == "" {
				return nil, errors.New("refresh token is empty")
			}

			drive := AliyunDrive{
				Config: &config,
				client: aliyundrive.New(aliyundrive.WithStore(&storage)),
			}

			err := drive.refreshToken(context.Background())
			if err != nil {
				log.Error(err)
				return nil, err
			}

			return &drive, nil
		})
}
