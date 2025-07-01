package summarize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/loispostula/terragrunt-plan-summary/pkg/core/log"
)

func TestFormatCounterAndColor(t *testing.T) {
	assert.Equal(t, "-", formatCounterAndColor(0))
	assert.Equal(t, "3", formatCounterAndColor(3))
}

func TestActionExists(t *testing.T) {
	actions := []resourceActions{
		{action: "[create]"},
		{action: "[update]"},
	}
	assert.True(t, actionExists("[create]", actions))
	assert.False(t, actionExists("[delete]", actions))
}

func TestGetResourceChanges(t *testing.T) {
	rawPlan := map[string]interface{}{
		"resource_changes": []interface{}{
			map[string]interface{}{
				"address": "aws_instance.test",
				"change": map[string]interface{}{
					"actions": []interface{}{"create"},
				},
			},
		},
	}
	res, actions, components := getResourceChanges(rawPlan, "test/component")
	assert.Equal(t, []string{"aws_instance.test"}, res)
	assert.Equal(t, []string{"[create]"}, actions)
	assert.Equal(t, []string{"test/component"}, components)
}

func TestFindPlanFiles(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "test.tfplan.json"), []byte("{}"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte(""), 0600)

	plans := findPlanFiles(dir)
	assert.Len(t, plans, 1)
	assert.Contains(t, plans[0], "test.tfplan.json")
}

func TestSummarizeDetailedPlan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "live__team__env__app.tfplan.json")
	content := []byte(`{
		"resource_changes": [
			{
				"address": "aws_s3_bucket.example",
				"change": { "actions": ["create"] }
			}
		]
	}`)
	err := os.WriteFile(path, content, 0600)
	assert.NoError(t, err)

	err = Summarize(dir, "live__team__env__app", "")
	assert.NoError(t, err)
}

func captureLogOutput(f func()) string {
	logrus.SetFormatter(log.NewTextFormat())
	var buf strings.Builder
	logrus.SetOutput(&buf)
	f()
	logrus.SetOutput(os.Stderr) // restore
	return buf.String()
}

func TestSummarizeAllPlansGolden(t *testing.T) {
	plansDir := filepath.Join("testdata", "input")
	goldenFile := filepath.Join("testdata", "golden", "summarize_all_output.txt")

	output := captureLogOutput(func() {
		err := Summarize(plansDir, "", "^/?([^/]+)/?(.*)")
		require.NoError(t, err)
	})

	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	assert.Equal(t, string(expected), output)
}

func TestSummarizeDetailedPlanGolden(t *testing.T) {
	plansDir := filepath.Join("testdata", "input")
	goldenFile := filepath.Join("testdata", "golden", "summarize_detailed_output.txt")

	output := captureLogOutput(func() {
		err := Summarize(plansDir, "live__team2__env2__app2", "^/?([^/]+)/?(.*)")
		require.NoError(t, err)
	})

	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	assert.Equal(t, string(expected), output)
}
