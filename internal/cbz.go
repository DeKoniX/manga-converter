package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func CreateCBZ(folder string, meta *Metadata, output string) error {
	xml := fmt.Sprintf(`<ComicInfo>
  <Title>%s</Title>
  <Writer>%s</Writer>
  <Summary>%s</Summary>
  <Genre>%s</Genre>
  <Web>%s</Web>
</ComicInfo>`, meta.Title, meta.Author, meta.Description, meta.Genres, meta.URL)

	xmlPath := filepath.Join(folder, "ComicInfo.xml")
	err := os.WriteFile(xmlPath, []byte(xml), 0644)
	if err != nil {
		log.Printf("❌ Ошибка записи ComicInfo.xml: %v", err)
		return err
	}

	outFile, err := os.Create(output)
	if err != nil {
		log.Printf("❌ Не удалось создать CBZ файл: %v", err)
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	log.Printf("📦 Упаковка CBZ: %s", output)

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("⚠️ Ошибка обхода файла: %v", err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(folder, path)
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			log.Printf("❌ Не удалось создать файл в архиве: %v", err)
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
		log.Printf("❌ Ошибка упаковки CBZ: %v", err)
	} else {
		log.Printf("✅ CBZ создан: %s", output)
	}

	return err
}
