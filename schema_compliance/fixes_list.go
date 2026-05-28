package schema_compliance

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		extractSurroundedFencedJSON,
	}
}

func jsonSyntaxFixes() []fixFunc {
	return []fixFunc{
		repairRelaxedJSON,
		repairTruncatedJSON,
	}
}

func schemaComplianceFixes() []schemaFixFunc {
	return []schemaFixFunc{
		unwrapResponseObject,
		unwrapSingleItemArray,
		repairObjectFieldNesting,
	}
}
