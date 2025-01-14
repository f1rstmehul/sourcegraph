// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
)

func main() {
	config := ci.NewConfig(time.Now())

	pipeline, err := ci.GeneratePipeline(config)
	if err != nil {
		panic(err)
	}

	_, err = pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
