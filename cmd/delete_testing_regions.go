package main

import (
	"bufio"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	workspace, _ := os.Getwd()

	var srcFilesPath string
	flag.StringVar(&srcFilesPath, "src", "", "source files path")
	debugFlag := flag.Bool("debug", false, "show debug messages")
	flag.Parse()

	if srcFilesPath == "" {
		srcFilesPath = "src"
	}
	srcFilesPath = filepath.Join(workspace, srcFilesPath)

	logger := CreateLogger(debugFlag)

	logger.Info("start application")

	var wg sync.WaitGroup
	err := filepath.Walk(srcFilesPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".bsl") {
			wg.Add(1)
			go func(file string) {
				defer wg.Done()

				removeTestingRegion(file, logger)
			}(path)
		}
		return nil
	})
	if err != nil {
		logger.Info(err.Error())
	}
	wg.Wait()

	logger.Info("end application")
}

func removeTestingRegion(filePath string, logger *slog.Logger) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Info(err.Error())
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inTestingRegion := false
	isRemoveRegion := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "#Область Тестирование" {
			logger.Info("Removing testing region", "delete region",filePath)
			isRemoveRegion = true
			inTestingRegion = true
			continue
		}
		if line == "#КонецОбласти" && inTestingRegion {
			inTestingRegion = false
			continue
		}
		if !inTestingRegion {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Info(err.Error())
	}

	err = file.Close()
	if err != nil {
		logger.Info(err.Error())
	}

	if isRemoveRegion {
		err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			logger.Info(err.Error())
		}
	}
}

func CreateLogger(debugFlag *bool) *slog.Logger {
	var programLevel = new(slog.LevelVar)
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	logger := slog.New(stdoutHandler)
	
	slog.SetDefault(logger)

	if *debugFlag {
        programLevel.Set(slog.LevelDebug)
    } else {
        programLevel.Set(slog.LevelInfo)
    }
	return logger
}