package main

import (
	"os"

	"github.com/rancher/wrangler/pkg/cleanup"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cleanup.Cleanup("./pkg/simulator/apis"); err != nil {
		logrus.Fatal(err)
	}
	if err := os.RemoveAll("./pkg/simulator/generated"); err != nil {
		logrus.Fatal(err)
	}
}
