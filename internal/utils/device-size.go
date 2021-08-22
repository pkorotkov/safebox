package utils

/*
#define _FILE_OFFSET_BITS 64
#include <unistd.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

int
get_device_size(const char *path, unsigned long long *size) {
    int fd = open(path, O_RDONLY);
	if (fd < 0)
		return 1;
    off_t sz = lseek(fd, 0, SEEK_END);
	close(fd);
	if (sz < 0)
		return 2;
	*size = (unsigned long long)(sz);
	return 0;
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

func GetDeviceSize(path string) (uint64, error) {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	var size C.ulonglong
	if C.get_device_size(cp, &size) != 0 {
		return uint64(0), errors.New("failed to get volume byte size")
	}
	return uint64(size), nil
}
