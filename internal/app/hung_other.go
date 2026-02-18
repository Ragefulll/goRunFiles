//go:build !windows

package app

func isProcessHung(pid int) bool { return false }
