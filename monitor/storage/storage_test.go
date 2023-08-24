package storage

import (
	"testing"
)

var _ DataUpdate = &MyData{}

type MyData struct {
	Data        interface{}
	BlockNumber uint64
}

func (u *MyData) NeedUpdate(new interface{}) bool {
	return new.(*MyData).BlockNumber > u.BlockNumber
}

func TestNeedUpdate(t *testing.T) {
	m := &DatasStorage{}
	m.Store(
		[]interface{}{
			"1",
			"2",
		},
		[]interface{}{
			&MyData{1, 100},
			&MyData{2, 100},
		},
	)
	if m.Load("1").(*MyData).Data.(int) != 1 {
		t.Fatal(m.Load("1"))
	}
	if m.Load("2").(*MyData).Data.(int) != 2 {
		t.Fatal(m.Load("2"))
	}
	m.Store(
		[]interface{}{
			"1",
			"2",
			"3",
		},
		[]interface{}{
			&MyData{11, 88},
			&MyData{12, 188},
			&MyData{13, 100},
		},
	)
	if m.Load("1").(*MyData).Data.(int) != 1 {
		t.Fatal(m.Load("1"))
	}
	if m.Load("2").(*MyData).Data.(int) != 12 {
		t.Fatal(m.Load("2"))
	}
	if m.Load("3").(*MyData).Data.(int) != 13 {
		t.Fatal(m.Load("3"))
	}
}
