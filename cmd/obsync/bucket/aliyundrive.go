package bucket

import (
	"context"
	"log"
	"path"
	"sync"
	"time"

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
	uploadMutex    sync.Mutex
}

func (t *AliyunDrive) refreshToken(ctx context.Context) error {
	resp, err := t.client.RefreshToken(ctx, &aliyundrive.RefreshTokenReq{
		RefreshToken: t.Config.Key,
	})

	if err != nil {
		return err
	}

	log.Printf("mark default drive id is %s", resp.DefaultDriveID)
	t.DefaultDriveID = resp.DefaultDriveID
	return nil
}

func (t *AliyunDrive) OnStart(ctx context.Context) error {
	if err := t.refreshToken(ctx); err != nil {
		return err
	}

	t.ticker = time.NewTicker(time.Hour)
	t.done = make(chan bool)

	go func() {
		for {
			select {
			case <-t.done:
				return
			case <-t.ticker.C:
				err := t.refreshToken(context.Background())
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("update refresh token is successful")
				}
			}
		}
	}()

	return nil
}

func (t *AliyunDrive) OnStop(ctx context.Context) error {
	t.ticker.Stop()
	t.done <- true
	return nil
}

func (t *AliyunDrive) Info() (interface{}, error) {
	return "This is a test bucket", nil
}

func (t *AliyunDrive) Exists(path string) bool {
	return false
}

func (t *AliyunDrive) Put(task obsync.Task) error {
	t.uploadMutex.Lock()
	defer t.uploadMutex.Unlock()

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
	log.Printf("create remote folder %s is successful, file id is %s", resp.FileName, resp.FileID)

	log.Printf("start upload local file %s to %s", task.Local, task.Key)
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
	log.Printf("upload file is finished, bye~")

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
