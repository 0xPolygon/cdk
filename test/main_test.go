package test

import (
	"os"
	"strings"
	"testing"

	"github.com/0xPolygon/cdk/test/operations"
)

const thisRelativePath = "cdk/test"

func TestMain(t *testing.M) {
	var err error
	baseDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if !strings.HasSuffix(baseDir, thisRelativePath) {
		panic("run the test from the test directory")
	}
	err = operations.RunCDKStack()
	defer func() {
		if err := operations.StopCDKStack(); err != nil {
			panic(err)
		}
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		return
	}
	t.Run()
}
