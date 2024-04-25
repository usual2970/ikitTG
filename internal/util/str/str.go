package str

func IsString(i interface{}) bool {
	_, ok := i.(string)
	return ok
}

func IsNotEmptyString(i interface{}) bool {
	s, ok := i.(string)
	if !ok {
		return false
	}

	return len(s) > 0
}
