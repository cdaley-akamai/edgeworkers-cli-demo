package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountResourceTiersCmdUsesContractArgument(t *testing.T) {
	cmd := newAccountResourceTiersCmd()

	assert.Equal(t, "resource-tiers <contract-id>", cmd.Use)
	assert.Nil(t, cmd.Flags().Lookup("contract-id"))
	assert.NoError(t, cmd.Args(cmd, []string{"ctr_1-1TJ"}))
	assert.Error(t, cmd.Args(cmd, nil))
	assert.Error(t, cmd.Args(cmd, []string{"one", "two"}))
}
