package main

import (
	"runtime"
	"syscall"
	"unsafe"
)

var trayHwnd uintptr

// createTrayIcon creates a simple 32x32 icon programmatically.
func createTrayIcon() uintptr {
	const size = 32

	// Create a DIB section for the color bitmap
	var bmi BITMAPINFO
	bmi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmi.BmiHeader))
	bmi.BmiHeader.BiWidth = size
	bmi.BmiHeader.BiHeight = -size // top-down
	bmi.BmiHeader.BiPlanes = 1
	bmi.BmiHeader.BiBitCount = 32

	dc, _, _ := procCreateCompatibleDC.Call(0)
	var bits uintptr
	hbmColor, _, _ := procCreateDIBSection.Call(
		dc,
		uintptr(unsafe.Pointer(&bmi)),
		0, // DIB_RGB_COLORS
		uintptr(unsafe.Pointer(&bits)),
		0, 0,
	)
	procDeleteDC.Call(dc)

	if hbmColor == 0 || bits == 0 {
		return 0
	}

	// Fill pixels: blue square with "KS" feel
	pixels := unsafe.Slice((*[4]byte)(unsafe.Pointer(bits)), size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			i := y*size + x
			// Blue background (#2196F3)
			pixels[i] = [4]byte{0xF3, 0x96, 0x21, 0xFF} // BGRA

			// White border (1px)
			if x == 0 || x == size-1 || y == 0 || y == size-1 {
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}

			// Simple "K" shape (columns 6-14, rows 8-24)
			if x >= 7 && x <= 9 && y >= 8 && y <= 23 {
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF} // vertical bar
			}
			// K diagonals
			dy := y - 15
			if dy < 0 {
				dy = -dy
			}
			if x >= 10 && x <= 14 && y >= 8 && y <= 23 && (x-10) == dy/2 {
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}

			// Simple "S" shape (columns 17-25, rows 8-24)
			if y >= 8 && y <= 10 && x >= 18 && x <= 25 { // top bar
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}
			if y >= 11 && y <= 14 && x >= 17 && x <= 19 { // left middle-top
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}
			if y >= 14 && y <= 16 && x >= 18 && x <= 24 { // middle bar
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}
			if y >= 17 && y <= 20 && x >= 23 && x <= 25 { // right middle-bottom
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}
			if y >= 21 && y <= 23 && x >= 17 && x <= 24 { // bottom bar
				pixels[i] = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
			}
		}
	}

	// Mask bitmap (all zeros = fully opaque)
	maskSize := size * size / 8
	maskBits := make([]byte, maskSize)
	hbmMask, _, _ := procCreateBitmap.Call(size, size, 1, 1, uintptr(unsafe.Pointer(&maskBits[0])))

	// Create icon
	var ii ICONINFO
	ii.FIcon = 1
	ii.HbmMask = hbmMask
	ii.HbmColor = hbmColor

	hIcon, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))

	procDeleteObject.Call(hbmColor)
	procDeleteObject.Call(hbmMask)

	return hIcon
}

// trayWndProc handles messages for the tray's hidden window.
func trayWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		if lParam == WM_RBUTTONUP || lParam == WM_LBUTTONUP {
			showTrayMenu(hwnd)
		}
		return 0
	case WM_COMMAND:
		if wParam == IDM_QUIT {
			removeTrayIcon(hwnd)
			procDestroyWindow.Call(hwnd)
			return 0
		}
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

func showTrayMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}

	// "KeyboardSwitcher" label (grayed)
	procAppendMenuW.Call(hMenu, MF_STRING|MF_GRAYED, 0, uintptr(unsafe.Pointer(utf16Ptr("KeyboardSwitcher v1.0"))))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(hMenu, MF_STRING|MF_GRAYED, 0, uintptr(unsafe.Pointer(utf16Ptr("Ctrl+` — конвертувати"))))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(hMenu, MF_STRING, IDM_QUIT, uintptr(unsafe.Pointer(utf16Ptr("Вихід"))))

	// Get cursor position
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	procSetForegroundWindow.Call(hwnd)
	procTrackPopupMenu.Call(hMenu, TPM_BOTTOMALIGN|TPM_LEFTALIGN, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	procDestroyMenu.Call(hMenu)
}

func addTrayIcon(hwnd, hIcon uintptr) {
	var nid NOTIFYICONDATAW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = 1
	nid.UFlags = NIF_ICON | NIF_TIP | NIF_MESSAGE
	nid.UCallbackMessage = WM_TRAYICON
	nid.HIcon = hIcon

	tip := syscall.StringToUTF16("KeyboardSwitcher — Ctrl+`")
	copy(nid.SzTip[:], tip)

	procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
}

func removeTrayIcon(hwnd uintptr) {
	var nid NOTIFYICONDATAW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = 1
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

// startTray creates the hidden window, tray icon, and runs the message pump.
// This is the main goroutine — it blocks.
func startTray() {
	runtime.LockOSThread()

	hInstance := getModuleHandle()
	className := utf16Ptr("KeyboardSwitcherClass")

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   syscall.NewCallback(trayWndProc),
		HInstance:     hInstance,
		LpszClassName: uintptr(unsafe.Pointer(className)),
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("KeyboardSwitcher"))),
		0, 0, 0, 0, 0, 0, 0, hInstance, 0,
	)
	trayHwnd = hwnd

	hIcon := createTrayIcon()
	addTrayIcon(hwnd, hIcon)

	// Message pump for the tray window
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	if hIcon != 0 {
		procDestroyIcon.Call(hIcon)
	}
}
