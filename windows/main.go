package main

import (
	"fmt"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func main() {
	fmt.Println("Hello world")
	// Initialize COM library
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		panic(fmt.Errorf("CoInitializeEx failed: %w", err))
		return
	}
	defer ole.CoUninitialize()

	// Create WScript.Shell COM object
	wsShellOle, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		panic(fmt.Errorf("CreateObject failed: %w", err))
	}
	shell, err := wsShellOle.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		panic(fmt.Errorf("QueryInterface failed: %w", err))
	}
	defer shell.Release()

	// Call the Popup method to display "Hello World"
	_, err = oleutil.CallMethod(shell, "Popup", "Hello World from COM (Go)!")
	if err != nil {
		panic(fmt.Errorf("CallMethod failed: %w", err))
	}
}

//https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nn-shobjidl_core-icustomdestinationlist
