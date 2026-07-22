//go:build !linux && !darwin

package service

func init() {
	runtimeGOOS = func() string { return "other" }
}
