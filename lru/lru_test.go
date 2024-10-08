package lru

import (
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil, 60)

	lru.Add("key1", String("1234"), 60)
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 faild")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	Cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(Cap), nil, 60)
	lru.Add(k1, String(v1), 60)
	lru.Add(k2, String(v2), 60)
	lru.Add(k3, String(v3), 60)

	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicited(t *testing.T) {

}
