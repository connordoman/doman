package openmeteo

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type WeatherApiResponse struct {
	_tab flatbuffers.Table
}

func GetRootAsWeatherApiResponse(buf []byte, offset flatbuffers.UOffsetT) *WeatherApiResponse {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := WeatherApiResponse{}
	x.Init(buf, n+offset)
	return &x
}

func (rcv *WeatherApiResponse) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *WeatherApiResponse) Latitude() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WeatherApiResponse) Longitude() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WeatherApiResponse) Elevation() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WeatherApiResponse) GenerationTimeMilliseconds() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WeatherApiResponse) LocationId() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WeatherApiResponse) Model() Model {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return Model(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return ModelUndefined
}

func (rcv *WeatherApiResponse) UtcOffsetSeconds() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WeatherApiResponse) Timezone() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *WeatherApiResponse) TimezoneAbbreviation() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *WeatherApiResponse) Current(obj *VariablesWithTime) *VariablesWithTime {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(VariablesWithTime)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *WeatherApiResponse) Daily(obj *VariablesWithTime) *VariablesWithTime {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(VariablesWithTime)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *WeatherApiResponse) Hourly(obj *VariablesWithTime) *VariablesWithTime {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(26))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(VariablesWithTime)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *WeatherApiResponse) Minutely15(obj *VariablesWithTime) *VariablesWithTime {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(28))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(VariablesWithTime)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *WeatherApiResponse) Monthly(obj *VariablesWithMonth) *VariablesWithMonth {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(30))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(VariablesWithMonth)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}
