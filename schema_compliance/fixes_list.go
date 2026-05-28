package schema_compliance

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		extractSurroundedFencedJSON,
	}
}

func jsonSyntaxFixes() []fixFunc {
	return []fixFunc{
		repairRelaxedJSON,
	}
}

func schemaComplianceFixes() []schemaFixFunc {
	return []schemaFixFunc{
		unwrapResponseObject,
		unwrapSingleItemArray,
	}
}
