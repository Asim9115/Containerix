package main

import (
	"fmt"
	"os"
	//"os/exec"
	"path/filepath"
"encoding/json"
	
	
	"regexp"
	
)

func main() {
	// var url string
	// fmt.Println("testing github clone")
	// fmt.Println("Enter Url")
	// fmt.Scanln(&url)

	// cmd := exec.Command("git", "clone", url)
	// output , err := cmd.CombinedOutput()
	// if err !=  nil {
	// 	fmt.Println(err)
	// 	fmt.Println(string(output))
	// }
	fmt.Println("cloning Done")
	//  op := buildDockerImage("node","./VirtualStox/nodeSocket" )
	// // fmt.Sprintf(err)
	// fmt.Println(op)
		detected := Detect("./VirtualStox/Backend")
	
	fmt.Println(detected)
}






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
		return "express"
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

// func buildDockerImage(language string, repoPath string) error {
// 	imageId := uuid.New().String()
// 	tag := "containerix-" + imageId

// 	fmt.Println("building Docker image:", tag, "framework", language)

// 	dockerfilePath := filepath.Join(repoPath, "Dockerfile")

// 	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
// 		fmt.Println("No Dockerfile found, generating via nixpacks")

// 		nixCmd := exec.Command("nixpacks", "build", repoPath, "--name", tag)

// 		output, err := nixCmd.CombinedOutput()
// 		if err != nil {
// 			return fmt.Errorf("nixpacks failed: %v\n%s", err, string(output))
// 		}

// 		return nil
// 	}

// 	cmd := exec.Command("docker", "build", "-t", tag, repoPath)

// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("docker build failed: %v\n%s", err, string(output))
// 	}

// 	return nil
// }