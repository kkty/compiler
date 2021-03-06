package stringmap

type Map map[string]string

func New() Map {
	return Map{}
}

func (m Map) Copy() Map {
	copied := Map{}
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

func (m Map) Keys() []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Join joins two maps and returns a function to restore the original map.
// It can be used like `restore := m.Join(mm); ... ; restore(m)`
func (m Map) Join(mm Map) func(Map) {
	added := []string{}
	updated := Map{}
	for k, v := range mm {
		if originalValue, exists := m[k]; exists {
			updated[k] = originalValue
		} else {
			added = append(added, k)
		}
		m[k] = v
	}
	return func(m Map) {
		m.Remove(added)
		for k, v := range updated {
			m[k] = v
		}
	}
}

// Remove removes elements from the map.
func (m Map) Remove(keys []string) {
	for _, k := range keys {
		delete(m, k)
	}
}
