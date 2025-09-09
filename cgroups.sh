#!/usr/bin/env bash
set -euo pipefail

# Path for the new cgroup
CGROUP_PATH="/sys/fs/cgroup/user.slice/user-501.slice/user@501.service/app.slice/myapp.scope"

# Detect our current control group (so we can restore later)
ORIG_CGROUP=$(grep '^0::' /proc/$$/cgroup | cut -d: -f3)

create_cgroup() {
  echo ">>> Creating cgroup at $CGROUP_PATH"

  sudo mkdir -p "$CGROUP_PATH"

  # Enable the pids controller in the parent
  echo "+pids" | sudo tee /sys/fs/cgroup/user.slice/user-501.slice/user@501.service/app.slice/cgroup.subtree_control

  # Limit processes to 5
  echo 5 | sudo tee "$CGROUP_PATH/pids.max"

  # Move current shell into the cgroup
  echo $$ | sudo tee "$CGROUP_PATH/cgroup.procs"
}

spawn_processes() {
  echo ">>> Spawning background processes"
  for i in {1..10}; do
    (sleep 5 &) && echo "Spawned Process $i with PID"
  done
}

show_status() {
  echo ">>> Status"
  echo -n "pids.current = "
  cat "$CGROUP_PATH/pids.current"

  echo "cgroup.procs:"
  cat "$CGROUP_PATH/cgroup.procs"
}

restore_and_cleanup() {
  echo ">>> Restoring shell back to original cgroup: $ORIG_CGROUP"
  echo $$ | sudo tee "/sys/fs/cgroup${ORIG_CGROUP}/cgroup.procs"

  echo ">>> Cleaning up $CGROUP_PATH"
  sudo rmdir "$CGROUP_PATH"
}

trap restore_and_cleanup EXIT

create_cgroup
spawn_processes
show_status

echo ">>> Sleeping for 5 seconds before cleanup..."
sleep 5
