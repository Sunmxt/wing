package kubernetes

func updateStringToStringMap(dst *map[string]string, src map[string]string) (updated bool) {
	updated = false
	if len(src) < 1 {
		return false
	}
	if *dst == nil {
		*dst = map[string]string{}
	}
	for k, v := range src {
		origin, ok := (*dst)[k]
		if origin != v || !ok {
			updated = true
			(*dst)[k] = v
		}
	}
	return
}
