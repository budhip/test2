package transformation

type Metadata struct {
	data map[string]interface{}
}

func NewMetadata(data map[string]interface{}) Metadata {
	if data == nil {
		data = make(map[string]interface{})
	}
	return Metadata{data: data}
}

func (m Metadata) Get(key string) (interface{}, bool) {
	val, exists := m.data[key]
	return val, exists
}

func (m Metadata) GetString(key string) (string, bool) {
	val, exists := m.data[key]
	if !exists {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}

func (m Metadata) Set(key string, value interface{}) {
	m.data[key] = value
}

func (m Metadata) Has(key string) bool {
	_, exists := m.data[key]
	return exists
}

func (m Metadata) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

func (m Metadata) IsEmpty() bool {
	return len(m.data) == 0
}
