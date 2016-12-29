package lily

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	st := NewStore(StoreNoExpiration, 0)
	if st == nil {
		t.Error("Fail to create store")
	}

	x, ok := st.Get("a", false)
	if ok {
		t.Error("got value while key is not exists:", x)
	}

	st.Lock()
	st.Set("a", 1, StoreNoExpiration)
	x, ok = st.Get("a", false)
	if !ok {
		t.Error("not found key that is exists")
	}
	st.Delete("a")
	x, ok = st.Get("a", false)
	if ok {
		t.Error("got value while key is not exists:", x)
	}
	st.Unlock()

	st.SetLock("a", true, StoreNoExpiration)
	x, ok = st.GetLock("a", false)
	if !ok {
		t.Error("not found key that is exists")
	}
	st.SetLock("b", false, StoreNoExpiration)
	x, ok = st.GetLock("b", false)
	if !ok {
		t.Error("not found key that is exists")
	}
	all := st.GetAllLock()
	if len(all) != 2 {
		t.Error("must be two elements")
	}
	st.DeleteLock("a")
	x, ok = st.GetLock("a", false)
	if ok {
		t.Error("got value while key is not exists:", x)
	}
	st.DeleteLock("b")
	x, ok = st.GetLock("b", false)
	if ok {
		t.Error("got value while key is not exists:", x)
	}
}

func TestStoreExpiration(t *testing.T) {
	st := NewStore(StoreDefaultDuration, 10*time.Millisecond)
	if st == nil {
		t.Error("Fail to create store")
	}
	if st.defaultDuration != StoreNoExpiration {
		t.Error("default duration-value must be StoreNoExpiration")
	}
	st = NewStore(20*time.Millisecond, 10*time.Millisecond)
	if st == nil {
		t.Error("Fail to create store")
	}

	st.SetLock("a", true, 20*time.Millisecond)
	<-time.After(15 * time.Millisecond)
	x, ok := st.GetLock("a", false)
	if !ok {
		t.Error("not found key that is exists")
	}
	a := x.(bool)
	if a != true {
		t.Error("bad value")
	}
	<-time.After(10 * time.Millisecond)
	x, ok = st.GetLock("a", false)
	if ok {
		t.Error("got value while key should not exists")
	}

	st.SetLock("a", true, StoreDefaultDuration)
	<-time.After(15 * time.Millisecond)
	x, ok = st.GetLock("a", false)
	if !ok {
		t.Error("not found key that is exists")
	}
	a = x.(bool)
	if a != true {
		t.Error("bad value")
	}
	st.RefreshExpirationLock("a")
	<-time.After(15 * time.Millisecond)
	x, ok = st.GetLock("a", false)
	if !ok {
		t.Error("not found key that is exists")
	}
	a = x.(bool)
	if a != true {
		t.Error("bad value")
	}
	<-time.After(15 * time.Millisecond)
	x, ok = st.GetLock("a", false)
	if ok {
		t.Error("got value while key should not exists")
	}

	st.SetLock("a", true, StoreNoExpiration)
	<-time.After(25 * time.Millisecond)
	x, ok = st.GetLock("a", false)
	if !ok {
		t.Error("not found key that should exists")
	}
	a = x.(bool)
	if a != true {
		t.Error("bad value")
	}
}
