package main

import (
	"archive/tar"
	"context"
	"errors"
	"flag"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
)

func main() {
	// Command Line Arguments
	imageName := flag.String("image", "", "Docker image name (e.g., ubuntu:latest)")
	outputDir := flag.String("output", "", "Path to store the extracted rootfs (Default: $HOME/rootfs/<image-name>)")
	flag.Parse()

	if *imageName == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *outputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		*outputDir = filepath.Join(homeDir, "rootfs", *imageName)
	}

	err := os.MkdirAll(*outputDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Create the Docker Client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalln("Error creating Docker Client:", err)
	}

	// Pull the image if it doesn't exist
	log.Println("Pulling image:", *imageName)
	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, *imageName, image.PullOptions{})
	if err != nil {
		log.Fatalln("Error pulling image:", err)
	}

	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		log.Fatalln("Error pulling image:", err)
	}

	err = reader.Close()
	if err != nil {
		log.Fatalln("Error closing image:", err)
	}

	// create a temporary container from the image
	containerConfig := &container.Config{Image: *imageName}
	resp, err := cli.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")
	if err != nil {
		log.Fatalln("Error creating container:", err)
	}

	containerID := resp.ID
	defer cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})

	log.Println("Created container:", containerID)
	log.Println("Exporting rootfs...")

	exportReader, err := cli.ContainerExport(ctx, containerID)
	if err != nil {
		log.Fatalln("Error exporting image:", err)
	}
	defer exportReader.Close()

	// Extract tar archive to the specified directory
	err = extractTar(exportReader, *outputDir)
	if err != nil {
		log.Fatalln("Error extracting rootfs:", err)
	}

	log.Println("RootFS extracted successfully to:", *outputDir)
}

// extractTar extracts a tar archive to a given destination directory
func extractTar(r io.Reader, outputDir string) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(outputDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err = os.Stat(target); err != nil {
				if err = os.MkdirAll(target, 0755); err != nil {
					return err
				}
			} else {
				return err
			}
		case tar.TypeReg:
			var inFile, outFile *os.File
			inFile, err = os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err = os.Create(target)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err = io.Copy(outFile, tr); err != nil {
				return err
			}
		}
	}
}
