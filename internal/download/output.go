package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// publishOutput creates the final name exclusively. Conversion happens in a
// private workspace so yt-dlp cannot reuse or delete an earlier download.
func publishOutput(ctx context.Context, source, destination string) (string, error) {
	name := filepath.Base(source)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for number := 0; number < 10000; number++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if number > 0 {
			name = fmt.Sprintf("%s (%d)%s", stem, number, ext)
		}
		target := filepath.Join(destination, name)
		if err := os.Link(source, target); err == nil {
			return target, nil
		} else if errors.Is(err, os.ErrExist) {
			continue
		}
		// FAT and some network filesystems do not support hard links. Keep the
		// same no-overwrite guarantee when copying bytes instead.
		err := copyOutput(ctx, source, target)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		return target, nil
	}
	return "", errors.New("Too many downloads have the same filename. Choose another save folder.")
}

func copyOutput(ctx context.Context, source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	finished := false
	defer func() {
		output.Close()
		if !finished {
			os.Remove(target)
		}
	}()
	buffer := make([]byte, 256*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, readErr := input.Read(buffer)
		if n > 0 {
			if written, err := output.Write(buffer[:n]); err != nil {
				return err
			} else if written != n {
				return io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if err := output.Close(); err != nil {
		return err
	}
	finished = true
	return nil
}
