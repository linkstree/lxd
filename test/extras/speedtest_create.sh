#!/bin/bash

MYDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
CNAME="speedtest"

count=${1}
if [ -z "${count}" ]; then
  echo "USAGE: ${0} 10"
  echo "This creates 10 busybox containers"
  exit 1
fi

if [ "${2}" != "notime" ]; then
  time ${0} "${count}" notime
  exit 0
fi

"${MYDIR}/deps/import-busybox" --alias busybox

PIDS=""
for c in $(seq "$count"); do
  lxc init busybox "${CNAME}${c}" 2>&1 &
  PIDS="$PIDS $!"
done

for pid in $PIDS; do
  wait "$pid"
done

echo -e "\nlxc list: All shutdown"
time lxc list 1>/dev/null

PIDS=""
for c in $(seq "$count"); do
  lxc start "${CNAME}${c}" 2>&1 &
  PIDS="$PIDS $!"
done

for pid in $PIDS; do
  wait "$pid"
done

echo -e "\nlxc list: All started"
time lxc list 1>/dev/null

echo -e "\nRun completed"
