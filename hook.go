package main

import (
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// Buffer stores all typed characters since the last Enter.
type Buffer struct {
	mu   sync.Mutex
	data []rune
}

func (b *Buffer) Add(ch rune) {
	b.mu.Lock()
	b.data = append(b.data, ch)
	if len(b.data) > 2000 {
		b.data = b.data[len(b.data)-2000:]
	}
	b.mu.Unlock()
}

func (b *Buffer) Pop() {
	b.mu.Lock()
	if len(b.data) > 0 {
		b.data = b.data[:len(b.data)-1]
	}
	b.mu.Unlock()
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	b.data = b.data[:0]
	b.mu.Unlock()
}

func (b *Buffer) Get() []rune {
	b.mu.Lock()
	cp := make([]rune, len(b.data))
	copy(cp, b.data)
	b.mu.Unlock()
	return cp
}

func (b *Buffer) Set(text []rune) {
	b.mu.Lock()
	b.data = append(b.data[:0], text...)
	b.mu.Unlock()
}

func (b *Buffer) Len() int {
	b.mu.Lock()
	n := len(b.data)
	b.mu.Unlock()
	return n
}

var (
	buffer     Buffer
	converting bool
	hookHandle uintptr
	convertCh  = make(chan struct{}, 1)
)

// hookCallback is the low-level keyboard hook procedure.
func hookCallback(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 {
		kb := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

		// Skip our own synthetic events
		if kb.DwExtraInfo == MAGIC_EXTRA {
			ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
			return ret
		}

		if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
			vk := kb.VkCode

			// Check modifiers
			ctrlDown, _, _ := procGetAsyncKeyState.Call(VK_CONTROL)
			altDown, _, _ := procGetAsyncKeyState.Call(VK_MENU)
			shiftDown, _, _ := procGetAsyncKeyState.Call(VK_SHIFT)
			ctrl := ctrlDown&0x8000 != 0
			alt := altDown&0x8000 != 0
			_ = shiftDown

			// Ctrl+` — trigger conversion
			if vk == VK_OEM_3 && ctrl && !alt {
				select {
				case convertCh <- struct{}{}:
				default:
				}
				return 1 // suppress the keystroke
			}

			// Skip if Ctrl or Alt held (shortcuts, not typing)
			if ctrl || alt {
				ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
				return ret
			}

			// Skip during our own conversion
			if converting {
				ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
				return ret
			}

			switch vk {
			case VK_RETURN, VK_TAB:
				buffer.Clear()
			case VK_BACK:
				buffer.Pop()
			case VK_SPACE:
				buffer.Add(' ')
			default:
				// Skip modifiers and special keys
				if vk >= 0x30 || vk == VK_OEM_3 { // printable range
					if ch, ok := vkToChar(vk, kb.ScanCode); ok {
						buffer.Add(ch)
					}
				}
			}
		}
	}

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

// doConvert replaces the buffered text with its EN↔UA conversion.
func doConvert() {
	text := buffer.Get()
	if len(text) == 0 {
		return
	}

	converted := convertText(text)

	// Check if anything actually changed
	same := true
	if len(converted) == len(text) {
		for i := range text {
			if text[i] != converted[i] {
				same = false
				break
			}
		}
	} else {
		same = false
	}
	if same {
		return
	}

	converting = true

	// Delete old text with backspaces
	for range text {
		sendBackspace()
		time.Sleep(2 * time.Millisecond)
	}

	time.Sleep(20 * time.Millisecond)

	// Type converted text
	for _, ch := range converted {
		sendUnicodeChar(ch)
		time.Sleep(2 * time.Millisecond)
	}

	buffer.Set(converted)
	converting = false
}

// startHook installs the keyboard hook and runs the message pump.
// Must be called from a goroutine with LockOSThread.
func startHook() {
	runtime.LockOSThread()

	hMod := getModuleHandle()

	cb := syscall.NewCallback(hookCallback)
	h, _, err := procSetWindowsHookExW.Call(
		WH_KEYBOARD_LL,
		cb,
		hMod,
		0,
	)
	if h == 0 {
		panic("SetWindowsHookEx failed: " + err.Error())
	}
	hookHandle = h

	// Message pump — required for LL hooks to work
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

	procUnhookWindowsHookEx.Call(hookHandle)
}

// convertLoop listens for convert requests and executes them.
func convertLoop() {
	for range convertCh {
		doConvert()
	}
}
