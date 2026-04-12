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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

func Unarchive(ctx context.Context, archivePath string, compressionFormat archives.Compression,
	archiveFormat archives.Extraction, tmpFolder string) (string, error) {

	// Use archives to unpack this into the temporary folder (includes path within archive)
	format := archives.CompressedArchive{
		Compression: compressionFormat,
		Extraction:  archiveFormat,
	}

	//archiveStat, err := os.Stat(archivePath)
	//if err != nil {
	//	fmt.Println("Error inspecting archive file:", archivePath, "error:", err)
	//	return "", err
	//}

	// Note: The following count via unarchive absolutely eats CPU time - seems to read the archive 3 times!
	//completeSize := archiveStat.Size()
	//completeSize := int64(0)
	//fsys, err := archives.FileSystem(ctx, archivePath, nil)
	//if err != nil {
	//	return "", err
	//}
	//err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//	if path == ".git" {
	//		return fs.SkipDir
	//	}
	//	if !d.IsDir() {
	//		fi, err := d.Info()
	//		if err != nil {
	//			completeSize += fi.Size()
	//		}
	//	}
	//	return nil
	//})
	//if err != nil {
	//	return "", err
	//}
	//fmt.Println(" - complete size in MB:", strconv.Itoa(int(completeSize/(1024*1024))))

	fh, err := os.Open(archivePath)
	if err != nil {
		fmt.Println("Error opening archive file:", err)
		return "", err
	}
	defer fh.Close()

	outAbs, err := filepath.Abs(tmpFolder)
	if err != nil {
		fmt.Println(fmt.Errorf("calling filepath.Abs on output dir '%s' failed: %w", tmpFolder, err))
		return "", err
	}
	type Symlink struct {
		Source string
		Target string
	}
	//totalSize := int64(0)
	//lastTen := int64(0)
	//fmt.Print(" - Progress: 0%")
	var symlinks []Symlink
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
			if info.Mode()&os.ModeSymlink != 0 {
				//fmt.Println("******** PROCESSING SYMLINK FROM", destPath, "TO", fi.LinkTarget)
				// Symlink
				symlinks = append(symlinks, Symlink{
					Source: destPath,
					Target: fi.LinkTarget,
				})
				return nil
			}
			if info.IsDir() {
				return os.MkdirAll(destAbs, 0o755)
			}

			//totalSize += fi.Size()
			//curTen := int64(10 * math.Floor(float64(10*totalSize/completeSize)))
			//if curTen != lastTen {
			//	fmt.Print("..." + strconv.Itoa(int(curTen))) // We know it's small enough at this point
			//	lastTen = curTen
			//}

			// Ensure parent directories exist
			if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
				//fmt.Println("...Abandoned...")
				return fmt.Errorf("mkdir on parent '%s' failed: %w", destAbs, err)
			}

			// Open archive entry for reading
			rc, err := fi.Open()
			if err != nil {
				//fmt.Println("...Abandoned...")
				return fmt.Errorf("open entry %q failed: %w", nameInArchive, err)
			}
			defer rc.Close()

			// Create destination file
			outFile, err := os.OpenFile(destAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				//fmt.Println("...Abandoned...")
				return fmt.Errorf("create on %q failed: %w", destAbs, err)
			}
			defer outFile.Close()

			// Copy contents
			if _, err := io.Copy(outFile, rc); err != nil {
				//fmt.Println("...Abandoned...")
				return fmt.Errorf("copy on %q failed: %w", nameInArchive, err)
			}

			return nil
		},
	)
	//if err != nil {
	//	fmt.Println("...Abandoned...")
	//	return "", err
	//}
	//fmt.Println("...100%")
	//fmt.Println(" - total size in MB:", strconv.Itoa(int(totalSize/(1024*1024))))

	for _, symlink := range symlinks {
		err = os.Symlink(symlink.Target, symlink.Source)
		if err != nil {
			return "", fmt.Errorf("symlink from %q to %q, failed: %w", symlink.Source, symlink.Target, err)
		}
	}

	return tmpFolder, nil
}

func GetFirstSubfolder(ctx context.Context, folderPath string) (string, error) {
	// Read the first (and only) item in the archive to get the package name and version - DO NOT use the filename
	folderName := ""
	fsys, err := archives.FileSystem(ctx, folderPath, nil)
	if err != nil {
		fmt.Println("Error reading kodpkg contents:", err)
		return "", err
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		fmt.Println("Error reading kodpkg root folder:", err)
		return "", err
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
	return folderName, nil
}

func GetFirstSubfolderFast(packagePath string) (string, error) {
	// Get root folder first (assuming there's always 1 folder)
	listExec := exec.Command("bash", "-c", "tar --list --no-recursion -f "+packagePath+" | head -n 1")
	//fmt.Println("Executing", listExec.String())
	var listOutput SaveOutput
	listOutput.NoEchoToStdOut = true
	listExec.Stdin = os.Stdin
	listExec.Stdout = &listOutput
	listExec.Stderr = os.Stderr
	err := listExec.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(listOutput.String()), nil
}

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
	folderName, err := GetFirstSubfolderFast(packagePath)
	if err != nil {
		os.Exit(1)
	}
	if folderName == "" {
		fmt.Println("Archive doesn't have a root package folder. Not a valid kodpkg file. Exiting.")
		os.Exit(1)
	}

	tmpDir := os.TempDir()

	subPath := filepath.Join(tmpDir, folderName)
	_, err = os.Stat(subPath)
	if err == nil {
		if deleteIfExists {
			fmt.Println("Temporary folder", subPath, "already exists. Deleting.")
			err = os.RemoveAll(subPath)
			if err != nil {
				fmt.Println("Error deleting existing temporary folder. Folder:", subPath, "Error:", err)
				return "", nil
			}
		} else {
			fmt.Println("Temporary folder", subPath, "already exists. Skipping re-unpacking of archive.")
			return subPath, nil
		}
	}

	//tmpFolder, err := Unarchive(ctx, packagePath, archives.Xz{}, archives.Tar{}, tmpDir)
	err = UnarchiveTarXzFast(packagePath, tmpDir)
	if err != nil {
		fmt.Println("Error unpacking kodpkg package:", err)
		os.Exit(1)
	}
	tmpFolder := filepath.Join(tmpDir, folderName)
	fmt.Println(fmt.Sprintf("Package '%s' for target '%s' unpacked to temporary folder '%s'", packagePath, folderName, tmpFolder))
	return tmpFolder, nil
}

func UnarchiveTarXzFast(packagePath string, tmpDir string) error {
	tarExec := exec.Command("tar", "xf", packagePath, "-C", tmpDir, "--checkpoint=1000", "--checkpoint-action=dot")
	//fmt.Println("Executing", tarExec.String())
	var tarOutput SaveOutput
	//tarOutput.Prefix = "\xF0\x9F\x93\x82 " // don't do this as we'll get one per 'dot' in the progress bar!
	fmt.Print("  \xF0\x9F\x93\x82 ")
	tarExec.Stdin = os.Stdin
	tarExec.Stdout = &tarOutput
	tarExec.Stderr = os.Stderr
	err := tarExec.Run()
	if err != nil {
		return fmt.Errorf("unpacking kodpkg archive failed: %s", tarOutput.String())
	}
	return nil
}
