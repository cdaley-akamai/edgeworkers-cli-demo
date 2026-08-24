package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseSetPolicyCmdUsesPolicyArgument(t *testing.T) {
	cmd := newDatabaseSetPolicyCmd()

	assert.Equal(t, "set-policy <data-access-policy>", cmd.Use)
	assert.Nil(t, cmd.Flags().Lookup("data-access-policy"))
	assert.NoError(t, cmd.Args(cmd, []string{"restrictDataAccess=true,allowNamespacePolicyOverride=false"}))
	assert.Error(t, cmd.Args(cmd, nil))
	assert.Error(t, cmd.Args(cmd, []string{"one", "two"}))
}
