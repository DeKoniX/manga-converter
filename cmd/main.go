package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dekonix/manga-converter/internal"
	"github.com/fsnotify/fsnotify"
)

// how long a file's size must remain unchanged before we process it
const stableWindow = 2 * time.Second
const checkInterval = 300 * time.Millisecond

func main() {
	log.SetOutput(os.Stdout)
	log.Println("🚀 Manga converter (fsnotify) started")

	inputDir := "input"

	// One-time scan on startup (in case files already exist)
	scanExisting(inputDir)

	// Start fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("watcher error: %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(inputDir); err != nil {
		log.Fatalf("cannot watch %s: %v", inputDir, err)
	}

	// debounce map per filename -> timer
	var mu sync.Mutex
	timers := map[string]*time.Timer{}

	scheduleProcess := func(path string) {
		// we only care about .zip files
		if !internal.IsZip(filepath.Base(path)) {
			return
		}
		mu.Lock()
		if t, ok := timers[path]; ok {
			// reset the existing timer
			t.Reset(stableWindow)
			mu.Unlock()
			return
		}
		// create a new timer that fires after stableWindow
		t := time.AfterFunc(stableWindow, func() {
			// wait until file size is stable
			if waitStable(path, stableWindow) {
				name := filepath.Base(path)
				log.Printf("📦 Обработка файла: %s", name)
				if err := internal.ProcessZip(name); err != nil {
					log.Printf("❌ Ошибка при обработке %s: %v", name, err)
				} else {
					log.Printf("✅ Успешно обработано: %s", name)
				}
			}
			// cleanup timer entry
			mu.Lock()
			delete(timers, path)
			mu.Unlock()
		})
		timers[path] = t
		mu.Unlock()
	}

	log.Println("👂 Watching for new .zip files in input/")

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			// React to create, write, rename, chmod
			if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Chmod) != 0 {
				scheduleProcess(ev.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			// Single-line error, no spammy loops
			log.Printf("⚠️ watcher error: %v", err)
		}
	}
}

func scanExisting(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("⚠️ Ошибка чтения %s: %v", dir, err)
		return
	}
	for _, f := range entries {
		if f.IsDir() || !internal.IsZip(f.Name()) {
			continue
		}
		log.Printf("🔎 Найден существующий архив: %s", f.Name())
		if err := internal.ProcessZip(f.Name()); err != nil {
			log.Printf("❌ Ошибка при обработке %s: %v", f.Name(), err)
		} else {
			log.Printf("✅ Успешно обработано: %s", f.Name())
		}
	}
}

func fileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// waitStable waits until file size is unchanged over the stableWindow period
func waitStable(path string, window time.Duration) bool {
	deadline := time.Now().Add(window)
	lastSize := int64(-1)
	for time.Now().Before(deadline) {
		sz, err := fileSize(path)
		if err != nil {
			// file might be moved/removed; abort
			return false
		}
		if sz == lastSize {
			// size unchanged for one interval; keep accumulating until deadline
		} else {
			// reset window if size changed
			deadline = time.Now().Add(window)
			lastSize = sz
		}
		time.Sleep(checkInterval)
	}
	return true
}
