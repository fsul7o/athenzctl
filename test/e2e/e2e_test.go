//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

func TestFeatures(t *testing.T) {
	if os.Getenv("ATHENZCTL_E2E_CONFIG") == "" {
		t.Skip("ATHENZCTL_E2E_CONFIG not set; run `make e2e-up` first")
	}
	tags := os.Getenv("GODOG_TAGS")
	if tags == "" {
		tags = "~@skip"
	}
	suite := godog.TestSuite{
		Name:                 "athenzctl-e2e",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     tags,
			Output:   colors.Colored(os.Stdout),
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite failed")
	}
}
