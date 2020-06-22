package tools

func (g getFunctions) String(source []string) *[]string {
	newSlice := make([]string, len(source))
	copy(source, newSlice)
	return &newSlice
}

func (g getFunctions) Int(source int) *int {
	if source == 0 {
		return nil
	}
	return &source
}

func (g getFunctions) Float(source float64) *float64 {
	if source == 0 {
		return nil
	}
	return &source
}
