#!/bin/bash
set -e

cd "$(dirname "$0")"

# Create VM disk directory and empty disk for qemu-bmc
mkdir -p vm-disks
if [ ! -f vm-disks/server01.qcow2 ]; then
  echo "Creating empty VM disk for server01..."
  qemu-img create -f qcow2 vm-disks/server01.qcow2 1G
  echo "Done."
else
  echo "VM disk already exists: vm-disks/server01.qcow2"
fi

echo ""
echo "Setup complete. Deploy the lab with:"
echo "  sudo clab deploy -t clabnoc-test.clab.yml"
echo ""
echo "Destroy with:"
echo "  sudo clab destroy -t clabnoc-test.clab.yml"
