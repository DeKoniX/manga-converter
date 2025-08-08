package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func generateUUID() string {
	return uuid.New().String()
}

func CreateEPUB(folder string, meta *Metadata, output string) error {
	tempOPF := filepath.Join(folder, "metadata.opf")
	opf := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:description>%s</dc:description>
    <dc:subject>%s</dc:subject>
    <dc:source>%s</dc:source>
    <dc:identifier id="BookId">urn:uuid:%s</dc:identifier>
  </metadata>
  <manifest>
    <item id="cover" href="cover.jpg" media-type="image/jpeg"/>
    <item id="opf" href="metadata.opf" media-type="application/oebps-package+xml"/>
  </manifest>
  <spine>
  </spine>
</package>`, meta.Title, meta.Author, meta.Description, meta.Genres, meta.URL, generateUUID())

	err := os.WriteFile(tempOPF, []byte(opf), 0644)
	if err != nil {
		log.Printf("❌ Ошибка записи OPF файла: %v", err)
		return err
	}

	coverPath := filepath.Join(folder, "cover.jpg")
	if meta.CoverURL != "" {
		log.Printf("🖼 Загрузка обложки: %s", meta.CoverURL)
		err := DownloadFile(meta.CoverURL, coverPath)
		if err != nil {
			log.Printf("⚠️ Ошибка загрузки обложки: %v", err)
		}
	}

	outFile, err := os.Create(output)
	if err != nil {
		log.Printf("❌ Не удалось создать EPUB файл: %v", err)
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	log.Printf("📘 Упаковка EPUB: %s", output)

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("⚠️ Ошибка обхода файла: %v", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !isImage(info.Name()) && !strings.HasSuffix(info.Name(), ".opf") && !strings.HasSuffix(info.Name(), ".jpg") {
			return nil
		}

		relPath, _ := filepath.Rel(folder, path)
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			log.Printf("❌ Не удалось добавить файл в EPUB: %v", err)
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			log.Printf("❌ Не удалось открыть файл: %v", err)
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		log.Printf("❌ Ошибка упаковки EPUB: %v", err)
	} else {
		log.Printf("✅ EPUB создан: %s", output)
	}

	return err
}
