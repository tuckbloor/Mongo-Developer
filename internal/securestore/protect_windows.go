//go:build windows

package securestore

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

var (
	crypt32            = syscall.NewLazyDLL("crypt32.dll")
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	cryptProtectData   = crypt32.NewProc("CryptProtectData")
	cryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
	localFree          = kernel32.NewProc("LocalFree")
)

func protect(data []byte, _ string) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}
	in := dataBlob{cbData: uint32(len(data)), pbData: &data[0]}
	var out dataBlob
	r, _, err := cryptProtectData.Call(
		uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, 1,
		uintptr(unsafe.Pointer(&out)),
	)
	runtime.KeepAlive(data)
	if r == 0 {
		return nil, fmt.Errorf("CryptProtectData failed: %w", err)
	}
	defer localFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	return append([]byte(nil), unsafe.Slice(out.pbData, int(out.cbData))...), nil
}

func unprotect(data []byte, _ string) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}
	in := dataBlob{cbData: uint32(len(data)), pbData: &data[0]}
	var out dataBlob
	r, _, err := cryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, 1,
		uintptr(unsafe.Pointer(&out)),
	)
	runtime.KeepAlive(data)
	if r == 0 {
		return nil, fmt.Errorf("CryptUnprotectData failed: %w", err)
	}
	defer localFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	return append([]byte(nil), unsafe.Slice(out.pbData, int(out.cbData))...), nil
}
