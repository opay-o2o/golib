package map2

func Append(m1, m2 map[string]interface{}) {
	for k, v := range m2 {
		m1[k] = v
	}
}
