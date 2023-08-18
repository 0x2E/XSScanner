package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "20060102 15:04:05",
		FullTimestamp:   true,
	})
}

func main() {
	var (
		target  string
		output  string
		verbose bool
	)
	flag.StringVar(&target, "u", "", "target url")
	flag.StringVar(&output, "o", "", "ouput file in json format")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if target == "" {
		logrus.Error("need target url")
		flag.Usage()
		os.Exit(1)
	}
	url, err := url.Parse(target)
	if err != nil {
		logrus.Fatalf("bad url %s\n", err)
	}

	vulns, err := scan(context.Background(), url)
	if err != nil {
		logrus.WithField("sys", "scan").Error(err)
	}

	logrus.Infof("found %d vulns\n", len(vulns))

	res := Res{
		URL:  url.String(),
		Vuln: vulns,
	}
	if err := save(output, res); err != nil {
		logrus.WithField("sys", "output").Error(err)
	}
}

func save(path string, res Res) error {
	jsonData, err := json.MarshalIndent(res, "", "	")
	if err != nil {
		return err
	}
	if path == "" {
		logrus.Info(string(jsonData))
		return nil
	}

	return os.WriteFile(path, jsonData, 0755)
}
