/*
Copyright © 2026 Kod project Contributors
*/
package internal

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

func DoUnpack(ctx context.Context, packagePath string, deleteIfExists bool) (string, error) {

	// Verify that the package exists and is a file
	packageFile, err := os.Stat(packagePath)
	if err != nil {
		fmt.Println(packagePath, "does not exist")
		os.Exit(1)
	}
	if packageFile.IsDir() {
		fmt.Println(packagePath, "is a directory and not a regular file")
		os.Exit(1)
	}
	// Create temporary folder based on package name and version
	// Read the first (and only) item in the archive to get the package name and version - DO NOT use the filename
	folderName := ""
	fsys, err := archives.FileSystem(ctx, packagePath, nil)
	if err != nil {
		fmt.Println("Error reading kodpkg contents:", err)
		os.Exit(1)
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		fmt.Println("Error reading kodpkg root folder:", err)
		os.Exit(1)
	}
	for _, entry := range entries {
		// ignore .git and .DS_Store and any other archiving artifacts
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		//fmt.Println("Walking:", entry.Name(), "Dir?", entry.IsDir())
		if entry.IsDir() {
			folderName = entry.Name()
		}
	}
	if folderName == "" {
		fmt.Println("Archive doesn't have a root package folder. Not a valid kodpkg file. Exiting.")
		os.Exit(1)
	}

	// Force delete the folder if it already exists (unpack ALWAYS recreates the folder)
	tmpDir := os.TempDir()
	tmpFolder := filepath.Join(tmpDir, folderName)
	_, err = os.Stat(tmpFolder)
	if err == nil {
		if deleteIfExists {
			fmt.Println("Temporary package folder", tmpFolder, "already exists. Deleting.")
			err = os.RemoveAll(tmpFolder)
			if err != nil {
				fmt.Println("Error deleting existing temporary folder for unpacking. Folder:", tmpFolder, "Error:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Temporary package folder", tmpFolder, "already exists. Skipping re-unpacking of archive.")
			return tmpFolder, nil
		}
	}

	// Use archives to unpack this into the temporary folder (includes path within archive)
	format := archives.CompressedArchive{
		Compression: archives.Xz{},
		Extraction:  archives.Tar{},
	}
	fh, err := os.Open(packagePath)
	if err != nil {
		fmt.Println("Error opening kodpkg archive file:", err)
		os.Exit(1)
	}
	outAbs, err := filepath.Abs(tmpDir)
	if err != nil {
		fmt.Println(fmt.Errorf("calling filepath.Abs on output dir '%s' failed: %w", tmpDir, err))
		os.Exit(1)
	}
	err = format.Extract(ctx, fh,
		func(ctx context.Context, fi archives.FileInfo) error {
			nameInArchive := fi.NameInArchive

			if nameInArchive == "" || nameInArchive == "." {
				return nil
			}

			cleanName := filepath.Clean(nameInArchive)
			destPath := filepath.Join(outAbs, cleanName)

			destAbs, err := filepath.Abs(destPath)
			if err != nil {
				return fmt.Errorf("calling filepath.Abs on dest path '%s' failed: %w", destPath, err)
			}

			// Avoid traversal attacks
			if !strings.HasPrefix(destAbs, outAbs+string(os.PathSeparator)) && destAbs != outAbs {
				return fmt.Errorf("unsafe path in archive: %q", nameInArchive)
			}

			// Create directory if in archive
			info, err := fi.Stat()
			if err != nil {
				return fmt.Errorf("stat on %q failed: %w", nameInArchive, err)
			}
			if info.IsDir() {
				return os.MkdirAll(destAbs, 0o755)
			}

			// Ensure parent directories exist
			if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
				return fmt.Errorf("mkdir on parent '%s' failed: %w", destAbs, err)
			}

			// Open archive entry for reading
			rc, err := fi.Open()
			if err != nil {
				return fmt.Errorf("open entry %q failed: %w", nameInArchive, err)
			}
			defer rc.Close()

			// Create destination file
			outFile, err := os.OpenFile(destAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return fmt.Errorf("create on %q failed: %w", destAbs, err)
			}
			defer outFile.Close()

			// Copy contents
			if _, err := io.Copy(outFile, rc); err != nil {
				return fmt.Errorf("copy on %q failed: %w", nameInArchive, err)
			}

			return nil
		})
	if err != nil {
		fmt.Println("Error unpacking kodpkg package:", err)
		os.Exit(1)
	}
	fmt.Println(fmt.Sprintf("Package '%s' for target '%s' unpacked to temporary folder '%s'", packagePath, folderName, tmpFolder))
	return tmpFolder, nil
}
