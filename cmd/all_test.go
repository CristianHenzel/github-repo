package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	testConcurrency = "10"
	testUser        = "golibz"
	testDir         = "testdir"
	testToken       = "invalidtoken"
	testURL         = "https://api.github.com"
	testRepo1       = "/apitest1"
	testRepoDir1    = testDir + testRepo2
	testRepoFile1   = testRepoDir1 + testRepo1 + ".txt"
	testRepo2       = "/apitest2"
	testRepoDir2    = testDir + testRepo2
	testRepo3       = "/apitest3"
	testRepoDir3    = testDir + testRepo3
	testRemote      = "/tmp" + testRepo3
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func resetCobraFlags() {
	cFlags.Username = ""
	cFlags.Token = ""
	cFlags.BaseURL = ""
	cFlags.BaseDir = "."
	cFlags.Concurrency = 0
	rootCmd.Flags().Lookup("concurrency").Changed = false
}

func runCobraCmdf(format string, a ...interface{}) {
	argStr := fmt.Sprintf(format, a...)
	fmt.Println("--- Running command with args:", argStr)
	args := strings.Fields(argStr)
	resetCobraFlags()
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	fatalIfError(err)
}

func cleanup() {
	fmt.Println("--- Running cleanup")
	err := os.Remove(configFile)
	fatalIfError(err)

	err = os.RemoveAll(testDir)
	fatalIfError(err)

	err = os.RemoveAll(testRemote)
	fatalIfError(err)
}

// Initialize test variables
func TestInit(t *testing.T) {
	Version = "1.0.0"
	doExit = func(i int) {}
	Execute()
}

// Test Basic commands
func TestBasicCommands(t *testing.T) {
	runCobraCmdf("help")

	runCobraCmdf("version")

	runCobraCmdf("version -u")
}

// Test standard workflow
func TestFlow(t *testing.T) {
	var token = getEnv("GR_TEST_TOKEN", testToken)

	runCobraCmdf("init -u %s -t %s -d %s -c %s", testUser, token, testDir, testConcurrency)

	runCobraCmdf("update")

	runCobraCmdf("pull -c 10")

	runCobraCmdf("status")

	runCobraCmdf("push")

	cleanup()
}

// Test concurrency flag behavior
func TestConcurrency(t *testing.T) {
	var token = getEnv("GR_TEST_TOKEN", testToken)

	runCobraCmdf("init -u %s -t %s -d %s -c 0", testUser, token, testDir)

	runCobraCmdf("pull")

	// Update concurrency in config
	conf := loadConfig()
	conf.Concurrency = 10
	conf.save()
	runCobraCmdf("pull")

	cleanup()
}

// Test init flags
func TestInitFlags(t *testing.T) {
	var token = getEnv("GR_TEST_TOKEN", testToken)

	// Test all init flags
	runCobraCmdf("init -u %s -d %s -c %s -t %s -r %s", testUser, testDir, testConcurrency, token, testURL)

	// Test init when config already exists
	runCobraCmdf("init -u %s", testUser)
	cleanup()

	// Test init with invalid user
	runCobraCmdf("init -u admin -t %s", token)
	cleanup()

	// Test init with no token
	runCobraCmdf("init -u %s", testUser)
	cleanup()

	// Test init with invalid token
	runCobraCmdf("init -u %s -t %s", testUser, testToken)
	cleanup()

	// Test init for user with name and email
	runCobraCmdf("init -u %s -t %s", testUser, token)
	cleanup()

	// Test init with invalid URL
	runCobraCmdf("init -u %s -r http://invalid", testUser)
	cleanup()
}

func TestOther(t *testing.T) {
	var token = getEnv("GR_TEST_TOKEN", testToken)

	runCobraCmdf("init -u %s -d %s -c %s -t %s -r %s", testUser, testDir, testConcurrency, token, testURL)
	runCobraCmdf("pull")
	os.RemoveAll(testRepoDir2)
	os.RemoveAll(testRepoDir3 + "/.git")

	runCobraCmdf("status")

	runCobraCmdf("push")

	runCobraCmdf("pull")
	cleanup()
}

func TestCommandsNoConfig(t *testing.T) {
	// Test pull when no config is present
	runCobraCmdf("pull")

	// Test push when no config is present
	runCobraCmdf("push")

	// Test status when no config is present
	runCobraCmdf("status")
}
