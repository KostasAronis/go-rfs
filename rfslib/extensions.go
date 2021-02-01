package rfslib

func (r *Record) ToString() string {
	return string(r[:])
}

func (r *Record) FromString(str string) {
	copy(r[:], str[:])
}

func (r *Record) FromFloatArrayInterface(i interface{}) *Record {
	switch i.(type) {
	case []uint8:
		arr := i.([]uint8)
		for i, v := range arr {
			r[i] = byte(v)
		}
		return r
	case []interface{}:
		arr := i.([]interface{})
		for i, v := range arr {
			r[i] = byte(v.(float64))
		}
		return r
	}
	panic("type not found")
}
