package schema_compliance

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		extractSurroundedFencedJSON,
	}
}

func iterativeFixes() []fixFunc {
	return nil
}
