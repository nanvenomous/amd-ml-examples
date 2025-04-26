package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	dirData = "data"
)

var (
	hm      string
	configs []InputConfig
)

func init() {
	hm, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	configs = []InputConfig{
		{
			Directory:       filepath.Join(hm, "projects", "templ"),
			FileExtensions:  []string{".go", ".templ"},
			ExcludePatterns: []string{},
		},
		{
			Directory:       filepath.Join(hm, "projects", "tex"),
			FileExtensions:  []string{".go"},
			ExcludePatterns: []string{},
		},
		{
			Directory:       filepath.Join(hm, "projects", "app"),
			FileExtensions:  []string{".go", ".templ", ".yml", ".yaml", ".ts"},
			ExcludePatterns: []string{},
		},
		{
			Directory:       filepath.Join(hm, "projects", "cobra"),
			FileExtensions:  []string{".go"},
			ExcludePatterns: []string{},
		},
		{
			Directory:       filepath.Join(hm, "projects", "docker-cli"),
			FileExtensions:  []string{".go"},
			ExcludePatterns: []string{},
		},
	}

}

type InputConfig struct {
	Directory       string
	FileExtensions  []string
	ExcludePatterns []string
}

type CodebaseFile struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

type CodebaseData struct {
	Name  string         `yaml:"name"`
	Files []CodebaseFile `yaml:"files"`
}

func base64Encode(toEncode []byte) string {
	return base64.StdEncoding.EncodeToString(toEncode)
}

func matchesExcludePattern(path string, excludes []string) bool {
	for _, pattern := range excludes {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func getOutputFileName(config InputConfig) string {
	base := filepath.Base(config.Directory)
	return filepath.Join(dirData, base+".yaml")
}

func createDatasetFromCodebase(config InputConfig) error {
	var (
		outFile  = getOutputFileName(config)
		codeData = CodebaseData{
			Name:  filepath.Base(config.Directory),
			Files: []CodebaseFile{},
		}
	)

	// Create output file based on config
	output, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("Error creating output file: %v\n", err)
	}
	defer output.Close()

	// Walk through the directory based on config
	err = filepath.Walk(config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and files matching exclude patterns
		if info.IsDir() || matchesExcludePattern(path, config.ExcludePatterns) {
			return nil
		}

		ext := filepath.Ext(path)
		if slices.Contains(config.FileExtensions, ext) {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Error reading file %s: %v\n", path, err)

			}

			relPath, err := filepath.Rel(config.Directory, path)
			if err != nil {
				return err
			}

			content := base64Encode(data)
			codeData.Files = append(codeData.Files, CodebaseFile{
				Path:    relPath,
				Content: content,
			})

		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error walking through directory: %v\n", err)
	}

	codeDataByt, err := yaml.Marshal(&codeData)
	if err != nil {
		return fmt.Errorf("could not marshal collected code data to yaml: %v\n", err)
	}

	err = os.WriteFile(outFile, codeDataByt, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write code data bytes to file: %v\n", err)
	}

	fmt.Printf("Dataset created successfully at %s\n", outFile)
	return nil
}

func run() error {
	for _, cg := range configs {
		err := createDatasetFromCodebase(cg)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
