package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func Run(ctx context.Context, tools Tools, request Request, notify func(Update)) (string, error) {
	if err := os.MkdirAll(request.Destination, 0755); err != nil {
		return "", fmt.Errorf("Cannot create the download folder: %w", err)
	}
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
				}
				notify(u)
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
	return savedPath, nil
}
