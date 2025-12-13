package openmeteo

import (
	"github.com/google/flatbuffers/go"
)

type VariableWithValues struct {
	_tab flatbuffers.Table
}

func GetRootAsVariableWithValues(buf []byte, offset flatbuffers.UOffsetT) *VariableWithValues {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &VariableWithValues{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *VariableWithValues) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *VariableWithValues) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *VariableWithValues) Variable() Variable {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return Variable(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return VariableUndefined
}

func (rcv *VariableWithValues) Unit() Unit {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return Unit(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return UnitUndefined
}

func (rcv *VariableWithValues) Value() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) Values(j int) float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetFloat32(a + flatbuffers.UOffsetT(j*4))
	}
	return 0
}

func (rcv *VariableWithValues) ValuesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *VariableWithValues) ValuesInt64(j int) int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetInt64(a + flatbuffers.UOffsetT(j*8))
	}
	return 0
}

func (rcv *VariableWithValues) ValuesInt64Length() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *VariableWithValues) Altitude() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) Aggregation() Aggregation {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return Aggregation(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return AggregationNone
}

func (rcv *VariableWithValues) PressureLevel() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) Depth() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) DepthTo() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) EnsembleMember() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) PreviousDay() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(26))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *VariableWithValues) ValuesBytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o)
	}
	return nil
}

func (rcv *VariableWithValues) ValuesInt64Bytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.ByteVector(o)
	}
	return nil
}
