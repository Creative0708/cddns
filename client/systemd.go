package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrg/xdg"
)

var SERVICE_TEMPLATE = strings.TrimSpace(`
[Unit]
Description=CDDNS (Colin's Dynamic DNS)
Wants=network.target
After=network-online.target

[Service]
Type=simple
ExecStart=%s

[Install]
WantedBy=default.target
`)
var TIMER_TEMPLATE = strings.TrimSpace(`
[Unit]
Description=CDDNS (Colin's Dynamic DNS)
Wants=network.target
After=network-online.target

[Timer]
%s

[Install]
WantedBy=timers.target
`)

type InstallType int

const (
	InstallTypeOnReboot InstallType = iota
	InstallTypeHourly
	InstallTypeDaily
	InstallTypeWeekly
	INSTALL_TYPE_LENGTH int = iota
)

var INSTALL_TYPE_TIMER_OPTS = [...]string{
	"OnCalendar=hourly\nPersistent=true", // hourly
	"OnCalendar=daily\nPersistent=true",  // daily
	"OnCalendar=weekly\nPersistent=true", // weekly
}

func installSystemdService(programPath string, install InstallType) error {
	servicePath, err := xdg.DataFile("systemd/user/cddns.service")
	if err != nil {
		return err
	}
	contents := fmt.Appendf([]byte{}, SERVICE_TEMPLATE, programPath)
	if err := os.WriteFile(servicePath, contents, 0o644); err != nil {
		return err
	}

	procAttr := os.ProcAttr{
		Files: []*os.File{nil, os.Stdout, os.Stderr},
	}

	process, err := os.StartProcess("/usr/bin/systemctl", []string{"systemctl", "--user", "enable", "cddns.service"}, &procAttr)
	if err != nil {
		return err
	}
	process.Wait()

	var timerOpts string
	switch install {
	case InstallTypeHourly:
		timerOpts = "OnCalendar=hourly\nPersistent=true"
	case InstallTypeDaily:
		timerOpts = "OnCalendar=daily\nPersistent=true"
	case InstallTypeWeekly:
		timerOpts = "OnCalendar=weekly\nPersistent=true"
	default:
		return nil
	}
	timerPath, err := xdg.DataFile("systemd/user/cddns.timer")
	if err != nil {
		return err
	}
	contents = fmt.Appendf([]byte{}, TIMER_TEMPLATE, timerOpts)
	if err := os.WriteFile(timerPath, contents, 0o644); err != nil {
		return err
	}

	process, err = os.StartProcess("/usr/bin/systemctl", []string{"systemctl", "--user", "enable", "cddns.timer"}, &procAttr)
	if err != nil {
		return err
	}
	process.Wait()

	return nil
}
