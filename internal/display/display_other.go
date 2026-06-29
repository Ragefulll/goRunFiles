//go:build !windows

package display

func ListScreens() ([]Screen, error) {
	return nil, nil
}

func MoveProcessWindowToScreen(pid int, screenIndex int) error {
	return nil
}
