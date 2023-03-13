package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func UnzipArtifact(pipelineName string) error {
	// for all files in pipelineName folder
	// if file is zip, unzip it
	// if file is not zip, do nothing
	files, err := os.ReadDir(pipelineName)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", pipelineName, err)
	}

	for _, f := range files {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".zip") {
			fmt.Println("unzipping file ", fileName)
			err = unzip(pipelineName, fileName)
			if err != nil {
				return fmt.Errorf("failed to unzip file %s: %w", fileName, err)
			}
		}
	}
	return nil
}

func CleanupZipFiles(pipelineName string) error {
	files, err := os.ReadDir(pipelineName)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", pipelineName, err)
	}

	for _, f := range files {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".zip") {
			err := os.Remove(path.Join(pipelineName, fileName))
			if err != nil {
				return fmt.Errorf("error removing zip file %s: %w", path.Join(pipelineName, fileName), err)
			}
		}
	}
	return nil
}

func unzip(folderName, fileName string) error {
	archive, err := zip.OpenReader(filepath.Join(folderName, fileName))
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(folderName, f.Name)
		// fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(folderName)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path")
		}
		if f.FileInfo().IsDir() {
			// fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf(err.Error())
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return fmt.Errorf(err.Error())
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	return nil
}
