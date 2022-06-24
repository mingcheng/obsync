package bucket

import (
	"context"
	"github.com/mingcheng/aliyundrive"
	"github.com/mingcheng/aliyundrive/store"
	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
	"log"
	"path"
	"sync"
	"time"
)

type AliyunDrive struct {
	Config         bucket.Config
	client         *aliyundrive.AliyunDrive
	DefaultDriveID string
	ticker         *time.Ticker
	done           chan bool
	uploadLock     sync.Mutex
}

func (r *AliyunDrive) refreshToken(ctx context.Context) error {
	resp, err := r.client.RefreshToken(ctx, &aliyundrive.RefreshTokenReq{
		RefreshToken: r.Config.Key,
	})

	if err != nil {
		return err
	}

	r.DefaultDriveID = resp.DefaultDriveID
	return nil
}

func (r *AliyunDrive) OnStart(ctx context.Context) error {
	if err := r.refreshToken(ctx); err != nil {
		return err
	}

	r.ticker = time.NewTicker(time.Hour)
	r.done = make(chan bool)

	go func() {
		for {
			select {
			case <-r.done:
				return
			case <-r.ticker.C:
				err := r.refreshToken(context.Background())
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	return nil
}

func (r *AliyunDrive) OnStop(ctx context.Context) error {
	r.ticker.Stop()
	r.done <- true
	return nil
}

func (t *AliyunDrive) Info() (interface{}, error) {
	return "This is a test bucket", nil
}

func (t *AliyunDrive) Exists(path string) bool {
	return false
}

func (t *AliyunDrive) Put(task obsync.Task) error {
	t.uploadLock.Lock()
	defer t.uploadLock.Unlock()

	client := t.client

	pathName := path.Join(task.SubDir, task.Key)
	folderName := path.Dir(pathName)
	fileName := path.Base(pathName)

	createFolderReq := &aliyundrive.CreateFolderReq{
		DriveID:       t.DefaultDriveID,
		ParentFileID:  aliyundrive.RootFileID,
		CheckNameMode: aliyundrive.ModeRefuse,
		Type:          aliyundrive.TypeFolder,
		Name:          folderName,
	}

	resp, err := client.CreateFolder(context.Background(), createFolderReq)
	if err != nil {
		return err
	}

	_, err = client.UploadFile(context.Background(), &aliyundrive.UploadFileReq{
		DriveID:       t.DefaultDriveID,
		ParentID:      resp.FileID,
		FilePath:      task.Local,
		CheckNameMode: aliyundrive.ModeRefuse,
		Name:          fileName,
	})

	if err != nil {
		return err
	}

	return nil
}

func init() {
	bucket.Register("aliyundrive", func(config bucket.Config) (bucket.Bucket, error) {

		return &AliyunDrive{
			Config: config,
			client: aliyundrive.New(aliyundrive.WithStore(store.NewMemoryStore())),
		}, nil
	})
}
