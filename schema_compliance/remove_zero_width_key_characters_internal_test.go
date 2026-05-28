package schema_compliance

import "testing"

func TestRemoveZeroWidthCharactersFromKeysReturnsUnchangedOnInvalidJSON(t *testing.T) {
	if got, changed := removeZeroWidthCharactersFromKeys("{\"na\u200bme\":"); changed || got != "{\"na\u200bme\":" {
		t.Fatalf("removeZeroWidthCharactersFromKeys() = %q, %v; want unchanged false", got, changed)
	}
}

func TestRemoveZeroWidthCharactersFromKeysReturnsUnchangedOnCollision(t *testing.T) {
	input := "{\"name\":\"Ada\",\"na\u200bme\":\"Grace\"}"
	if got, changed := removeZeroWidthCharactersFromKeys(input); changed || got != input {
		t.Fatalf("removeZeroWidthCharactersFromKeys() = %q, %v; want unchanged false", got, changed)
	}
}
