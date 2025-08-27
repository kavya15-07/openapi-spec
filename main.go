package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type ExternalDocs struct {
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
}

type Server struct {
	URL string `yaml:"url"`
}

type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type OpenAPISpec struct {
	OpenAPI      string                 `yaml:"openapi"`
	Info         Info                   `yaml:"info"`
	ExternalDocs ExternalDocs           `yaml:"externalDocs"`
	Servers      []Server               `yaml:"servers"`
	Tags         []Tag                  `yaml:"tags"`
	Paths        map[string]interface{} `yaml:"paths"`
	Components   map[string]interface{} `yaml:"components"`
}

type ComponentsWrapper struct {
	Components map[string]interface{} `yaml:"components"`
}

func loadYAMLFile(filePath string, target interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("error unmarshaling YAML from %s: %w", filePath, err)
	}
	return nil
}

func saveYAMLFile(filePath string, data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data to YAML: %w", err)
	}

	err = os.WriteFile(filePath, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}
	return nil
}

func bundleOpenAPIFiles(tagsFolder, componentsFile, outputFile string) error {
	bundled := OpenAPISpec{
		OpenAPI: "3.0.2",
		Info: Info{
			Title:       "Creo Collaboration service",
			Description: "API specification for Creo Collaboration service",
			Version:     "1.0.71",
		},
		ExternalDocs: ExternalDocs{
			Description: "Error Codes Documentation",
			URL:         "https://gitlab.rd-services.aws.ptc.com/creo/cgm/collabsvc/-/blob/master/errors/error_codes.go",
		},
		Servers: []Server{
			{URL: "https://creo.staging.atlas.ptc.com/collabsvc/api/cs"},
		},
		Tags: []Tag{
			{Name: "Sessions", Description: "Session endpoints"},
			{Name: "Branches", Description: "Branches endpoints"},
			{Name: "Checkpoints", Description: "Checkpoint endpoints"},
			{Name: "Chapters", Description: "Chapters endpoints"},
			{Name: "Comments", Description: "Comments endpoints"},
			{Name: "ConnectionSpeed", Description: "ConnectionSpeed endpoints"},
		},
		Paths:      make(map[string]interface{}),
		Components: make(map[string]interface{}),
	}

	// Read files directly from the tagsFolder
	files, err := os.ReadDir(tagsFolder)
	if err != nil {
		return fmt.Errorf("error reading tags folder %s: %w", tagsFolder, err)
	}

	for _, fileEntry := range files {
		if !fileEntry.IsDir() && filepath.Ext(fileEntry.Name()) == ".yaml" {
			filePath := filepath.Join(tagsFolder, fileEntry.Name())

			tagData := make(map[string]interface{})
			err := loadYAMLFile(filePath, &tagData)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				continue
			}

			for k, v := range tagData {
				bundled.Paths[k] = v
			}
		}
	}

	componentsData := ComponentsWrapper{}
	err = loadYAMLFile(componentsFile, &componentsData)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		fmt.Println("No valid components will be included.")
	} else {
		if len(componentsData.Components) > 0 {
			bundled.Components = componentsData.Components
		} else {
			fmt.Printf("Warning: 'components' section is empty or not found in %s\n", componentsFile)
		}
	}

	err = saveYAMLFile(outputFile, bundled)
	if err != nil {
		return fmt.Errorf("error saving bundled file: %w", err)
	}

	fmt.Printf("Successfully bundled OpenAPI spec to %s\n", outputFile)
	return nil
}

func main() {
	outputFileArg := flag.String("output", "collabsvc_bundled.yaml", "Path to the output bundled OpenAPI file")
	flag.Parse()

	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	tagsFolder := filepath.Join(baseDir, "tags")
	componentsFile := filepath.Join(baseDir, "components", "components.yaml")
	outputFile := filepath.Join(baseDir, *outputFileArg)

	err = bundleOpenAPIFiles(tagsFolder, componentsFile, outputFile)
	if err != nil {
		fmt.Printf("Bundling failed: %v\n", err)
		os.Exit(1)
	}
}
