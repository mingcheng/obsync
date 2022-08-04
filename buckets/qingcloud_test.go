package buckets

import (
	"context"
	"github.com/mingcheng/obsync"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewQingClient(t *testing.T) {
	if os.Getenv("QING_KEY") == "" {
		log.Warn("QING_KEY is not set in environment variable")
		return
	}

	client, err := NewQingClient(obsync.BucketConfig{
		Key:      os.Getenv("QING_KEY"),
		Secret:   os.Getenv("QING_SECRET"),
		EndPoint: os.Getenv("QING_ENDPOINT"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)

	err = client.Put(context.TODO(), "/etc/hosts", "hosts.txt")
	assert.NoError(t, err)

	err = client.Del(context.TODO(), "hosts.txt")
	assert.NoError(t, err)
}
