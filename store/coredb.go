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

	// Check if timestamp file exists, create if not
	timestampFile := "store/dev_data_timestamp"
	b, err := os.ReadFile(timestampFile)
	if err != nil {
		// If file doesn't exist, create it with timestamp 0 to force refresh
		if os.IsNotExist(err) {
			g.Info("Timestamp file not found, creating it and forcing data refresh")
			timestamp := int64(0)
			os.WriteFile(timestampFile, []byte(cast.ToString(timestamp)), 0644)
			b = []byte("0")
		} else {
			g.LogFatal(err, "could not read timestamp file")
		}
	}
	timestamp := cast.ToInt64(string(b))

	filePaths := []string{
		path.Join(os.Getenv("DBT_TARGET_FOLDER"), "core_obras_plus.csv.gz"),
		path.Join(os.Getenv("DBT_TARGET_FOLDER"), "core_obras_plus_phone.csv.gz"),
		path.Join(os.Getenv("DBT_TARGET_FOLDER"), "core_obras_plus_email.csv.gz"),
	}

	refresh := false
	for _, filePath := range filePaths {
		stat, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				g.Warn("Data file not found: %s", filePath)
				continue // Skip missing files
			} else {
				g.LogFatal(err, "could not stat file: %s", filePath)
			}
		}

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
		// Check if source file exists before copying
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			g.Warn("Skipping missing file: %s", filePath)
			continue
		}
		
		err := exec.Command("cp", "-f", filePath, "./data/core/").Run()
		if err != nil {
			g.Warn("Could not copy %s: %v", filePath, err)
			// Don't fatal, just warn and continue
		} else {
			g.Info("Copied: %s", filePath)
		}
	}

	// load the files
	proc, err := process.NewProc("sling")
	g.LogFatal(err, "could not prep sling process")

	proc.Print = true
	proc.Env = g.KVArrToMap(os.Environ()...)
	err = proc.Run("run", "-d", "-r", "store/sling/build.sqlite.core.yaml")
	g.LogFatal(err, "could not run sling process")

	timestamp = time.Now().Unix()
	os.WriteFile(timestampFile, []byte(cast.ToString(timestamp)), 0644)

	return nil
}
