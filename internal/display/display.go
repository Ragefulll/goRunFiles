package display

// Screen describes a monitor available to the desktop session.
type Screen struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Primary bool   `json:"primary"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}
