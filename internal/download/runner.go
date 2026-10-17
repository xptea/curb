package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func Run(ctx context.Context, tools Tools, request Request, notify func(Update)) (string, error) {
	if err := os.MkdirAll(request.Destination, 0755); err != nil {
		return "", fmt.Errorf("Cannot create the download folder: %w", err)
	}
	destination, err := filepath.Abs(request.Destination)
	if err != nil {
		return "", err
	}
	workspace, err := os.MkdirTemp(destination, ".curb-download-*")
	if err != nil {
		return "", fmt.Errorf("Cannot prepare the download folder: %w", err)
	}
	defer os.RemoveAll(workspace)
	request.Destination = workspace
	cmd := exec.CommandContext(ctx, tools.Downloader, downloadArgs(tools, request)...)
	configureProcess(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		return "", fmt.Errorf("Could not start the download engine: %w", err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var savedPath, errorLine string
	var scanErr error
	scan := func(stream io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(stream)
		scanner.Buffer(make([]byte, 4096), 2*1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			mu.Lock()
			if u, ok := parseUpdate(line); ok {
				if u.Path != "" {
					savedPath = u.Path
				} else {
					notify(u)
				}
			}
			if strings.HasPrefix(line, "ERROR:") {
				errorLine = strings.TrimPrefix(line, "ERROR: ")
			}
			mu.Unlock()
		}
		if err := scanner.Err(); err != nil {
			mu.Lock()
			scanErr = err
			mu.Unlock()
		}
	}
	wg.Add(2)
	go scan(stdout)
	go scan(stderr)
	wg.Wait()
	err = cmd.Wait()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		if errorLine != "" {
			return "", errors.New(errorLine)
		}
		return "", fmt.Errorf("Download failed: %w", err)
	}
	if scanErr != nil {
		return "", fmt.Errorf("Could not read download progress: %w", scanErr)
	}
	if savedPath == "" {
		return "", errors.New("The download engine finished without returning a file.")
	}
	if info, err := os.Stat(savedPath); err != nil || !info.Mode().IsRegular() {
		return "", errors.New("The downloaded file could not be found.")
	}
	if relative, err := filepath.Rel(workspace, savedPath); err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("The download engine returned a file outside its workspace.")
	}
	path, err := publishOutput(ctx, savedPath, destination)
	if err != nil {
		return "", fmt.Errorf("Could not save the finished file: %w", err)
	}
	notify(Update{Path: path, Status: "Saved", Progress: 1})
	return path, nil
}
