package bucket

import (
	"testing"
)

func TestTestBucket_Info(t *testing.T) {
	bucket := TestBucket{}
	if _, err := bucket.Info(); err != nil {
		t.Error(err)
	}
}
