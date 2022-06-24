package bucket

import (
	"testing"
)

// TestBucket_Info to test bucket information functions
func TestBucket_Info(t *testing.T) {
	bucket := TestBucket{}
	if _, err := bucket.Info(); err != nil {
		t.Error(err)
	}
}
