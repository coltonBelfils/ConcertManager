package appPath

import (
	"fmt"
	"github.com/cockroachdb/errors"
	"os"
	"strings"
	"sync"
)

var (
	base      string
	baseMutex sync.Mutex
)

/*
Path takes a path and returns the full path to the file or directory starting from the environment variable APP_ROOT, which should be a path.
It will create the path if it does not exist.
If APP_ROOT is not set, it will default to the current working directory.
*/
func Path(path string) string {
	baseMutex.Lock()
	defer baseMutex.Unlock()

	if base == "" {
		var ok bool
		base, ok = os.LookupEnv("APP_ROOT")
		if !ok {
			workingDir, getWDErr := os.Getwd()
			if getWDErr != nil {
				panic(errors.Wrap(getWDErr, "error getting working directory"))
			}
			fmt.Printf("working directory: %s\n", workingDir)
			base = fmt.Sprintf("%s/", workingDir)
		}

		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
	}

	path = strings.TrimPrefix(path, "/")

	makePathSplit := strings.Split(path, "/")
	makePathSplit = makePathSplit[:len(makePathSplit)-1]
	makePath := strings.Join(makePathSplit, "/")
	makePath = base + makePath

	makePathErr := os.MkdirAll(makePath, os.ModePerm)
	if makePathErr != nil {
		fmt.Printf("last index: %d\n", strings.LastIndex(path, "/"))
		panic(errors.Wrapf(makePathErr, "error creating path %s", makePath))
	}

	return base + path
}
