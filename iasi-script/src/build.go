package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func build(workDir string, descriptor Descriptor, config Config) error {
	if strings.TrimSpace(config.Template) == "" && len(config.Templates) == 0 {
		// A configuration without templates is valid. Some scripts may use only
		// configuration data or gain non-template actions later.
		return nil
	}
	if len(descriptor.Targets) == 0 {
		return fmt.Errorf("iasi.yml must define at least one target")
	}

	inputDir := resolveInputDir(workDir, descriptor)
	outputDir := resolveOutputDir(workDir, descriptor)

	if strings.TrimSpace(config.Template) != "" {
		name := artifactName(workDir, descriptor.Name)
		for _, target := range descriptor.Targets {
			if err := buildTemplate(inputDir, outputDir, name, config.Template, target, config.Data); err != nil {
				return err
			}
		}
		return nil
	}

	for name, templateName := range config.Templates {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("template artifact name is empty")
		}
		for _, target := range descriptor.Targets {
			if err := buildTemplate(inputDir, outputDir, name, templateName, target, config.Data); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildTemplate(inputDir, outputDir, name, templateName, target string, data map[string]any) error {
	templatePath, err := resolveTemplate(inputDir, templateName, target)
	if err != nil {
		return err
	}

	outputPath, err := resolveOutput(outputDir, name, target)
	if err != nil {
		return err
	}

	tpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("cannot parse template %s: %w", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	return render(tpl, outputPath, data)
}

func render(tpl *template.Template, outputPath string, data map[string]any) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", outputPath, err)
	}
	defer output.Close()

	if err := tpl.Execute(output, data); err != nil {
		return fmt.Errorf("cannot render %s: %w", outputPath, err)
	}
	return nil
}
