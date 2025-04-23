#!/usr/bin/env python3
# Copyright 2025 Tim Andersson
# See LICENSE file for licensing details.

"""Charm the application."""

import logging
import subprocess

import ops

logger = logging.getLogger(__name__)


SNAPS = {}

DEBS = [
]


class UbuntuGuiTestingSpawnerCharm(ops.CharmBase):
    """Charm the application."""
    snaps = {
        "yarf": "beta",
    }
    debs = [
        "qemu-system-x86",
        "libvirt-daemon-system",
    ]

    def __init__(self, framework: ops.Framework):
        super().__init__(framework)
        framework.observe(self.on.start, self._on_start)
        framework.observe(self.on.install, self._on_install)

    def _on_start(self, event: ops.StartEvent):
        """Handle start event."""
        self.unit.status = ops.ActiveStatus()

    def _on_install(self, event: ops.InstallEvent):
        """Handle install event."""
        self.unit.status = ops.MaintenanceStatus("Installing snaps...")
        self.machine_setup()

    def machine_setup(self):
        self.install_snaps()
        self.install_debs()

    def install_snaps(self):
        for snap, channel in self.snaps.items():
            self.unit.status = ops.MaintenanceStatus(f"Installing snap `{snap}` from channel `{channel}`...")
            subprocess.run(
                f"snap install {snap} --{channel}",
                check=True,
            )

    def install_debs(self):
        debs = " ".join(self.debs)
        self.unit.status = ops.MaintenanceStatus(f"Installing debs:\n`{debs}`")
        subprocess.run(
            f"apt-get -y install {debs}",
            check=True,
        )

    def kvm_libvirtd_permissions(self):
"""
sudo usermod -G kvm -a "ubuntu"
sudo usermod -G libvirt-qemu -a "ubuntu"
sudo touch /etc/udev/rules.d/99-kvm.rules
KERNEL=="kvm", GROUP="kvm", MODE="0660" > /etc/udev/rules.d/99-kvm.rules
sudo udevadm control --reload-rules
sudo systemctl restart libvirtd.service
sudo systemctl restart qemu-kvm.service
# then, this can work:
# qemu-system-x86_64 -drive format=raw,file=ubuntu25-04.img -enable-kvm -m 8192M -smp 2 -machine type=q35,accel=kvm -usbdevice tablet -vga virtio -vnc :0,share=ignore
"""
        pass


if __name__ == "__main__":  # pragma: nocover
    ops.main(UbuntuGuiTestingSpawnerCharm)
