package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ProcessZip(name string) error {
	zipPath := filepath.Join("input", name)
	workPath := filepath.Join("workdir", strings.TrimSuffix(name, ".zip"))

	log.Printf("📁 Распаковка архива: %s в %s", zipPath, workPath)
	err := Unzip(zipPath, workPath)
	if err != nil {
		return fmt.Errorf("распаковка: %w", err)
	}

	deferCleanup := func() {
		log.Printf("🧹 Удаление: %s и %s", zipPath, workPath)
		os.Remove(zipPath)
		os.RemoveAll(workPath)
	}

	mangaDirs, err := os.ReadDir(workPath)
	if err != nil {
		return fmt.Errorf("чтение каталога %s: %w", workPath, err)
	}
	if len(mangaDirs) == 0 {
		return fmt.Errorf("в архиве %s нет папок", name)
	}

	var mangaRoot string
	for _, dir := range mangaDirs {
		name := dir.Name()
		if dir.IsDir() && !strings.HasPrefix(name, "__MACOSX") && !strings.HasPrefix(name, ".") {
			mangaRoot = filepath.Join(workPath, name)
			break
		}
	}

	if mangaRoot == "" {
		return fmt.Errorf("не найдена валидная папка с мангой в архиве %s", name)
	}

	mangaName := filepath.Base(mangaRoot)
	log.Printf("🔍 Получение метаданных для: %s", mangaName)
	meta, err := FetchMetadata(mangaName)
	if err != nil {
		log.Printf("⚠️ Не удалось получить метаданные, продолжаем без них: %v", err)
		meta = &Metadata{Title: mangaName}
	}

	entries, err := os.ReadDir(mangaRoot)
	if err != nil {
		return fmt.Errorf("чтение манга-корня %s: %w", mangaRoot, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			volPath := filepath.Join(mangaRoot, entry.Name())
			if ContainsImages(volPath) {
				err := convertVolume(volPath, entry.Name(), mangaRoot, meta)
				if err != nil {
					log.Printf("❌ Ошибка тома %s: %v", entry.Name(), err)
				} else {
					log.Printf("✅ Том %s успешно обработан", entry.Name())
				}
			} else {
				log.Printf("⏭ Пропущен каталог (нет изображений): %s", volPath)
			}
		}
	}

	deferCleanup()

	return nil
}

func convertVolume(volumePath string, volumeName string, mangaRoot string, meta *Metadata) error {
	mangaName := filepath.Base(mangaRoot)
	cbzDir := filepath.Join("output/cbz", meta.Title)
	os.MkdirAll(cbzDir, os.ModePerm)
	outputBase := SafeName(fmt.Sprintf("%s__%s", mangaName, volumeName))

	volumeMeta := *meta
	volumeMeta.Title = fmt.Sprintf("%s — Том %s", meta.Title, volumeName)

	cbzOut := filepath.Join(cbzDir, outputBase+".cbz")
	// epubOut := filepath.Join("output/epub", outputBase+".epub")

	if err := CreateCBZ(volumePath, &volumeMeta, cbzOut); err != nil {
		return fmt.Errorf("ошибка CBZ: %w", err)
	}

	// if err := CreateEPUB(volumePath, &volumeMeta, epubOut); err != nil {
	// 	return fmt.Errorf("ошибка EPUB: %w", err)
	// }

	return nil
}

func IsZip(name string) bool {
	return strings.HasSuffix(name, ".zip")
}
