package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const HelperFlag = "--curb-update-helper"

// StartHelper acknowledges that a detached helper has read the staged plan.
// The caller can then quit; the helper waits for its PID before replacing files.
func StartHelper(plan *Plan) error {
	if err := plan.validate(); err != nil {
		return err
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	planFile := filepath.Join(plan.Work, "plan.json")
	if err := os.WriteFile(planFile, data, 0o600); err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	helper := filepath.Join(plan.Work, "helper"+filepath.Ext(executable))
	if err := copyExecutable(executable, helper); err != nil {
		return err
	}
	command := exec.Command(helper, HelperFlag, planFile)
	command.Env = cleanEnvironment()
	command.Dir = plan.Work
	configureProcess(command)
	if err := command.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-done:
			return errors.New("update helper could not start")
		case <-timer.C:
			command.Process.Kill()
			<-done
			return errors.New("update helper did not respond")
		case <-tick.C:
			if token, err := os.ReadFile(filepath.Join(plan.Work, "helper-ready")); err == nil && string(token) == plan.Token {
				return nil
			}
		}
	}
}

func copyExecutable(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return err
	}
	_, err = io.Copy(output, input)
	if closeErr := output.Close(); err == nil {
		err = closeErr
	}
	return err
}

func loadPlan(path string) (*Plan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var plan Plan
	if err := json.NewDecoder(io.LimitReader(file, 64<<10)).Decode(&plan); err != nil {
		return nil, err
	}
	if filepath.Clean(path) != filepath.Join(plan.Work, "plan.json") {
		return nil, errors.New("unexpected update plan location")
	}
	if err := plan.validate(); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (p *Plan) validate() error {
	if !filepath.IsAbs(p.Target) || !filepath.IsAbs(p.Work) || !filepath.IsAbs(p.Staged) ||
		filepath.Clean(p.Work) != p.Work || filepath.Clean(p.Target) != p.Target || filepath.Dir(p.Target) == p.Target ||
		filepath.Dir(p.Work) != filepath.Dir(p.Target) || !strings.HasPrefix(filepath.Base(p.Work), ".curb-update-") {
		return errors.New("invalid update installation paths")
	}
	relative, err := filepath.Rel(p.Work, p.Staged)
	if err != nil || relative == "." || !filepath.IsLocal(relative) {
		return errors.New("update is outside staging")
	}
	if p.Kind != "appimage" && p.Kind != "bundle" && p.Kind != "directory" {
		return errors.New("unsupported installation")
	}
	if p.Kind != "appimage" && !filepath.IsLocal(p.Executable) {
		return errors.New("invalid app executable")
	}
	if len(p.Token) != 48 || p.ParentPID <= 0 {
		return errors.New("invalid update session")
	}
	_, err = parseVersion(p.Version)
	return err
}

// RunHelper is entered before the GUI starts. No shell or administrator access
// is needed to replace a user-writable installation.
func RunHelper(path string) error {
	plan, err := loadPlan(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(plan.Work, "helper-ready"), []byte(plan.Token), 0o600); err != nil {
		return err
	}
	defer func() {
		// A failed rollback must leave its recovery copy intact.
		if _, err := os.Stat(filepath.Join(plan.Work, "previous")); plan.committed || os.IsNotExist(err) {
			cleanupWork(plan.Work)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for processRunning(plan.ParentPID) {
		select {
		case <-ctx.Done():
			return errors.New("original app did not exit; update canceled")
		case <-time.After(100 * time.Millisecond):
		}
	}
	return apply(plan, 30*time.Second)
}

// apply keeps the old application until the new UI sends its startup receipt.
// Failed replacements and failed launches restore and reopen the old app.
func apply(plan *Plan, readyTimeout time.Duration) error {
	if plan.Kind == "directory" {
		if _, err := readMarker(plan.Target, plan.Platform, ""); err != nil {
			_, _ = launch(plan, false, err.Error())
			return err
		}
	}
	previous := filepath.Join(plan.Work, "previous")
	if err := os.Rename(plan.Target, previous); err != nil {
		_, _ = launch(plan, false, "Could not replace the installed app. Try again.")
		return err
	}
	if err := os.Rename(plan.Staged, plan.Target); err != nil {
		_ = os.Rename(previous, plan.Target)
		_, _ = launch(plan, false, "Could not install the update. The previous version was restored.")
		return err
	}
	command, err := launch(plan, true, "")
	if err == nil {
		err = waitForReady(plan, command, readyTimeout)
	}
	if err == nil {
		plan.committed = true
		return os.RemoveAll(previous)
	}
	failed := filepath.Join(plan.Work, "failed")
	if moveErr := os.Rename(plan.Target, failed); moveErr != nil {
		return fmt.Errorf("update failed and could not restore previous version: %w", moveErr)
	}
	if restoreErr := os.Rename(previous, plan.Target); restoreErr != nil {
		_ = os.Rename(failed, plan.Target)
		return fmt.Errorf("could not restore previous version: %w", restoreErr)
	}
	_, _ = launch(plan, false, "The update could not start. The previous version was restored.")
	return err
}

func launch(plan *Plan, receipt bool, failure string) (*exec.Cmd, error) {
	executable := plan.Target
	if plan.Kind != "appimage" {
		executable = filepath.Join(plan.Target, plan.Executable)
	}
	command := exec.Command(executable, plan.Arguments...)
	command.Dir, command.Env = filepath.Dir(plan.Target), cleanEnvironment()
	if plan.Kind == "appimage" {
		// Also support installations originally launched without FUSE. The
		// runtime handles extraction and waits for the GUI child to finish.
		command.Env = append(command.Env, "APPIMAGE_EXTRACT_AND_RUN=1", "APPIMAGELAUNCHER_DISABLE=1")
	}
	if receipt {
		command.Env = append(command.Env, "CURB_UPDATE_PLAN="+filepath.Join(plan.Work, "plan.json"))
	}
	if failure != "" {
		command.Env = append(command.Env, "CURB_UPDATE_ERROR="+failure)
	}
	configureApp(command)
	if err := command.Start(); err != nil {
		return nil, err
	}
	return command, nil
}

func waitForReady(plan *Plan, command *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-done:
			return errors.New("updated app exited before opening")
		case <-timer.C:
			_ = command.Process.Kill()
			<-done
			return errors.New("updated app did not open in time")
		case <-tick.C:
			if token, err := os.ReadFile(filepath.Join(plan.Work, "app-ready")); err == nil && string(token) == plan.Token {
				return nil
			}
		}
	}
}

// Acknowledge is called only after MyGo has opened the new application's window.
func Acknowledge(version string) error {
	path := os.Getenv("CURB_UPDATE_PLAN")
	os.Unsetenv("CURB_UPDATE_PLAN")
	if path == "" {
		return nil
	}
	plan, err := loadPlan(path)
	if err != nil {
		return err
	}
	if version != plan.Version {
		return errors.New("updated app has an unexpected version")
	}
	return os.WriteFile(filepath.Join(plan.Work, "app-ready"), []byte(plan.Token), 0o600)
}

func cleanEnvironment() []string {
	environment := os.Environ()
	original, bundled := os.LookupEnv("CURB_ORIGINAL_LD_LIBRARY_PATH")
	var result []string
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		switch name {
		case "APPIMAGE", "APPDIR", "ARGV0", "OWD", "CURB_BUNDLED_RUNTIME", "CURB_RUNTIME_LIBDIR", "CURB_ORIGINAL_LD_LIBRARY_PATH",
			"CURB_UPDATE_PLAN", "CURB_UPDATE_ERROR", "FONTCONFIG_FILE", "GIO_MODULE_DIR", "GTK_IM_MODULE_FILE", "GDK_PIXBUF_MODULE_FILE", "GSETTINGS_SCHEMA_DIR":
			continue
		case "LD_LIBRARY_PATH":
			if bundled {
				continue
			}
		}
		result = append(result, entry)
	}
	if bundled && original != "" {
		result = append(result, "LD_LIBRARY_PATH="+original)
	}
	return result
}
