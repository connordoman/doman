package openmeteo

import (
	"github.com/google/flatbuffers/go"
)

type VariablesWithMonth struct {
	_tab flatbuffers.Table
}

func (rcv *VariablesWithMonth) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func GetRootAsVariablesWithMonth(buf []byte, offset flatbuffers.UOffsetT) *VariablesWithMonth {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &VariablesWithMonth{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *VariablesWithMonth) Year() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariablesWithMonth) Month() int8 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return int8(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *VariablesWithMonth) Count() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariablesWithMonth) Variables(obj *VariableWithValues, j int) *VariableWithValues {
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

func (rcv *VariablesWithMonth) VariablesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}
