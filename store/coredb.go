package store

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	_ "embed"

	"github.com/flarco/g"
	"github.com/flarco/g/process"
	"github.com/spf13/cast"
)

var ObrasCities = map[string]struct{}{}
var Users = map[string]User{}

func init() {
	///////////////////  Init //////////////
	if strings.ToLower(os.Getenv("ENV")) == "development" {
		importDevData()
	}
}

func importDevData() error {

	b, err := os.ReadFile("store/dev_data_timestamp")
	g.LogFatal(err)
	timestamp := cast.ToInt64(string(b))

	filePaths := []string{
		path.Join(os.Getenv("DBT_TARGET_FOLDER"), "core_obras_plus.csv.gz"),
		path.Join(os.Getenv("DBT_TARGET_FOLDER"), "core_obras_plus_phone.csv.gz"),
		path.Join(os.Getenv("DBT_TARGET_FOLDER"), "core_obras_plus_email.csv.gz"),
	}

	refresh := false
	for _, filePath := range filePaths {
		stat, err := os.Stat(filePath)
		g.LogFatal(err)

		if stat.ModTime().Unix() > timestamp {
			g.Info("%d > timestamp (%d)", stat.ModTime().Unix(), timestamp)
			refresh = true
		}
	}

	// refresh = true
	if !refresh {
		return nil
	}

	g.Info("Importing DEV data")

	// copy the files
	for _, filePath := range filePaths {
		err := exec.Command("cp", "-f", filePath, "./data/core/").Run()
		g.LogFatal(err, "could not copy %s", filePath)
	}

	// load the files
	proc, err := process.NewProc("sling")
	g.LogFatal(err, "could not prep sling process")

	proc.Print = true
	proc.Env = g.KVArrToMap(os.Environ()...)
	err = proc.Run("run", "-d", "-r", "store/sling/build.sqlite.core.yaml")
	g.LogFatal(err, "could not run sling process")

	timestamp = time.Now().Unix()
	os.WriteFile("store/dev_data_timestamp", []byte(cast.ToString(timestamp)), 0777)

	return nil
}
