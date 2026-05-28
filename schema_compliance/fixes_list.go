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

func schemaComplianceFixes() []fixFunc {
	return nil
}
