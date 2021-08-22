package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/miquella/ask"
	"github.com/olekukonko/tablewriter"
	"github.com/pkorotkov/safebox/internal/metadata"
	"github.com/pkorotkov/safebox/internal/utils"
	"github.com/pkorotkov/safebox/internal/veracrypt"
	"github.com/urfave/cli/v2"
)

func before(ctx *cli.Context) error {
	vis, err := veracrypt.GetVolumeInfos()
	if err != nil {
		return cli.Exit("failed to get veracrypt volumes infos", 2)
	}
	md := metadata.New()
	md.SetVolumeInfos(vis)
	if os.Geteuid() == 0 {
		md.SetEnoughPriviledges()
	}
	realUser, err := user.Lookup(os.Getenv("SUDO_USER"))
	if err == nil {
		md.SetRealUser(realUser)
	}
	ctx.App.Metadata = md
	return nil
}

func statusCommand() *cli.Command {
	c := &cli.Command{
		Name:    "status",
		Aliases: []string{"s"},
		Usage:   "Shows status of mounted containers",
		Action: func(ctx *cli.Context) error {
			md := metadata.Metadata(ctx.App.Metadata)
			vis := md.VolumeInfos()
			if nothingMounted := (vis == nil); nothingMounted {
				fmt.Fprintln(os.Stdout, "no container mounted")
				return nil
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Container Path", "Mount Point", "Virtual Device", "Size", "Read-Only", "Mounted Folders"})
			for _, vi := range vis {
				line := []string{
					vi.SlotNumber,
					// TODO: Apply a limit for max string length (30 ASCII chars).
					vi.ContainerPath,
					// TODO: Apply a limit for max string length (30 ASCII chars).
					vi.MountPoint,
					vi.VirtualDevice,
					vi.Size,
					vi.ReadOnly,
					strings.Join(vi.MountedFolders, "\n"),
				}
				table.Append(line)
			}
			table.Render()
			return nil
		},
	}
	return c
}

func mountCommand() *cli.Command {
	var (
		withSSH   bool
		withGNUPG bool
	)
	c := &cli.Command{
		Name:    "mount",
		Aliases: []string{"m"},
		Usage:   "Mount a crypto container from the given file",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "with-ssh",
				Value:       false,
				Usage:       "Indicate whether to mount .ssh directory found in the container",
				Destination: &withSSH,
			},
			&cli.BoolFlag{
				Name:        "with-gnupg",
				Value:       false,
				Usage:       "Indicate whether to mount .gnupg directory found in the container",
				Destination: &withGNUPG,
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() < 2 {
				return cli.Exit("error: wrong number of command arguments", 3)
			}
			md := metadata.Metadata(ctx.App.Metadata)
			realUser, ok := md.RealUser()
			if !ok || !md.EnoughPriviledges() {
				return cli.Exit("error: app has not enough privileges", 5)
			}
			containerPath, mountPoint := ctx.Args().Get(0), ctx.Args().Get(1)
			if !utils.FileExists(containerPath) {
				return cli.Exit("error: failed to find container file", 4)
			}
			pass, err := ask.HiddenAsk("Enter container password: ")
			if err != nil {
				return cli.Exit(fmt.Errorf("error: failed to read master password: %w", err), 6)
			}
			mountTarget := filepath.Join(mountPoint, filepath.Base(containerPath))
			uid64, _ := strconv.ParseUint(realUser.Uid, 10, 32)
			gid64, _ := strconv.ParseUint(realUser.Gid, 10, 32)
			uid, gid := uint32(uid64), uint32(gid64)
			if err := utils.MakeDirectory(mountTarget, uid, gid); err != nil {
				return cli.Exit(fmt.Sprintf("error: failed to create a folder for the container: %s", err), 7)
			}
			_, _, err = utils.LaunchProgram(
				"veracrypt",
				[]string{"--stdin", "--non-interactive", "--mount", containerPath, mountTarget},
				[]byte(pass),
				nil,
			)
			if err != nil {
				return cli.Exit(fmt.Errorf("error: veracrypt failed to perform container mounting: %w", err), 8)
			}
			var (
				folder   string
				exitCode int
			)
			if withSSH {
				folder = ".ssh"
				err, exitCode = mountFolder(realUser.HomeDir, uid, gid, mountTarget, folder), 9
			}
			if withGNUPG {
				folder = ".gnupg"
				err, exitCode = mountFolder(realUser.HomeDir, uid, gid, mountTarget, folder), 10
			}
			if err != nil {
				return cli.Exit(fmt.Errorf("error: failed to mount %s folder: %w", folder, err), exitCode)
			}
			return nil
		},
	}
	return c
}

func mountFolder(realUserHomeFolder string, uid, gid uint32, mountTarget, folder string) error {
	homeFolder := filepath.Join(realUserHomeFolder, folder)
	err := utils.MakeDirectory(homeFolder, uid, gid)
	if err != nil {
		return fmt.Errorf("failed to create %s folder: %w", folder, err)
	}
	containerFolder := filepath.Join(mountTarget, folder)
	// Check that the container has `folder`.
	if utils.FileExists(containerFolder) {
		_, _, err = utils.LaunchProgram(
			"mount",
			[]string{"--bind", containerFolder, homeFolder},
			nil,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to mount %s folder: %w", folder, err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "warning: failed to find %s folder in the container\n", folder)
	}
	return nil
}

func unmountCommand() *cli.Command {
	c := &cli.Command{
		Name:    "unmount",
		Aliases: []string{"u"},
		Usage:   "Unmount a specified crypto container",
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() < 1 {
				return cli.Exit("error: wrong number of command arguments", 11)
			}
			md := metadata.Metadata(ctx.App.Metadata)
			if !md.EnoughPriviledges() {
				return cli.Exit("error: app has not enough privileges", 12)
			}
			containerPath := ctx.Args().Get(0)
			vis := md.VolumeInfos()
			var (
				err error
				svi *veracrypt.VolumeInfo
			)
			for _, vi := range vis {
				if (containerPath == vi.SlotNumber) || (containerPath == vi.ContainerPath) {
					svi = vi
					break
				}
			}
			if svi == nil {
				return cli.Exit("error: failed to recognize container to unmount", 13)
			}
			for _, mf := range svi.MountedFolders {
				if _, _, err = utils.LaunchProgram("umount", []string{mf}, nil, nil); err != nil {
					return cli.Exit(fmt.Errorf("error: failed to unmount .ssh folder: %w", err), 14)
				}
			}
			_, _, err = utils.LaunchProgram(
				"veracrypt",
				[]string{"--dismount", svi.ContainerPath},
				nil,
				nil,
			)
			if err != nil {
				return cli.Exit(fmt.Errorf("error: failed to unmount the crypto container: %w", err), 15)
			}
			return nil
		},
	}
	return c
}
