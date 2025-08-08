#!/bin/sh
set -u

print_usage() {
   echo "Usage: $0 --file <filepath> --exec <bin> [--cmd <command>] [--max-attempts <N>] [--sleep-seconds <S>]"
   exit 1
}

file=""
exec_cmd=""
exec_bin=""
max_attempts=60
sleep_seconds=1

while [ $# -gt 0 ]; do
   case "$1" in
      --file)
         file="$2"
         shift 2
         ;;
      --cmd)
         exec_cmd="$2"
         shift 2
         ;;
      --exec)
         exec_bin="$2"
         shift 2
         ;;
      --max-attempts)
         max_attempts="$2"
         shift 2
         ;;
      --sleep-seconds)
         sleep_seconds="$2"
         shift 2
         ;;
      *)
         echo "Unknown argument: $1"
         print_usage
         ;;
   esac
done

if [ -z "$file" ] || [ -z "$exec_bin" ]; then
   print_usage
fi

echo "[waitfile]  ⏳  Waiting for $file..."

attempt=0
while [ "$attempt" -lt "$max_attempts" ]; do
   if [ -s "$file" ]; then
      echo "[waitfile]  ✅    File found, execute command"
      if [ -n "$exec_cmd" ]; then
         eval "$exec_cmd"
      fi
      exec $exec_bin
   fi
   attempt=$((attempt + 1))
   sleep "$sleep_seconds"
done

echo "[waitfile]  ❌    File '$file' not found or empty after $max_attempts attempts."
exit 1
