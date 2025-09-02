#!/bin/sh

DARK_PURPLE="$(printf '\033[0;38;5;57m')"
PURPLE="$(printf '\033[0;35m')"
DARK_GRAY="$(printf '\033[0;38;5;235m')"
NC="$(printf '\033[0m')"

echoc() {
   printf "%srun.dev.sh %s| %s%s%s\n" "$DARK_PURPLE" "$DARK_GRAY" "$PURPLE" "$1" "$NC"
}

runProcess() {
   "$GO_BIN_PATH" &
   process_pid=$!
   echoc "process is running, PID: $process_pid"
}

killRunningProcess() {
   process_pid=$(pidof $GO_BIN_PATH)
   if [ -n "$process_pid" ]; then
      echoc "killing old process, PID: $process_pid"
      kill -s TERM "$process_pid"
      wait "$process_pid" 2>/dev/null
      echoc "old process terminated, PID: $process_pid"
   else
      echoc "no running process found"
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
   eval go build $GO_BUILD_FLAGS -o "$GO_BIN_PATH" "$GO_SOURCE_FILE"
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

trap restart USR1
trap cleanup INT TERM

buildProcess
runProcess

echoc "awaiting signal USR1 - INT - TERM"

while :; do
   sleep 1
done
