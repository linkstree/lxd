package apparmor

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lxc/lxd/lxd/sys"
	"github.com/lxc/lxd/shared"
	"github.com/lxc/lxd/shared/subprocess"
)

var qemuImgProfileTpl = template.Must(template.New("qemuImgProfile").Parse(`#include <tunables/global>
profile "{{ .name }}" flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  capability dac_override,
  capability dac_read_search,
  capability ipc_lock,

  /sys/devices/**/block/*/queue/max_segments  r,

{{range $index, $element := .allowedCmdPaths}}
  {{$element}} mixr,
{{- end }}

  {{ .pathToImg }} rk,

{{- if .dstPath }}
  {{ .dstPath }} rwk,
{{- end }}

{{- if .snap }}
  # Snap-specific libraries
  /snap/lxd/*/lib/**.so* mr,
{{- end }}

{{if .libraryPath -}}
  # Entries from LD_LIBRARY_PATH
{{range $index, $element := .libraryPath}}
  {{$element}}/** mr,
{{- end }}
{{- end }}
}
`))

type nullWriteCloser struct {
	*bytes.Buffer
}

func (nwc *nullWriteCloser) Close() error {
	return nil
}

// QemuImg runs qemu-img with an AppArmor profile based on the imgPath and dstPath supplied.
// The first element of the cmd slice is expected to be a priority limiting command (such as nice or prlimit) and
// will be added as an allowed command to the AppArmor profile. The remaining elements of the cmd slice are
// expected to be the qemu-img command and its arguments.
func QemuImg(sysOS *sys.OS, cmd []string, imgPath string, dstPath string) (string, error) {
	//It is assumed that command starts with a program which sets resource limits, like prlimit or nice
	allowedCmds := []string{"qemu-img", cmd[0]}

	allowedCmdPaths := []string{}
	for _, c := range allowedCmds {
		cmdPath, err := exec.LookPath(c)
		if err != nil {
			return "", fmt.Errorf("Failed to find executable %q: %w", c, err)
		}

		allowedCmdPaths = append(allowedCmdPaths, cmdPath)
	}

	// Attempt to deref all paths.
	imgFullPath, err := filepath.EvalSymlinks(imgPath)
	if err == nil {
		imgPath = imgFullPath
	}

	if dstPath != "" {
		dstFullPath, err := filepath.EvalSymlinks(dstPath)
		if err == nil {
			dstPath = dstFullPath
		}
	}

	err = qemuImgProfileLoad(sysOS, imgPath, dstPath, allowedCmdPaths)
	if err != nil {
		return "", fmt.Errorf("Failed to load qemu-img profile: %w", err)
	}

	defer func() {
		_ = qemuImgUnload(sysOS, imgPath)
		_ = qemuImgDelete(sysOS, imgPath)
	}()

	var buffer bytes.Buffer
	var output bytes.Buffer
	p := subprocess.NewProcessWithFds(cmd[0], cmd[1:], nil, &nullWriteCloser{&output}, &nullWriteCloser{&buffer})
	if err != nil {
		return "", fmt.Errorf("Failed creating qemu-img subprocess: %w", err)
	}

	p.SetApparmor(qemuImgProfileName(imgPath))

	err = p.Start(context.Background())
	if err != nil {
		return "", fmt.Errorf("Failed running qemu-img: %w", err)
	}

	_, err = p.Wait(context.Background())
	if err != nil {
		return "", shared.NewRunError(cmd[0], cmd[1:], err, nil, &buffer)
	}

	return output.String(), nil
}

// qemuImgProfileLoad ensures that the qemu-img's policy is loaded into the kernel.
func qemuImgProfileLoad(sysOS *sys.OS, imgPath string, dstPath string, allowedCmdPaths []string) error {
	profile := filepath.Join(aaPath, "profiles", qemuImgProfileFilename(imgPath))
	content, err := ioutil.ReadFile(profile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	updated, err := qemuImgProfile(imgPath, dstPath, allowedCmdPaths)
	if err != nil {
		return err
	}

	if string(content) != string(updated) {
		err = ioutil.WriteFile(profile, []byte(updated), 0600)
		if err != nil {
			return err
		}
	}

	err = loadProfile(sysOS, qemuImgProfileFilename(imgPath))
	if err != nil {
		return err
	}

	return nil
}

// qemuImgUnload ensures that the qemu-img's policy namespace is unloaded to free kernel memory.
// This does not delete the policy from disk or cache.
func qemuImgUnload(sysOS *sys.OS, imgPath string) error {
	err := unloadProfile(sysOS, qemuImgProfileName(imgPath), qemuImgProfileFilename(imgPath))
	if err != nil {
		return err
	}

	return nil
}

// qemuImgDelete removes the profile from cache/disk.
func qemuImgDelete(sysOS *sys.OS, imgPath string) error {
	return deleteProfile(sysOS, qemuImgProfileName(imgPath), qemuImgProfileFilename(imgPath))
}

// qemuImgProfile generates the AppArmor profile template from the given destination path.
func qemuImgProfile(imgPath string, dstPath string, allowedCmdPaths []string) (string, error) {
	// Render the profile.
	var sb *strings.Builder = &strings.Builder{}
	err := qemuImgProfileTpl.Execute(sb, map[string]any{
		"name":            qemuImgProfileName(imgPath),
		"pathToImg":       imgPath,
		"dstPath":         dstPath,
		"allowedCmdPaths": allowedCmdPaths,
		"snap":            shared.InSnap(),
		"libraryPath":     strings.Split(os.Getenv("LD_LIBRARY_PATH"), ":"),
	})
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// qemuImgProfileName returns the AppArmor profile name.
func qemuImgProfileName(outputPath string) string {
	return getProfileName(outputPath)
}

// qemuImgProfileFilename returns the name of the on-disk profile name.
func qemuImgProfileFilename(outputPath string) string {
	return getProfileName(outputPath)
}

func getProfileName(outputPath string) string {
	name := strings.Replace(strings.Trim(outputPath, "/"), "/", "-", -1)
	return profileName("qemu-img", name)
}
