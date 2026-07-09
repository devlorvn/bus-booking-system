package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	var protoFiles []string
	err := filepath.Walk("proto", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking proto directory: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Println("No proto files found")
		os.Exit(1)
	}

	args := []string{
		"--proto_path=.",
		"--go_out=.",
		"--go_opt=paths=source_relative",
		"--go-grpc_out=.",
		"--go-grpc_opt=paths=source_relative",
	}
	args = append(args, protoFiles...)

	cmd := exec.Command("protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Running protoc compiler...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running protoc: %v\n", err)
		fmt.Println("\nTip: Please ensure 'protoc', 'protoc-gen-go', and 'protoc-gen-go-grpc' are installed and in your system PATH.")
		os.Exit(1)
	}

	fmt.Println("Protobuf code generated successfully!")
}
