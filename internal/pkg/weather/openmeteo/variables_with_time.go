package openmeteo

import (
	"github.com/google/flatbuffers/go"
)

type VariablesWithTime struct {
	_tab flatbuffers.Table
}

func (rcv *VariablesWithTime) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func GetRootAsVariablesWithTime(buf []byte, offset flatbuffers.UOffsetT) *VariablesWithTime {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &VariablesWithTime{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *VariablesWithTime) Time() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariablesWithTime) TimeEnd() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariablesWithTime) Interval() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariablesWithTime) Variables(obj *VariableWithValues, j int) *VariableWithValues {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		a := rcv._tab.Vector(o)
		if obj == nil {
			obj = new(VariableWithValues)
		}
		obj.Init(rcv._tab.Bytes, rcv._tab.Indirect(a+flatbuffers.UOffsetT(j*4)))
		return obj
	}
	return nil
}

func (rcv *VariablesWithTime) VariablesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}
