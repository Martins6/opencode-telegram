//go:build darwin

package service

func init() {
	runtimeGOOS = func() string { return "darwin" }
}
