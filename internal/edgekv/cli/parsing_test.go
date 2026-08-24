package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDatabasePolicy(t *testing.T) {
	policy, err := parseDatabasePolicy("restrictDataAccess=true,allowNamespacePolicyOverride=false")
	require.NoError(t, err)
	assert.True(t, policy.RestrictDataAccess)
	assert.False(t, policy.AllowNamespacePolicyOverride)
}

func TestParseDatabasePolicyRequiresBothValues(t *testing.T) {
	_, err := parseDatabasePolicy("restrictDataAccess=true")
	require.Error(t, err)
}

func TestParseTokenPermissions(t *testing.T) {
	permissions, err := parseTokenPermissions("default+rwd,marketing+rw")
	require.NoError(t, err)
	assert.Equal(t, []sdkew.Permission{sdkew.PermissionRead, sdkew.PermissionWrite, sdkew.PermissionDelete}, permissions["default"])
	assert.Equal(t, []sdkew.Permission{sdkew.PermissionRead, sdkew.PermissionWrite}, permissions["marketing"])
}

func TestParseTokenPermissionsRejectsInvalidPermission(t *testing.T) {
	_, err := parseTokenPermissions("default+rx")
	require.Error(t, err)
}

func TestItemValue(t *testing.T) {
	value, contentType, err := itemValue("text", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", value)
	assert.Equal(t, "text/plain", contentType)
}

func TestSaveTokenFileRequiresOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "edgekv_tokens.js")
	token := savedToken{Name: "example", Value: "token"}
	require.NoError(t, saveTokenFile(path, []string{"default"}, token, false))
	require.Error(t, saveTokenFile(path, []string{"default"}, savedToken{Name: "other", Value: "token"}, false))
	require.NoError(t, saveTokenFile(path, []string{"default"}, savedToken{Name: "other", Value: "token"}, true))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content, err := parseTokenFile(data)
	require.NoError(t, err)
	assert.Equal(t, "other", content["default"].Name)
}

func TestSaveTokenBundle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle.tgz")
	writeBundle(t, path, "index.js", "export default {};\n")
	require.NoError(t, saveTokenBundle(path, []string{"default"}, savedToken{Name: "example", Reference: "uuid"}, false))
	assert.Contains(t, readBundleFile(t, path, "edgekv_tokens.js"), "\"default\"")
}

func writeBundle(t *testing.T, path, name, content string) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	require.NoError(t, tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(content))}))
	_, err = tarWriter.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())
	require.NoError(t, file.Close())
}

func readBundleFile(t *testing.T, path, name string) string {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			require.NoError(t, err)
		}
		if header.Name == name {
			var data bytes.Buffer
			_, err = data.ReadFrom(tarReader)
			require.NoError(t, err)
			return data.String()
		}
	}
}
