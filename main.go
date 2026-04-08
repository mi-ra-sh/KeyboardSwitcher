package main

import "os"

func main() {
	// Convert loop handles Ctrl+` events from the hook
	go convertLoop()

	// Keyboard hook on a separate OS thread
	go startHook()

	// Tray icon + message pump on main thread (blocks)
	startTray()

	os.Exit(0)
}
