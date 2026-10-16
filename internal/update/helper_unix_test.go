//go:build !windows

package update

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadyChild(t *testing.T) {
	if os.Getenv("CURB_TEST_UPDATE_CHILD") != "1" {
		return
	}
	if err := Acknowledge("0.2.0"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
}

func testPlan(t *testing.T) *Plan {
	t.Helper()
	root := t.TempDir()
	work, err := os.MkdirTemp(root, ".curb-update-")
	if err != nil {
		t.Fatal(err)
	}
	return &Plan{Installation: Installation{Kind: "appimage", Target: filepath.Join(root, "Curb.AppImage"), Platform: "linux-amd64"},
		Work: work, Staged: filepath.Join(work, "new"), Version: "0.2.0", Token: strings.Repeat("a", 48), ParentPID: os.Getpid()}
}

func writePlan(t *testing.T, plan *Plan) {
	t.Helper()
	data, _ := json.Marshal(plan)
	if err := os.WriteFile(filepath.Join(plan.Work, "plan.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestApplyCommitsOnlyAfterNewProcessAcknowledges(t *testing.T) {
	plan := testPlan(t)
	if err := os.WriteFile(plan.Target, []byte("old app"), 0o755); err != nil {
		t.Fatal(err)
	}
	executable, _ := os.Executable()
	if err := copyExecutable(executable, plan.Staged); err != nil {
		t.Fatal(err)
	}
	plan.Arguments = []string{"-test.run=^TestReadyChild$"}
	t.Setenv("CURB_TEST_UPDATE_CHILD", "1")
	writePlan(t, plan)
	if err := apply(plan, 5*time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(plan.Work, "app-ready")); err != nil {
		t.Fatal("new process did not acknowledge startup")
	}
	if _, err := os.Stat(filepath.Join(plan.Work, "previous")); !os.IsNotExist(err) {
		t.Fatal("old application was not removed")
	}
	data, _ := os.ReadFile(plan.Target)
	if string(data) == "old app" {
		t.Fatal("new application was not installed")
	}
}

func TestFailedLaunchRollsBackAndReopensOldApp(t *testing.T) {
	plan := testPlan(t)
	old := []byte("#!/bin/sh\nprintf restored > \"$0.restored\"\n")
	if err := os.WriteFile(plan.Target, old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.Staged, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	writePlan(t, plan)
	if err := apply(plan, 2*time.Second); err == nil {
		t.Fatal("failed app reported success")
	}
	data, _ := os.ReadFile(plan.Target)
	if string(data) != string(old) {
		t.Fatal("previous application was not restored")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(plan.Target + ".restored"); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("previous application was not reopened")
}

func TestHelperChild(t *testing.T) {
	path := os.Getenv("CURB_TEST_HELPER_PLAN")
	if path == "" {
		return
	}
	if err := RunHelper(path); err != nil {
		t.Fatal(err)
	}
}

func TestHelperWaitsForOldProcessAndCleansStaging(t *testing.T) {
	plan := testPlan(t)
	if err := os.WriteFile(plan.Target, []byte("old app"), 0o755); err != nil {
		t.Fatal(err)
	}
	executable, _ := os.Executable()
	if err := copyExecutable(executable, plan.Staged); err != nil {
		t.Fatal(err)
	}
	old := exec.Command("sh", "-c", "sleep 0.4")
	if err := old.Start(); err != nil {
		t.Fatal(err)
	}
	plan.ParentPID, plan.Arguments = old.Process.Pid, []string{"-test.run=^TestReadyChild$"}
	writePlan(t, plan)
	helper := exec.Command(executable, "-test.run=^TestHelperChild$")
	helper.Env = append(os.Environ(), "CURB_TEST_HELPER_PLAN="+filepath.Join(plan.Work, "plan.json"), "CURB_TEST_UPDATE_CHILD=1")
	if err := helper.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	data, _ := os.ReadFile(plan.Target)
	if string(data) != "old app" {
		t.Fatal("helper replaced the app before its process exited")
	}
	_ = old.Wait()
	if err := helper.Wait(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(plan.Work); !os.IsNotExist(err) {
		t.Fatal("helper left old/staging files")
	}
}

func TestEnvironmentDoesNotCarryOldAppImageLibraries(t *testing.T) {
	t.Setenv("APPDIR", "/old/mount")
	t.Setenv("LD_LIBRARY_PATH", "/old/mount/usr/lib")
	t.Setenv("CURB_ORIGINAL_LD_LIBRARY_PATH", "/user/lib")
	t.Setenv("FONTCONFIG_FILE", "/old/mount/fonts.conf")
	environment := strings.Join(cleanEnvironment(), "\n")
	if strings.Contains(environment, "/old/mount") || !strings.Contains(environment, "LD_LIBRARY_PATH=/user/lib") {
		t.Fatal("old bundle resources leaked into the new process")
	}
}
