package detector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Language constants 
const (
	LangNode   = "node"
	LangPython = "python"
	LangGo     = "go"
)

// DetectResult holds everything the builder and docker generator need
type DetectResult struct {
	Language       string
	Framework      string
	Version        string 
	PackageManager string
	BuildCommand   []string
	RunCommand     []string
	Port           int
	DependencyFile string
	EnvFile        string
	MultiStage     bool
	BaseImage      string
	HasDockerfile  bool
}

func Detect(path string) DetectResult {
	result := DetectResult{}

	// Check for an existing Dockerfile (case-insensitive search for both
	// "Dockerfile" and "dockerfile").
	for _, name := range []string{"Dockerfile", "dockerfile"} {
		if fileExists(filepath.Join(path, name)) {
			//detect port and add it to result
			result.HasDockerfile = true
			
		}
	}

	indicators := []struct {
		file     string
		language string
	}{
		{"package.json", LangNode},
		{"requirements.txt", LangPython},
		{"pyproject.toml", LangPython},
		{"go.mod", LangGo},
	}

	for _, ind := range indicators {
		if fileExists(filepath.Join(path, ind.file)) {
			result.Language = ind.language
			result.DependencyFile = ind.file
			break
		}
	}

	// Detect .env file presence.
	if fileExists(filepath.Join(path, ".env")) {
		result.EnvFile = ".env"
	}

	DetectFramework(path, &result)

	return result
}

func DetectFramework(path string, result *DetectResult) {
	switch result.Language {
	case LangNode:
		detectNode(path, result)
	case LangPython:
		detectPython(path, result)
	case LangGo:
		detectGo(path, result)
	}
}

// ---------------------------------------------------------------------------
// Node / JavaScript
// ---------------------------------------------------------------------------

// packageJSON mirrors the fields we care about from package.json.
type packageJSON struct {
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	Engines      struct {
		Node string `json:"node"`
	} `json:"engines"`
	Main string `json:"main"`
}

func detectNode(path string, result *DetectResult) {
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	// --- Package manager ---
	switch {
	case fileExists(filepath.Join(path, "pnpm-lock.yaml")):
		result.PackageManager = "pnpm"
	case fileExists(filepath.Join(path, "yarn.lock")):
		result.PackageManager = "yarn"
	default:
		result.PackageManager = "npm"
	}

	// --- Version from engines.node (e.g. ">=18 <21" → "18") ---
	if v := cleanSemver(pkg.Engines.Node); v != "" {
		result.Version = v
	} else {
		result.Version = "20" // LTS default
	}

	// --- Framework detection from dependencies ---
	deps := pkg.Dependencies
	switch {
	case hasDep(deps, "next"):
		result.Framework = "nextjs"
		result.Port = 3000
		result.BuildCommand = []string{result.PackageManager, "run", "build"}
		result.RunCommand = []string{result.PackageManager, "start"}
		result.MultiStage = true

	case hasDep(deps, "react-scripts") || hasDep(deps, "vite"):
		result.Framework = "react"
		result.Port = 3000
		result.BuildCommand = []string{result.PackageManager, "run", "build"}
		result.RunCommand = []string{"serve", "-s", "dist", "-l", "3000"}

	case hasDep(deps, "express"):
		result.Framework = "express"
		result.Port = 3000
		entry := pkg.Main
		if entry == "" {
			entry = "index.js"
		}
		result.RunCommand = []string{"node", entry}

	case hasDep(deps, "fastify"):
		result.Framework = "fastify"
		result.Port = 3000
		result.RunCommand = []string{"node", "index.js"}

	case hasDep(deps, "koa"):
		result.Framework = "koa"
		result.Port = 3000
		result.RunCommand = []string{"node", "index.js"}

	case hasDep(deps, "hapi") || hasDep(deps, "@hapi/hapi"):
		result.Framework = "hapi"
		result.Port = 3000
		result.RunCommand = []string{"node", "index.js"}

	default:
		// Plain Node app — try to use the start script.
		result.Framework = "node"
		result.Port = 3000
		if start, ok := pkg.Scripts["start"]; ok {
			result.RunCommand = strings.Fields(start)
		} else {
			result.RunCommand = []string{"node", "index.js"}
		}
	}

	// Install step is always first in BuildCommand for non-Next projects.
	if len(result.BuildCommand) == 0 {
		result.BuildCommand = []string{result.PackageManager, "install"}
	} else {
		// Prepend install before the framework build step.
		result.BuildCommand = append(
			[]string{result.PackageManager, "install", "&&"},
			result.BuildCommand...,
		)
	}

	result.BaseImage = "node:" + result.Version + "-alpine"
}

// ---------------------------------------------------------------------------
// Python
// ---------------------------------------------------------------------------

func detectPython(path string, result *DetectResult) {
	// --- Version from .python-version or runtime.txt ---
	for _, vf := range []string{".python-version", "runtime.txt"} {
		if v := readFirstLine(filepath.Join(path, vf)); v != "" {
			result.Version = strings.TrimPrefix(strings.TrimSpace(v), "python-")
			break
		}
	}
	if result.Version == "" {
		result.Version = "3.12" // sensible default
	}

	// Collect all dependency content for framework detection.
	var depContent string
	if result.DependencyFile == "requirements.txt" {
		depContent = readFileContent(filepath.Join(path, "requirements.txt"))
	} else if result.DependencyFile == "pyproject.toml" {
		depContent = readFileContent(filepath.Join(path, "pyproject.toml"))
	}

	lowerDeps := strings.ToLower(depContent)

	switch {
	case strings.Contains(lowerDeps, "django"):
		result.Framework = "django"
		result.Port = 8000
		result.RunCommand = []string{"python", "manage.py", "runserver", "0.0.0.0:8000"}
		result.BuildCommand = []string{"pip", "install", "-r", "requirements.txt"}

	case strings.Contains(lowerDeps, "fastapi"):
		result.Framework = "fastapi"
		result.Port = 8000
		result.RunCommand = []string{"uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"}
		result.BuildCommand = []string{"pip", "install", "-r", "requirements.txt"}

	case strings.Contains(lowerDeps, "flask"):
		result.Framework = "flask"
		result.Port = 5000
		result.RunCommand = []string{"python", "main.py"}
		result.BuildCommand = []string{"pip", "install", "-r", "requirements.txt"}

	case strings.Contains(lowerDeps, "tornado"):
		result.Framework = "tornado"
		result.Port = 8888
		result.RunCommand = []string{"python", "main.py"}
		result.BuildCommand = []string{"pip", "install", "-r", "requirements.txt"}

	default:
		result.Framework = "python"
		result.Port = 8000
		result.RunCommand = []string{"python", "main.py"}
		result.BuildCommand = []string{"pip", "install", "-r", "requirements.txt"}
	}

	result.BaseImage = "python:" + result.Version + "-slim"
}

// ---------------------------------------------------------------------------
// Go
// ---------------------------------------------------------------------------

var goVersionRe = regexp.MustCompile(`^go\s+([\d.]+)`)

func detectGo(path string, result *DetectResult) {
	modPath := filepath.Join(path, "go.mod")
	content := readFileContent(modPath)

	// Extract the Go version declared in go.mod (e.g. "go 1.22.3").
	for _, line := range strings.Split(content, "\n") {
		if m := goVersionRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			result.Version = m[1]
			break
		}
	}
	if result.Version == "" {
		result.Version = "1.22"
	}

	// Detect web framework from go.mod require block.
	lowerContent := strings.ToLower(content)
	switch {
	case strings.Contains(lowerContent, "gin-gonic/gin"):
		result.Framework = "gin"
	case strings.Contains(lowerContent, "labstack/echo"):
		result.Framework = "echo"
	case strings.Contains(lowerContent, "go-chi/chi"):
		result.Framework = "chi"
	case strings.Contains(lowerContent, "gofiber/fiber"):
		result.Framework = "fiber"
	case strings.Contains(lowerContent, "gorilla/mux"):
		result.Framework = "gorilla"
	default:
		result.Framework = "go"
	}

	result.Port = 8080
	result.BuildCommand = []string{"go", "build", "-o", "app", "./..."}
	result.RunCommand = []string{"./app"}
	result.MultiStage = true
	result.BaseImage = "golang:" + result.Version + "-alpine"
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// hasDep returns true if the dependency map contains the given key.
func hasDep(deps map[string]string, name string) bool {
	_, ok := deps[name]
	return ok
}

// cleanSemver strips range operators from a semver string and returns the
// major version number (e.g. ">=18 <21" → "18", "^20.1.0" → "20").
func cleanSemver(raw string) string {
	re := regexp.MustCompile(`(\d+)`)
	if m := re.FindString(strings.TrimSpace(raw)); m != "" {
		return m
	}
	return ""
}

// readFirstLine returns the first non-empty line of a file.
func readFirstLine(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if line := strings.TrimSpace(s.Text()); line != "" {
			return line
		}
	}
	return ""
}

// readFileContent returns the full text of a file, or "" on error.
func readFileContent(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}