package test

import (
	"testing"
)

func TestSliceCompaction1(t *testing.T) {
	sli := make([]int,10)
	for i:=0; i<10; i++ {
		sli[i] = i
	}
	s2 := append(sli[:3],sli[5:]...)

	t.Logf("Sli : %v",sli)
	t.Logf("S2  : %v",s2)
}

func TestSliceCompaction2(t *testing.T) {
	sli := make([]int,10)
	for i:=0; i<10; i++ {
		sli[i] = i
	}
	i:=9 // last element
	sli = append(sli[:i],sli[i+1:]...)
	t.Logf("Sli : (len %v) %v ",len(sli),sli)
}

type data struct {
	str string
}
func TestDefer(t *testing.T) {
	val := data{"first val"}
	vp := &val
	defer printValue(t, "value", val)
	defer printValue(t, "pointer", vp)
	defer printValue(t, "field", vp.str)
	defer printValue(t, "fieldP", &(vp.str))

	if true {
		val.str = "second"
	}
}

func printValue(t *testing.T, name string, str interface{}) {
	v := str
	v2, ok := str.(*string)
	if ok {
		v = *v2
	}
	t.Logf("%v is %v ",name,v)
}

