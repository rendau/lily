package main

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	st := NewStore(NoExpiration, 0)
	if st == nil {
		t.Error("Fail to create store")
	}

	x, ok := st.Get("a")
	if ok {
		t.Error("got value while key is not exists:", x)
	}

	st.Lock()
	st.Set("a", 1, NoExpiration)
	x, ok = st.Get("a")
	if !ok {
		t.Error("not found key that is exists")
	}
	st.Delete("a")
	x, ok = st.Get("a")
	if ok {
		t.Error("got value while key is not exists:", x)
	}
	st.Unlock()

	st.SetLock("b", 2, NoExpiration)
	x, ok = st.GetLock("b")
	if !ok {
		t.Error("not found key that is exists")
	}
	st.DeleteLock("b")
	x, ok = st.GetLock("b")
	if ok {
		t.Error("got value while key is not exists:", x)
	}
}

func TestStoreExpiration(t *testing.T) {
	st := NewStore(DefaultDuration, 10*time.Millisecond)
	if st == nil {
		t.Error("Fail to create store")
	}
	if st.defaultDuration != NoExpiration {
		t.Error("default duration-value must be NoExpiration")
	}
	st = NewStore(20*time.Millisecond, 10*time.Millisecond)
	if st == nil {
		t.Error("Fail to create store")
	}

	st.SetLock("a", true, 20*time.Millisecond)
	<-time.After(15 * time.Millisecond)
	x, ok := st.GetLock("a")
	if !ok {
		t.Error("not found key that is exists")
	}
	a := x.(bool)
	if a != true {
		t.Error("bad value")
	}
	<-time.After(10 * time.Millisecond)
	x, ok = st.GetLock("a")
	if ok {
		t.Error("got value while key should not exists")
	}

	st.SetLock("a", true, DefaultDuration)
	<-time.After(15 * time.Millisecond)
	x, ok = st.GetLock("a")
	if !ok {
		t.Error("not found key that is exists")
	}
	a = x.(bool)
	if a != true {
		t.Error("bad value")
	}
	st.RefreshExpirationLock("a")
	<-time.After(15 * time.Millisecond)
	x, ok = st.GetLock("a")
	if !ok {
		t.Error("not found key that is exists")
	}
	a = x.(bool)
	if a != true {
		t.Error("bad value")
	}
	<-time.After(15 * time.Millisecond)
	x, ok = st.GetLock("a")
	if ok {
		t.Error("got value while key should not exists")
	}

	st.SetLock("a", true, NoExpiration)
	<-time.After(25 * time.Millisecond)
	x, ok = st.GetLock("a")
	if !ok {
		t.Error("not found key that should exists")
	}
	a = x.(bool)
	if a != true {
		t.Error("bad value")
	}
}
