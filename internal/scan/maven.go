package scan

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

type mvnString string

func gatherMvn(mvn string) (string, string, error) {
	mvnDep := strings.ReplaceAll(string(mvn), "\"", "")
	mvnDep = strings.TrimSpace(mvnDep)
	mvnDep = strings.TrimRight(mvnDep, ";")

	mvnDep = mvnDep[:strings.LastIndex(mvnDep, ":")]

	versionidx := strings.LastIndex(mvnDep, ":")

	return strings.TrimRight(mvnDep[:versionidx], ":jar"), "v" + mvnDep[versionidx+1:], nil
}

func GetMvnDeps(path string) (map[string]string, error) {
	var gathered map[string]string

	dirPath := filepath.Dir(path)

	cmd := exec.Command("mvn", "--no-transfer-progress", "dependency:tree", "-DoutputType=dot")
	cmd.Dir = dirPath

	data, err := cmd.Output()

	if err != nil {
		fmt.Println(err)
	}

	res := strings.Split(string(data), "\n")

	gathered = make(map[string]string)

	for _, s := range res {
		// example:
		// [INFO] 	"com.google.inject:guice:jar:4.0:compile (optional) " -> "javax.inject:javax.inject:jar:1:compile (optional) " ;

		// do the lookup once
		sepIdx := strings.Index(s, "->")

		if sepIdx != -1 {
			// skip import and test
			if strings.Contains(s, ":test") || strings.Contains(s, ":import") {
				continue
			}

			// only get the second part
			part := s[sepIdx+len("-> "):]

			repo, version, err := gatherMvn(part)

			// only if no error append
			if err == nil {
				// just in case do the semver thing
				if _, ok := gathered[repo]; ok {
					gathered[repo] = semver.Max(gathered[repo], version)
				} else {
					gathered[repo] = version
				}
			}

		}
	}
	return gathered, nil
}
