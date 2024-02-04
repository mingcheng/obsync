package obsync

//func TestBucketCallback(t *testing.T) {
//	var name = "test"
//	Register(name, func(_ Config) (Bucket, error) {
//		return nil, nil
//	})
//
//	if _, err := Func(name); err != nil {
//		t.Error(err)
//	}
//
//	_ = RemoveBucketSyncFunc(name)
//
//	if callback, err := Func(name); callback != nil {
//		t.Errorf("not cleaned, %v", err)
//	}
//}
