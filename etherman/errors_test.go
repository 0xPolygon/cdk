package etherman

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryParseWithExactMatch(t *testing.T) {
	expected := ErrTimestampMustBeInsideRange
	smartContractErr := expected

	actualErr, ok := TryParseError(smartContractErr)

	assert.ErrorIs(t, actualErr, expected)
	assert.True(t, ok)
}

func TestTryParseWithContains(t *testing.T) {
	expected := ErrTimestampMustBeInsideRange
	smartContractErr := fmt.Errorf(" execution reverted: ProofOfEfficiency::sequenceBatches: %w", expected)

	actualErr, ok := TryParseError(smartContractErr)

	assert.ErrorIs(t, actualErr, expected)
	assert.True(t, ok)
}

func TestTryParseWithNonExistingErr(t *testing.T) {
	smartContractErr := fmt.Errorf("some non-existing err")

	actualErr, ok := TryParseError(smartContractErr)

	assert.Nil(t, actualErr)
	assert.False(t, ok)
}
