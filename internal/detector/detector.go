package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

type Language string

const (
	LangDockerfile Language = "dockerfile"
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
	HasDockerfile bool
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
		result.HasDockerfile = true
		result.DockerFilePath = dockerfilepath
		result.Language = LangDockerfile
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
		if err == nil {
			result.Framework = detectFramework(pkg)
			result.Version = detectNodeVersion(pkg)
		} else {
			result.Framework = "generic-node"
			result.Version = "20"
		}
	}
	return result
}

func detectFramework(pkg *PackageJSON) string {
	if _, ok := pkg.Dependencies["next"]; ok {
		return "nextjs"
	}
	if _, ok := pkg.DevDependencies["vite"]; ok {
		return "vite"
	}
	if _, ok := pkg.Dependencies["react-scripts"]; ok {
		return "cra"
	}
	if _, ok := pkg.Dependencies["express"]; ok {
		return "express"
	}
	if _, ok := pkg.Dependencies["fastify"]; ok {
		return "fastify"
	}
	return "generic-node"
}

func detectNodeVersion(pkg *PackageJSON) string {
	raw, ok := pkg.Engines["node"]
	if !ok || raw == "" {
		return "20"
	}

	// strip symbols like >=, ^, ~, spaces — keep only the major version number
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(raw)
	if match == "" {
		return "20"
	}

	return match
}


func readPackageJson(repoPath string) (*PackageJSON, error) {
	path := filepath.Join(repoPath, "package.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}