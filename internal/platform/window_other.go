//go:build !linux

package platform

import "github.com/egoist/mygo"

func ConfigureWindow(*mygo.Window) {}

func PrepareApplication() {}
