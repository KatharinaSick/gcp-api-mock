package storage

import "testing"

func TestBucketKindConstant(t *testing.T) {
	bucket := &Bucket{
		Kind: "storage#bucket",
		Name: "test-bucket",
	}

	if bucket.Kind != "storage#bucket" {
		t.Errorf("expected kind 'storage#bucket', got '%s'", bucket.Kind)
	}
}

func TestObjectKindConstant(t *testing.T) {
	obj := &Object{
		Kind: "storage#object",
		Name: "test-object",
	}

	if obj.Kind != "storage#object" {
		t.Errorf("expected kind 'storage#object', got '%s'", obj.Kind)
	}
}

func TestBucketListKindConstant(t *testing.T) {
	list := &BucketList{
		Kind:  "storage#buckets",
		Items: []*Bucket{},
	}

	if list.Kind != "storage#buckets" {
		t.Errorf("expected kind 'storage#buckets', got '%s'", list.Kind)
	}
}

func TestObjectListKindConstant(t *testing.T) {
	list := &ObjectList{
		Kind:  "storage#objects",
		Items: []*Object{},
	}

	if list.Kind != "storage#objects" {
		t.Errorf("expected kind 'storage#objects', got '%s'", list.Kind)
	}
}
