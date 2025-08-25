package native

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Foundation
// #import <Foundation/Foundation.h>
// void GoNSLog(void* messagePtr) {
//     NSString* format = (__bridge NSString*)messagePtr;
//     NSLog(@"%@", format);
// }
import "C"
import (
	"fmt"

	"github.com/progrium/darwinkit/macos/foundation"
)

func NSLog(message string) {
	foundationMsg := foundation.String_StringWithString(message)
	C.GoNSLog(foundationMsg.Ptr())
}
func FNSLog(format string, args ...any) {
	NSLog(fmt.Sprintf(format, args...))
}
