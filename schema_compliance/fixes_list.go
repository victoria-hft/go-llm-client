package schema_compliance

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		stripTransportJunk,
		extractSurroundedFencedJSON,
	}
}

func jsonSyntaxFixes() []fixFunc {
	return []fixFunc{
		repairSmartQuoteDelimiters,
		repairRelaxedJSON,
		removeZeroWidthCharactersFromKeys,
		repairTruncatedJSON,
	}
}

func schemaComplianceFixes() []schemaFixFunc {
	return []schemaFixFunc{
		unwrapResponseObject,
		repairKeyValueArrayObject,
		unwrapSingleItemArray,
		repairNumericKeyObjectArray,
		repairObjectFieldNesting,
		repairScalarSchemaValues,
		repairEnumStringValues,
	}
}
