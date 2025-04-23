#!/usr/bin/env python3
# Copyright 2025 Tim Andersson
# See LICENSE file for licensing details.

"""Charm the application."""

import logging
import subprocess
from pathlib import Path

import ops

logger = logging.getLogger(__name__)


SNAPS = {}

DEBS = [
]


class UbuntuGuiTestingHandlerCharm(ops.CharmBase):
    """Charm the application."""
    snaps = {
        "hello": "stable"
    }
    debs = [
        "hello",
    ]

    def __init__(self, framework: ops.Framework):
        super().__init__(framework)
        self.framework.observe(self.on.start, self._on_start)
        self.framework.observe(self.on.install, self._on_install)

        self.framework.observe(self.on.update_status, self._update_status)

        self.framework.observe(self.on.db_relation_joined, self._db_relation_changed)
        self.framework.observe(self.on.db_relation_changed, self._db_relation_changed)

    def _db_relation_changed(self, event: ops.charm.RelationChangedEvent) -> None:
        unit_data = event.relation.data[event.unit]  # a dict
        # required_relation_data = ["master", "allowed-units", "port", "user"]
        # write unit_data to a file
        datafile = ""
        for key, val in unit_data.items():
            datafile += f"{key}={val}\n"
        Path("/home/ubuntu/test-db-relation-changed").write_text(datafile)
        """
        ubuntu@juju-2aeeca-stg-test-openstack-vm-testing-over-vnc-7:~$ cat test-db-relation-changed 
        allowed-subnets=10.142.182.88/32
        allowed-units=ubuntu-gui-testing-handler/0
        database=ubuntu-gui-testing-handler
        egress-subnets=10.142.182.81/32
        host=10.142.182.81
        ingress-address=10.142.182.81
        master=dbname=ubuntu-gui-testing-handler host=10.142.182.81 password=8q4Gfh4rYtCZkXkl port=5432 user=relation-9
        password=8q4Gfh4rYtCZkXkl
        port=5432
        private-address=10.142.182.81
        schema_password=8q4Gfh4rYtCZkXkl
        schema_user=relation-9
        state=standalone
        user=relation-9
        version=14.15
        """

    def _update_status(self, event: ops.charm.UpdateStatusEvent):
        pass

    def _on_start(self, event: ops.StartEvent):
        """Handle start event."""
        self.unit.status = ops.ActiveStatus()

    def _on_install(self, event: ops.InstallEvent):
        """Handle install event."""
        # FIXME
        # This function needs to somehow wait for the db relation to be all set up
        self.unit.status = ops.MaintenanceStatus("Installing snaps...")
        self.machine_setup()

    def machine_setup(self):
        self.install_snaps()
        self.install_debs()

    def install_snaps(self):
        for snap, channel in self.snaps.items():
            self.unit.status = ops.MaintenanceStatus(f"Installing snap `{snap}` from channel `{channel}`...")
            subprocess.run(
                f"/usr/bin/snap install {snap} --{channel}".split(" "),
                check=True,
            )

    def install_debs(self):
        debs = " ".join(self.debs)
        self.unit.status = ops.MaintenanceStatus(f"Installing debs:\n`{debs}`")
        subprocess.run(
            f"/usr/bin/apt-get -y install {debs}".split(" "),
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
    ops.main(UbuntuGuiTestingHandlerCharm)
