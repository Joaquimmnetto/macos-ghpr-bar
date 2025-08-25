package view

import (
	"encoding/base64"
	"errors"
	"fmt"
	"macos-gh-bar/native"
	"strings"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/helper/action"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

func MenuItem(title string, charCode string, handler action.Handler) appkit.MenuItem {
	item := appkit.NewMenuItemWithAction(title, charCode, handler)
	objc.Retain(&item)
	return item
}

func MenuSeparator() appkit.MenuItem {
	sep := appkit.MenuItem_SeparatorItem()
	objc.Retain(&sep)
	return sep
}

func SubsectionTitleLabel(title string) appkit.TextField {
	label := appkit.NewLabel(title) // returns a TextField styled as a label
	label.SetBezeled(false)         // no border
	label.SetDrawsBackground(false) // transparent background
	label.SetEditable(false)        // not editable
	label.SetSelectable(false)      // not selectable
	label.SetAlignment(appkit.CenterTextAlignment)
	label.SetFont(appkit.FontClass.SystemFontOfSize(13))
	label.SetTextColor(appkit.Color_GrayColor())
	return label
}

func joinErrorMessages(err error, sep string) string {
	var msgs []string
	for err != nil {
		msgs = append(msgs, err.Error())
		err = errors.Unwrap(err)
	}
	return strings.Join(msgs, sep)
}

func DispatchErrorAlert(err error) {
	DispatchAlert("GithubBar Error", joinErrorMessages(err, "\n\t"))
}

func DispatchAlert(title, message string) {
	dispatch.MainQueue().DispatchAsync(func() {
		alert := appkit.NewAlert()
		alert.SetMessageText(title)
		alert.SetInformativeText(message)
		alert.AddButtonWithTitle("OK")
		alert.RunModal()
	})
}

func DispatchAlertOnError(err error) bool {
	if err != nil {
		native.NSLog(joinErrorMessages(err, " "))
		DispatchErrorAlert(err)
		return true
	}
	return false
}

func DispatchMarkBarButtonOnError(statusItem appkit.StatusItem, err error) {
	dispatch.MainQueue().DispatchAsync(func() {
		currentTitle := statusItem.Button().Title()
		if !strings.HasSuffix(currentTitle, "❗") {
			statusItem.Button().SetTitle(currentTitle + "❗")
		}
		menuItems := statusItem.Menu().ItemArray()
		lastItem := menuItems[len(menuItems)-1]
		if lastItem.Title() == "Show last error" {
			statusItem.Menu().RemoveItem(lastItem)
		}
		statusItem.Menu().AddItem(MenuItem("Show last error", "e", func(sender objc.Object) {
			DispatchErrorAlert(err)
		}))
	})
}

func AppkitImageFromBase64(base64String string) appkit.Image {
	// Decode base64 string to byte slice
	data, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		native.NSLog(fmt.Sprintf("Error decoding base64: %v", err))
		panic(err)
	}
	image := appkit.NewImageWithData(data)
	image.SetSize(foundation.Size{
		Width:  20,
		Height: 20,
	}) // Set the size to 20x20
	return image
}
