package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "text/template"
)

const defaultOutputDir = "_outputs"

func build(workDir string, config Config) error {
    templatePath := filepath.Join(workDir, config.Template)
    outputPath, err := resolveOutput(workDir, config.Template)
    if err != nil { return err }

    tpl, err := template.ParseFiles(templatePath)
    if err != nil { return fmt.Errorf("cannot parse template %s: %w", templatePath, err) }

    if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
        return fmt.Errorf("cannot create output directory: %w", err)
    }

    return render(tpl, outputPath, config.Data)
}

func resolveOutput(workDir, templateName string) (string, error) {
    name := filepath.Base(templateName)
    if !strings.HasSuffix(name, ".tpl") {
        return "", fmt.Errorf("template must end in .tpl: %s", templateName)
    }

    name = strings.TrimSuffix(name, ".tpl")
    return filepath.Join(workDir, defaultOutputDir, name), nil
}

func render(tpl *template.Template, outputPath string, data map[string]any) error {
    output, err := os.Create(outputPath)
    if err != nil { return fmt.Errorf("cannot create %s: %w", outputPath, err) }
    defer output.Close()

    if err := tpl.Execute(output, data); err != nil {
        return fmt.Errorf("cannot render %s: %w", outputPath, err)
    }
    return nil
}
