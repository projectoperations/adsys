#! /usr/bin/python3

import argparse
import dbus
from sys import exit
from subprocess import Popen
from gi.repository import GLib

from session_daemons import start_session_bus, run_session_mocks
from system_daemons import start_system_bus, run_system_mocks

POLKIT_PATH = "/usr/libexec/polkitd"
ADSYS_POLICY_PATH_SRC = "/usr/share/polkit-1/actions.orig/com.ubuntu.adsys.policy"
ADSYS_POLICY_PATH_DST = "/usr/share/polkit-1/actions/com.ubuntu.adsys.policy"

# For testing purposes
# ADSYS_POLICY_PATH_SRC = "/tmp/actions.orig/com.ubuntu.adsys.policy"
# ADSYS_POLICY_PATH_DST = "/tmp/actions/com.ubuntu.adsys.policy"


def main() -> int:
    """main routine"""

    parser = argparse.ArgumentParser(description="dbus mocks")
    parser.add_argument(
        "mode", type=str,
        choices=[
            "polkit_yes", "polkit_no",
            "no_startup_time", "invalid_startup_time",
            "no_nextrefresh_time", "invalid_nextrefresh_time",
            "subscription_disabled"])

    parser.add_argument(
        '-t', '--type', type=str, default='all',
        choices=['session', 'system', 'all']
    )

    args = parser.parse_args()

    dbus.mainloop.glib.DBusGMainLoop(set_as_default=True)
    main_loop = GLib.MainLoop()

    bus_type = args.type

    if bus_type == 'system' or bus_type == 'all':
        system_bus = start_system_bus()
        run_system_mocks(system_bus, args.mode)
        system_bus.add_signal_receiver(main_loop.quit, signal_name="Disconnected",
                                       path="/org/freedesktop/DBus/Local",
                                       dbus_interface="org.freedesktop.DBus.Local")

    if bus_type == 'session' or bus_type == 'all':
        session_bus = start_session_bus()
        run_session_mocks(session_bus)
        session_bus.add_signal_receiver(main_loop.quit, signal_name="Disconnected",
                                        path="/org/freedesktop/DBus/Local",
                                        dbus_interface="org.freedesktop.DBus.Local")

    # quit mock when the bus is going down
    polkitd = allow_adsys_and_start_polkitd_in_bg(args.mode)

    main_loop.run()

    polkitd.terminate()

    return 0


def allow_adsys_and_start_polkitd_in_bg(mode: str) -> Popen:
    """Replace adsys policy depending on mode and starts polkitd in background"""

    allow = "yes"
    if mode == "polkit_no":
        allow = "no"

    with open(ADSYS_POLICY_PATH_SRC, "r") as r:
        with open(ADSYS_POLICY_PATH_DST, "w") as w:
            for line in r:
                for token in ["<allow_any>", "<allow_inactive>", "<allow_active>"]:
                    if token not in line:
                        continue
                    line = "      " + token + allow + "</" + token[1:] + "\n"
                w.write(line)
    return Popen([POLKIT_PATH])


if __name__ == '__main__':
    exit(main())
