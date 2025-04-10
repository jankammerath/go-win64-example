package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// User32
	messageBoxW      = user32.NewProc("MessageBoxW")
	createWindowExW  = user32.NewProc("CreateWindowExW")
	defWindowProcW   = user32.NewProc("DefWindowProcW")
	registerClassExW = user32.NewProc("RegisterClassExW")
	showWindow       = user32.NewProc("ShowWindow")
	updateWindow     = user32.NewProc("UpdateWindow")
	getMessageW      = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	dispatchMessageW = user32.NewProc("DispatchMessageW")
	postQuitMessage  = user32.NewProc("PostQuitMessage")
	loadCursorW      = user32.NewProc("LoadCursorW")

	// Kernel32
	getModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	// Constants
	nullHandle        syscall.Handle = 0
	nullPtr           unsafe.Pointer = nil
	className                        = "myWindowClass"
	windowTitle                      = "Go GUI"
	buttonText                       = "Hello World"
	messageBoxTitle                  = "Hello"
	messageBoxContent                = "Hello Windows"

	// Window Styles
	wsOverlappedWindow = 0x00CF0000
	wsVisible          = 0x10000000
	cwUseDefault       = -2147483648 //0x80000000

	// Message Box Styles
	mbOk = 0x00000000
	// Button Styles
	bsPushButton = 0x00000000

	// Control IDs
	idButtonHello = 1001
)

type (
	wndProc func(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr
)

type msg struct {
	hwnd     syscall.Handle
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

type point struct {
	x int32
	y int32
}

type wndClassEx struct {
	Size          uint32
	Style         uint32
	WndProc       uintptr
	ClsExtra      int32
	WndExtra      int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	MenuName      *uint16
	ClassName     *uint16
	hIconSm       syscall.Handle
}

func strToUint16Ptr(s string) *uint16 {
	utf16, err := syscall.UTF16FromString(s)
	if err != nil {
		panic(err)
	}
	return &utf16[0]
}

func messageBox(hwnd syscall.Handle, text, title string, style uint) int {
	textPtr := strToUint16Ptr(text)
	titlePtr := strToUint16Ptr(title)

	ret, _, _ := messageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(style),
	)
	return int(ret)
}

func createWindow(exStyle uint32, className, windowName string, style uint32, x, y, width, height int32, hWndParent, hMenu, hInstance syscall.Handle, lpParam unsafe.Pointer) (syscall.Handle, error) {
	classNamePtr := strToUint16Ptr(className)
	windowNamePtr := strToUint16Ptr(windowName)

	ret, _, err := createWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windowNamePtr)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(hWndParent),
		uintptr(hMenu),
		uintptr(hInstance),
		uintptr(lpParam),
	)

	if ret == 0 {
		return nullHandle, err
	}

	return syscall.Handle(ret), nil
}

func defWindowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := defWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		uintptr(wParam),
		uintptr(lParam),
	)
	return ret
}

func registerClassEx(wcx *wndClassEx) (uint16, error) {
	ret, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(wcx)))
	if ret == 0 {
		return 0, err
	}
	return uint16(ret), nil
}

func showTheWindow(hwnd syscall.Handle, cmdShow int) bool {
	ret, _, _ := showWindow.Call(
		uintptr(hwnd),
		uintptr(cmdShow))

	return ret != 0
}

func updateTheWindow(hwnd syscall.Handle) bool {
	ret, _, _ := updateWindow.Call(uintptr(hwnd))
	return ret != 0
}

func getTheMessage(msg *msg, hwnd syscall.Handle, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := getMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))

	return int(ret)
}

func translateTheMessage(msg *msg) bool {
	ret, _, _ := translateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return ret != 0
}

func dispatchTheMessage(msg *msg) uintptr {
	ret, _, _ := dispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
	return ret
}

func postAQuitMessage(exitCode int) {
	postQuitMessage.Call(uintptr(exitCode))
}

func getTheModuleHandle(moduleName string) syscall.Handle {
	var moduleNamePtr *uint16
	if moduleName == "" {
		moduleNamePtr = nil // Explicitly use NULL to get current module
	} else {
		moduleNamePtr = strToUint16Ptr(moduleName)
	}

	ret, _, err := getModuleHandleW.Call(uintptr(unsafe.Pointer(moduleNamePtr)))

	if ret == 0 {
		fmt.Fprintf(os.Stderr, "GetModuleHandle failed: %v\n", err)
	}

	return syscall.Handle(ret)
}

func wndProcCallback(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	const (
		wmCreate  = 0x0001
		wmCommand = 0x0111
		wmDestroy = 0x0002
		bnClicked = 0
	)

	// Log all messages for debugging
	fmt.Fprintf(os.Stderr, "WndProc: hwnd=%v, msg=0x%04x, wParam=0x%08x, lParam=0x%08x\n", hwnd, msg, wParam, lParam)

	switch msg {
	case wmCreate:
		fmt.Fprintf(os.Stderr, "Window created successfully\n")
		return 0
	case wmCommand:
		controlID := wParam & 0xFFFF
		notificationCode := (wParam >> 16) & 0xFFFF
		fmt.Fprintf(os.Stderr, "WM_COMMAND: controlID=%d, notificationCode=%d\n", controlID, notificationCode)

		if controlID == uintptr(idButtonHello) && notificationCode == bnClicked {
			fmt.Fprintf(os.Stderr, "Button clicked, showing message box\n")
			messageBox(hwnd, messageBoxContent, messageBoxTitle, uint(mbOk))
		}
		return 0
	case wmDestroy:
		fmt.Fprintf(os.Stderr, "WM_DESTROY received\n")
		postAQuitMessage(0)
		return 0
	}
	return defWindowProc(hwnd, msg, wParam, lParam)
}

func main() {
	// Create log file for debugging
	logFile, err := os.Create("win64_debug.log")
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		messageBox(0, fmt.Sprintf("Failed to create log file: %v", err), "Error", uint(mbOk))
		return
	}
	defer logFile.Close()

	// Redirect stderr to log file
	os.Stderr = logFile

	fmt.Fprintf(os.Stderr, "Application starting\n")

	// Try multiple approaches to get the module handle
	instance := getTheModuleHandle("")
	if instance == nullHandle {
		fmt.Fprintf(os.Stderr, "First attempt to get module handle failed, trying NULL directly\n")

		ret, _, err := getModuleHandleW.Call(0) // Pass 0 (NULL) directly
		instance = syscall.Handle(ret)

		if instance == nullHandle {
			errMsg := fmt.Sprintf("Could not get module handle even with direct NULL: %v", err)
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			messageBox(0, errMsg, "Error", uint(mbOk))

			// Try a hardcoded fallback approach as last resort
			fmt.Fprintf(os.Stderr, "Attempting fallback with hardcoded instance\n")
			instance = syscall.Handle(1) // Try a fallback non-zero handle

			if instance == nullHandle {
				return
			}
		}
	}
	fmt.Fprintf(os.Stderr, "Got module handle: %v\n", instance)

	classNamePtr := strToUint16Ptr(className)

	// Make sure cursor and icon are set
	wc := wndClassEx{
		Size:          uint32(unsafe.Sizeof(wndClassEx{})),
		Style:         0x0003, // CS_HREDRAW | CS_VREDRAW
		WndProc:       syscall.NewCallback(wndProcCallback),
		ClsExtra:      0,
		WndExtra:      0,
		hInstance:     instance,
		hIcon:         syscall.Handle(0),       // Use default icon
		hCursor:       getLoadCursor(0, 32512), // IDC_ARROW
		hbrBackground: syscall.Handle(5),       // COLOR_WINDOW+1
		MenuName:      nil,
		ClassName:     classNamePtr,
		hIconSm:       syscall.Handle(0),
	}

	atom, err := registerClassEx(&wc)
	if err != nil || atom == 0 {
		errMsg := fmt.Sprintf("RegisterClassEx failed: %v, atom: %d", err, atom)
		fmt.Fprintf(os.Stderr, "%s\n", errMsg)
		messageBox(0, errMsg, "Error", uint(mbOk))
		return
	}
	fmt.Fprintf(os.Stderr, "Registered window class, atom: %d\n", atom)

	hwnd, err := createWindow(
		0,
		className,
		windowTitle,
		uint32(wsOverlappedWindow|wsVisible),
		int32(cwUseDefault),
		int32(cwUseDefault),
		500,
		400,
		nullHandle,
		nullHandle,
		instance,
		nullPtr)

	if err != nil || hwnd == nullHandle {
		errMsg := fmt.Sprintf("CreateWindow failed: %v, hwnd: %v", err, hwnd)
		fmt.Fprintf(os.Stderr, "%s\n", errMsg)
		messageBox(0, errMsg, "Error", uint(mbOk))
		return
	}

	fmt.Fprintf(os.Stderr, "Window created: %v\n", hwnd)

	if !showTheWindow(hwnd, 1) { // SW_SHOWNORMAL
		fmt.Fprintf(os.Stderr, "ShowWindow failed\n")
	} else {
		fmt.Fprintf(os.Stderr, "Window shown successfully\n")
	}

	if !updateTheWindow(hwnd) {
		fmt.Fprintf(os.Stderr, "UpdateWindow failed\n")
	} else {
		fmt.Fprintf(os.Stderr, "Window updated successfully\n")
	}

	// Create Button with control ID
	buttonHwnd, err := createWindow(
		0,
		"BUTTON",
		buttonText,
		uint32(wsVisible|bsPushButton),
		50,
		50,
		100,
		30,
		hwnd,
		syscall.Handle(idButtonHello), // Set control ID here
		instance,
		nil)

	if err != nil || buttonHwnd == nullHandle {
		fmt.Fprintf(os.Stderr, "CreateWindow for button failed: %v, buttonHwnd: %v\n", err, buttonHwnd)
	} else {
		fmt.Fprintf(os.Stderr, "Button created successfully\n")

		if !showTheWindow(buttonHwnd, 1) {
			fmt.Fprintf(os.Stderr, "ShowWindow for button failed\n")
		}

		if !updateTheWindow(buttonHwnd) {
			fmt.Fprintf(os.Stderr, "UpdateWindow for button failed\n")
		}
	}

	// Process messages
	fmt.Fprintf(os.Stderr, "Entering message loop\n")
	var msg msg
	for {
		ret := getTheMessage(&msg, nullHandle, 0, 0)
		if ret == 0 {
			fmt.Fprintf(os.Stderr, "WM_QUIT received, exiting message loop\n")
			break
		}
		if ret == -1 {
			fmt.Fprintf(os.Stderr, "Error in getMessage, last error: %v\n", syscall.GetLastError())
			continue // Don't exit on error, try to keep processing messages
		}
		translateTheMessage(&msg)
		dispatchTheMessage(&msg)
	}
	fmt.Fprintf(os.Stderr, "Application exiting\n")
}

// Add new functions needed
func getLoadCursor(hInstance syscall.Handle, cursorID uint32) syscall.Handle {
	ret, _, _ := loadCursorW.Call(
		uintptr(hInstance),
		uintptr(cursorID))
	return syscall.Handle(ret)
}
