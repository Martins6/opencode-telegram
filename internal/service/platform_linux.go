//go:build linux

package service

func init() {
	runtimeGOOS = func() string { return "linux" }
}
