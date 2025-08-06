package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var filePaths []string
func FetchFiles(noDirs bool, generate string, dirName string, archiveName string) {
	exePath, _ := os.Executable()
	exeName := filepath.Base(exePath)
	files, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Println("[!] Error reading directory =>\n", err)
		return
	}

	fmt.Printf("[+] Fetching files in the supplied directory...\n")

	for _, file := range files {
		if noDirs {
			if file.IsDir() {
				fmt.Println("[*] Skipping directory:\n", file.Name())
				continue
			}
		}
		if file.Name() == exeName {
			continue
		}

		fullPath := filepath.Join(dirName, file.Name())
		filePaths = append(filePaths, fullPath)
	}

	ArchiveFiles(generate, dirName, archiveName)
}

func ArchiveFiles(generate string, dirName string, archiveName string) {
	currentPath := dirName
	fmt.Printf("[+] User has specified the directory name --> %v\n[*] User selected output path --> %v\n", generate, dirName)

	backupPath := filepath.Join(currentPath, generate)
	fmt.Printf("[DEBUG] Full path for backup folder --> %v\n", backupPath)

	if _, err := os.Stat(backupPath); err == nil {
		fmt.Println("[+] Backup directory already exists, skipping creation...")
	} else {
		err := os.Mkdir(backupPath, 0755)
		if err != nil {
			fmt.Printf("[!] Failed to create directory: %v\n", err)
			return
		}
		fmt.Printf("[+] Directory created successfully --> %v\n", backupPath)
	}

	for _, item := range filePaths {
		fmt.Printf("[+] Reading item from list ---> %v\n", item)
		destPath := filepath.Join(backupPath, filepath.Base(item))
		err := os.Rename(item, destPath)
		if err != nil {
			fmt.Printf("[!] Failed to move %v: %v\n", item, err)
		}
	}

	outFile, err := os.Create(archiveName)
	if err != nil {
		fmt.Printf("[!] Error creating archive file: %v\n", err)
		return
	}

	defer outFile.Close()

	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(backupPath), path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tarWriter, f)
		return err
	})

	if err != nil {
		fmt.Printf("[!] Error compressing files: %v\n", err)
	} else {
		fmt.Printf("[+] Successfully created archive: %s\n", archiveName)
	}
}

func main() {
	dirName, generate, noDirs, archiveName := parseFlags()
	fmt.Println("Directory:", dirName)
	fmt.Println("Generate Folder:", generate)
	fmt.Println("NoDirs:", noDirs)
	fmt.Println("Archive Name:", archiveName)

	FetchFiles(noDirs, generate, dirName, archiveName)
}

func parseFlags() (directory string, generateFolder string, noDirs bool, archiveNameFlag string) {
	directoryFlag := flag.String("dir", "", "Path to directory (required)")
	generateFolderFlag := flag.String("generate", "default_dir", "Name of the folder to generate for files")
	noDirsFlag := flag.Bool("nodir", false, "If true, subdirectories will be excluded")
	archiveFlag := flag.String("an", "archive.tar.gz", "Name of output compressed archive (Default: archive.tar.gz)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *directoryFlag == "" {
		fmt.Printf("[!] Error: -dir flag is required.\n")
		flag.Usage()
		os.Exit(1)
	}

	return *directoryFlag, *generateFolderFlag, *noDirsFlag, *archiveFlag
}
