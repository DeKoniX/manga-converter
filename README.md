# Manga Converter

Manga Converter — это утилита на Go для автоматической конвертации архивов с мангой (.zip) в форматы **CBZ** с добавлением метаданных.

## Возможности
- Мониторинг директории `input/` в реальном времени через `fsnotify`.
- Поддержка вложенной структуры: `manga_name/volume/*images*`.
- Получение метаданных с Shikimori (или fallback на имя архива).
- Создание структуры:
  ```
  output/cbz/<Название манги>/<Название манги>__<Том>.cbz\
  ```
- Обработка только стабильных файлов (ожидание окончания записи).
- Логирование в stdout (для Docker).

## Требования
- Go 1.22+
- Linux с поддержкой inotify (ext4, btrfs и т.п.)

## Установка
```bash
git clone https://github.com/dekonix/manga-converter.git
cd manga-converter
```

## Сборка
```bash
go build -o bin/converter ./cmd
```

## Запуск
```bash
./bin/converter
```

В Docker:
```bash
docker build -t manga-converter .
docker run --rm \
  -v $(pwd)/input:/app/input \
  -v $(pwd)/output:/app/output \
  manga-converter
```

## Структура проекта
```
cmd/              — точка входа (main.go)
internal/         — пакет с логикой (convert.go, cbz.go, epub.go, utils.go, shikimori.go)
Dockerfile        — сборка образа
``` 

## Настройка метаданных
По умолчанию метаданные загружаются с Shikimori. Если манга не найдена, используется имя архива.

## Лицензия
MIT