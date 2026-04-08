package main

import (
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	// Keyboard hook
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetKeyState         = user32.NewProc("GetKeyState")

	// Character translation
	procGetForegroundWindow       = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	procGetKeyboardLayout         = user32.NewProc("GetKeyboardLayout")
	procToUnicodeEx               = user32.NewProc("ToUnicodeEx")

	// SendInput
	procSendInput = user32.NewProc("SendInput")

	// Module handle
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	// Tray
	procShellNotifyIconW  = shell32.NewProc("Shell_NotifyIconW")
	procCreateWindowExW   = user32.NewProc("CreateWindowExW")
	procRegisterClassExW  = user32.NewProc("RegisterClassExW")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procDestroyWindow     = user32.NewProc("DestroyWindow")
	procCreatePopupMenu   = user32.NewProc("CreatePopupMenu")
	procAppendMenuW       = user32.NewProc("AppendMenuW")
	procTrackPopupMenu    = user32.NewProc("TrackPopupMenu")
	procDestroyMenu       = user32.NewProc("DestroyMenu")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetCursorPos      = user32.NewProc("GetCursorPos")
	procCreateIconIndirect = user32.NewProc("CreateIconIndirect")
	procCreateCompatibleDC = user32.NewProc("CreateCompatibleDC")
	procDeleteDC          = user32.NewProc("DeleteDC")
	procDestroyIcon       = user32.NewProc("DestroyIcon")

	gdi32                    = syscall.NewLazyDLL("gdi32.dll")
	procCreateBitmap         = gdi32.NewProc("CreateBitmap")
	procCreateDIBSection     = gdi32.NewProc("CreateDIBSection")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procSelectObject         = gdi32.NewProc("SelectObject")
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	WM_DESTROY     = 0x0002
	WM_COMMAND     = 0x0111
	WM_APP         = 0x8000
	WM_TRAYICON    = WM_APP + 1
	WM_RBUTTONUP   = 0x0205
	WM_LBUTTONUP   = 0x0202

	VK_BACK    = 0x08
	VK_TAB     = 0x09
	VK_RETURN  = 0x0D
	VK_SHIFT   = 0x10
	VK_CONTROL = 0x11
	VK_MENU    = 0x12 // Alt
	VK_CAPITAL = 0x14
	VK_SPACE   = 0x20
	VK_OEM_3   = 0xC0 // ` ~ key

	INPUT_KEYBOARD      = 1
	KEYEVENTF_KEYUP     = 0x0002
	KEYEVENTF_UNICODE   = 0x0004

	NIM_ADD    = 0x00000000
	NIM_DELETE = 0x00000002
	NIF_ICON   = 0x00000002
	NIF_TIP    = 0x00000004
	NIF_MESSAGE = 0x00000001

	MF_STRING    = 0x00000000
	MF_SEPARATOR = 0x00000800
	MF_GRAYED    = 0x00000001

	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000

	MAGIC_EXTRA = 0x4B53 // "KS" marker for our synthetic events

	IDM_QUIT = 1001
)

// KBDLLHOOKSTRUCT from Win32 API
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// INPUT + KEYBDINPUT for SendInput
type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
	_pad        [8]byte // padding to match INPUT union size
}

type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
}

// MSG for GetMessage
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

// NOTIFYICONDATAW for Shell_NotifyIcon
type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

// WNDCLASSEXW for RegisterClassEx
type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  uintptr
	LpszClassName uintptr
	HIconSm      uintptr
}

type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func getModuleHandle() uintptr {
	h, _, _ := procGetModuleHandleW.Call(0)
	return h
}

func sendInputKey(vk uint16, scan uint16, flags uint32) {
	var input INPUT
	input.Type = INPUT_KEYBOARD
	input.Ki.WVk = vk
	input.Ki.WScan = scan
	input.Ki.DwFlags = flags
	input.Ki.DwExtraInfo = MAGIC_EXTRA
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))
}

func sendUnicodeChar(ch rune) {
	// Key down
	var down INPUT
	down.Type = INPUT_KEYBOARD
	down.Ki.WScan = uint16(ch)
	down.Ki.DwFlags = KEYEVENTF_UNICODE
	down.Ki.DwExtraInfo = MAGIC_EXTRA

	// Key up
	var up INPUT
	up.Type = INPUT_KEYBOARD
	up.Ki.WScan = uint16(ch)
	up.Ki.DwFlags = KEYEVENTF_UNICODE | KEYEVENTF_KEYUP
	up.Ki.DwExtraInfo = MAGIC_EXTRA

	inputs := [2]INPUT{down, up}
	procSendInput.Call(2, uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(inputs[0]))
}

func sendBackspace() {
	sendInputKey(VK_BACK, 0, 0)
	sendInputKey(VK_BACK, 0, KEYEVENTF_KEYUP)
}

func vkToChar(vkCode, scanCode uint32) (rune, bool) {
	// Get keyboard state for shift/capslock
	var keyState [256]byte

	shiftState, _, _ := procGetAsyncKeyState.Call(VK_SHIFT)
	if shiftState&0x8000 != 0 {
		keyState[VK_SHIFT] = 0x80
	}

	capsState, _, _ := procGetKeyState.Call(VK_CAPITAL)
	if capsState&0x0001 != 0 {
		keyState[VK_CAPITAL] = 0x01
	}

	// Get foreground window's keyboard layout
	hwnd, _, _ := procGetForegroundWindow.Call()
	tid, _, _ := procGetWindowThreadProcessId.Call(hwnd, 0)
	hkl, _, _ := procGetKeyboardLayout.Call(tid)

	var buf [4]uint16
	ret, _, _ := procToUnicodeEx.Call(
		uintptr(vkCode),
		uintptr(scanCode),
		uintptr(unsafe.Pointer(&keyState[0])),
		uintptr(unsafe.Pointer(&buf[0])),
		4,
		0,
		hkl,
	)

	if int32(ret) == 1 {
		return rune(buf[0]), true
	}
	// Dead key (-1) — call again to clear state
	if int32(ret) < 0 {
		procToUnicodeEx.Call(
			uintptr(vkCode), uintptr(scanCode),
			uintptr(unsafe.Pointer(&keyState[0])),
			uintptr(unsafe.Pointer(&buf[0])),
			4, 0, hkl,
		)
	}
	return 0, false
}
