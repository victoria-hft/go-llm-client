package schema_compliance

func oneTimeFixes() []fixFunc {
	return []fixFunc{
		stripTransportJunk,
		extractSurroundedFencedJSON,
	}
}

func schemaOneTimeFixes() []schemaOneTimeFixFunc {
	return []schemaOneTimeFixFunc{
		repairNDJSONArrayOutput,
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
		repairItemItemsShape,
		repairObjectFieldNesting,
		repairEmptyContainerNullability,
		repairScalarSchemaValues,
		repairEnumStringValues,
	}
}
