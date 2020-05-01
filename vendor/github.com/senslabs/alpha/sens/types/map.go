package types

type Map map[string]interface{}

func (this *Map) GetOrDefault(key string, val interface{}) interface{} {
	if v, ok := (*this)[key]; ok {
		return v
	}
	return val
}

func (this *Map) Get(key ...string) interface{} {
	var val interface{} = *this
	for _, k := range key {
		val = val.(Map)[k]
	}
	return val
}

func (this *Map) Put(key string, val interface{}) {
	(*this)[key] = val
}

func Create() *Map {
	return &Map{}
}
