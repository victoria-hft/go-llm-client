package schema_compliance

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		stripTransportJunk,
		extractSurroundedFencedJSON,
	}
}

func jsonSyntaxFixes() []fixFunc {
	return []fixFunc{
		repairRelaxedJSON,
		removeZeroWidthCharactersFromKeys,
		repairTruncatedJSON,
	}
}

func schemaComplianceFixes() []schemaFixFunc {
	return []schemaFixFunc{
		unwrapResponseObject,
		unwrapSingleItemArray,
		repairObjectFieldNesting,
		repairScalarSchemaValues,
		repairEnumStringValues,
	}
}
