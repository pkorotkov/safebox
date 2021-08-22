package veracrypt

/*
#include <stdlib.h>
#include <string.h>
#include <mntent.h>

struct mount_point {
	char *virtual_device;
	char *path;
};

struct mount_points {
	struct mount_point **points;
	size_t               count;
};

struct mount_points *
create_mount_points(void) {
	size_t count = 100000;
	struct mount_points *points = malloc(sizeof(struct mount_points));
	points->points = malloc(count * sizeof(struct mount_point *));
	FILE *mps = setmntent("/proc/mounts", "r");
	if (mps == NULL)
		return NULL;
	struct mntent *ent;
	size_t index = 0;
  	while (NULL != (ent = getmntent(mps))) {
		if (index == count)
			break;
		struct mount_point *mp = malloc(sizeof(struct mount_point));
		points->points[index] = mp;
		mp->virtual_device = malloc(strlen(ent->mnt_fsname) + 1);
		strcpy(mp->virtual_device, ent->mnt_fsname);
		mp->path = malloc(strlen(ent->mnt_dir) + 1);
		strcpy(mp->path, ent->mnt_dir);
		++index;
  	}
	endmntent(mps);
	if (index < count)
		count = index;
	points->count = count;
    void *np = realloc(points->points, count);
    if (np != NULL)
        points->points = np;
	return points;
}

void
free_mount_points(struct mount_points **points) {
    struct mount_points *ps = *points;
    struct mount_point *p;
    for (size_t i = 0; i < ps->count; ++i) {
        p = ps->points[i];
        free(p->virtual_device);
        free(p->path);
        free(p);
    }
    free(ps);
    *points = NULL;
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"unsafe"

	"github.com/pkorotkov/safebox/internal/utils"
)

var veracryptListRegExp = regexp.MustCompile(`Slot: (\d+)\s?\nVolume: (\S+)\s?\nVirtual Device: (\S+)\s?\nMount Directory: (\S+)\s?\n.*\n.*\nRead-Only: (\w+)`)

type VolumeInfo struct {
	SlotNumber     string
	ContainerPath  string
	MountPoint     string
	VirtualDevice  string
	Size           string
	ReadOnly       string
	MountedFolders []string
}

type MountPoint struct {
	VirtualDevice string
	Path          string
}

func GetVolumeInfos() ([]*VolumeInfo, error) {
	listso, listse, err := utils.LaunchProgram("veracrypt", []string{"--list", "--verbose"}, nil, nil)
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			errorString := []byte{69, 114, 114, 111, 114, 58, 32, 78, 111, 32, 118, 111, 108, 117, 109, 101, 115, 32, 109, 111, 117, 110, 116, 101, 100, 46, 10}
			if e.ExitCode() == 1 && bytes.Equal(listse, errorString) {
				return nil, nil
			}
		} else {
			return nil, fmt.Errorf("error: failed to get veracrypt volume list: %s", err)
		}
	}
	vis := parseVeracryptListOutput(listso)
	mps := readMointPoints()
	if mps != nil {
		for _, vi := range vis {
			for _, mp := range mps[vi.VirtualDevice] {
				if mp.Path != vi.MountPoint {
					vi.MountedFolders = append(vi.MountedFolders, mp.Path)
				}
			}
		}
	}
	return vis, nil
}

func parseVeracryptListOutput(output []byte) []*VolumeInfo {
	slots := veracryptListRegExp.FindAllSubmatchIndex(output, -1)
	var vis []*VolumeInfo
	for _, slot := range slots {
		vi := &VolumeInfo{}
		vi.SlotNumber = string(output[slot[2]:slot[3]])
		vi.ContainerPath = string(output[slot[4]:slot[5]])
		vi.VirtualDevice = string(output[slot[6]:slot[7]])
		vi.MountPoint = string(output[slot[8]:slot[9]])
		size, e := utils.GetDeviceSize(vi.VirtualDevice)
		if e != nil {
			vi.Size = "-"
		} else {
			vi.Size = strconv.FormatUint(size, 10)
		}
		vi.ReadOnly = string(output[slot[10]:slot[11]])
		vis = append(vis, vi)
	}
	return vis
}

func readMointPoints() map[string][]*MountPoint {
	var cmps *C.struct_mount_points = C.create_mount_points()
	if cmps == nil {
		return nil
	}
	defer C.free_mount_points(&cmps)
	ps := (*[1 << 28]*C.struct_mount_point)(unsafe.Pointer(cmps.points))[:cmps.count:cmps.count]
	mps := make(map[string][]*MountPoint)
	for i := C.size_t(0); i < cmps.count; i++ {
		vd := C.GoString(ps[i].virtual_device)
		mp := &MountPoint{vd, C.GoString(ps[i].path)}
		mps[vd] = append(mps[vd], mp)
	}
	return mps
}
