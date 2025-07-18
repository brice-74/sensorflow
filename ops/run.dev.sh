#!/bin/bash

DARK_PURPLE='\033[0;38;5;57m'
PURPLE='\033[0;35m'
DARK_GRAY='\033[0;38;5;235m'
NC='\033[0m'

echoc() {
   echo -e "${DARK_PURPLE}run.dev.sh ${DARK_GRAY}| ${PURPLE}$1${NC}"
}

runProcess() {
   $GO_BIN_PATH &
   process_pid=$!
   echoc "process is running, PID: $process_pid"
}

killRunningProcess() {
   process_pid=$(pidof $GO_BIN_PATH)
   if [ -n "$process_pid" ]; then
      echoc "killing old process, PID: $process_pid"
      kill -s SIGTERM $process_pid
      wait $process_pid
      echoc "old process terminated, PID: $process_pid"
   else
      echoc "no running process found with PID: $process_pid"
   fi
}

buildProcess() {
   if [ -z "$GO_SOURCE_FILE" ]; then
      echoc "error: No source file provided for building the process (SOURCE_FILE)"
      exit 1
   fi

   if [ -z "$GO_BIN_PATH" ]; then
      echoc "error: No bin path provided for building the process (BIN_PATH)"
      exit 1
   fi

   echoc "building process at $GO_BIN_PATH from source $GO_SOURCE_FILE"
   go build $GO_BUILD_FLAGS -o $GO_BIN_PATH "$GO_SOURCE_FILE"
}

restart() {
   echoc "manual restart"
   killRunningProcess
   buildProcess
   runProcess
}

cleanup() {
   echoc "received SIGINT or SIGTERM. Terminating..."
   killRunningProcess
   exit 0
}

trap restart SIGUSR1
trap cleanup SIGINT SIGTERM

buildProcess
runProcess

echoc "awaiting signal SIGUSR1 - SIGINT - SIGTERM"
while true ; do
  sleep 1
done