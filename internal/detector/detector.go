package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Language string

const (
	LangDockerFile Language = "dockerfile"
	LangNode       Language = "node"
	LangPython     Language = "python"
	LangGo         Language = "go"
	LangJava       Language = "java"
	LangRust       Language = "rust"
	LangRuby	   Language = "ruby"
	LangUnknown    Language = "unknown"
)

type DetectResult struct {
	Language Language
	HasDockerFile bool
	DockerFilePath string
	Framework string
	Version string
}

type PackageJSON struct {
	Dependencies map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts map[string]string `json:"scripts"`
	Engines map[string]string `json:"engines"`

}

//inspect root of the repo and return detected language

func Detect(repoPath string) DetectResult {
	result := DetectResult{}

	dockerfilepath := filepath.Join(repoPath, "Dockerfile")
	if fileExists(dockerfilepath) {
		result.HasDockerFile = true
		result.DockerFilePath = dockerfilepath
		result.Language = LangDockerFile
		return result
	}

	indicators := []struct {
		file string
		language Language
	} {
		{"package.json", LangNode},
		{"requirements.txt", LangPython},
		{"pyproject.toml", LangPython},
		{"setup.py", LangPython},
		{"go.mod", LangGo},
		{"Gemfile", LangRuby},
		{"pom.xml", LangJava},
		{"build.gradle", LangJava},
		{"Cargo.toml", LangRust},
	}
	for _, ind := range indicators {
		if fileExists(filepath.Join(repoPath, ind.file)) {
			result.Language = ind.language
			break
		}
	}

	if result.Language == "" {
		result.Language = LangUnknown
		return result
	}

	if result.Language == LangNode {
		pkg, err := readPackageJson(repoPath)
	}
}


func readPackageJson(path string) (*PackageJSON, error) {
	
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}