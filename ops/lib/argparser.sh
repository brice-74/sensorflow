#!/bin/bash

# --------------------------------------------------------
# args_parser.sh — Minimal argument parser for shell scripts
#
# Usage:
#   source /args_parser.sh
#
#   define_option "key" "default_value" "required|optional"
#   define_flag "flag_name"
#   parse_args "$@" or parse_args_and_validate "$@" || exit 1
#
#   get_option "key"    → returns value
#   has_flag "flag"     → returns true/false
#
# Supported formats:
#   --key value
#   key=value
#   --flag              (boolean flag)
#
# Supports --help to display this message and exit.
# --------------------------------------------------------

# Internal storage
declare -A OPTIONS_DEFAULTS
declare -A OPTIONS_VALUES
declare -A OPTIONS_FLAGS
declare -A OPTIONS_REQUIRED

define_option() {
   local name="$1"
   local default="$2"
   local required="${3:-optional}"
   OPTIONS_DEFAULTS["$name"]="$default"
   OPTIONS_VALUES["$name"]="$default"
   OPTIONS_REQUIRED["$name"]="$required"
}

define_flag() {
   local name="$1"
   OPTIONS_FLAGS["$name"]=false
}

get_option() {
   echo "${OPTIONS_VALUES[$1]}"
}

has_flag() {
   [[ "${OPTIONS_FLAGS[$1]}" == "true" ]]
}

print_usage() {
   echo "Usage: $0 [options]"
   echo
   echo "Options:"
   for key in "${!OPTIONS_DEFAULTS[@]}"; do
      local req="${OPTIONS_REQUIRED[$key]}"
      local def="${OPTIONS_DEFAULTS[$key]}"
      printf "  --%-15s %s%s\n" "$key" \
         "$( [[ "$req" == "required" ]] && echo "(required) " )" \
         "$( [[ -n "$def" ]] && echo "default: $def" )"
   done
   echo
   echo "Flags:"
   for flag in "${!OPTIONS_FLAGS[@]}"; do
      printf "  --%s\n" "$flag"
   done
   echo
   echo "  --help            Display this help and exit"
}

error_exit() {
   echo "❌ Error: $1" >&2
   echo
   print_usage
   exit 1
}

parse_args() {
   while [[ $# -gt 0 ]]; do
      arg="$1"
      shift

      # Handle --help
      if [[ "$arg" == "--help" ]]; then
         print_usage
         exit 0
      fi

      if [[ "$arg" == --* ]]; then
         key="${arg:2}"

         if [[ ${OPTIONS_FLAGS[$key]+_} ]]; then
            OPTIONS_FLAGS["$key"]=true
         elif [[ ${OPTIONS_DEFAULTS[$key]+_} ]]; then
            if [[ $# -gt 0 ]]; then
               OPTIONS_VALUES["$key"]="$1"
               shift
            else
               error_exit "Missing value for option --$key"
            fi
         else
            error_exit "Unknown option: --$key"
         fi

      elif [[ "$arg" == *=* ]]; then
         key="${arg%%=*}"
         value="${arg#*=}"

         if [[ ${OPTIONS_DEFAULTS[$key]+_} ]]; then
            OPTIONS_VALUES["$key"]="$value"
         else
            error_exit "Unknown option: $key"
         fi

      else
         error_exit "Unknown argument: $arg"
      fi
   done
}

validate_required_options() {
   local missing=false
   for key in "${!OPTIONS_REQUIRED[@]}"; do
      if [[ "${OPTIONS_REQUIRED[$key]}" == "required" ]]; then
         if [[ -z "${OPTIONS_VALUES[$key]}" ]]; then
            echo "❌ Missing required option: --$key" >&2
            missing=true
         fi
      fi
   done

   if $missing; then
      echo
      print_usage
      return 1
   fi

   return 0
}

parse_args_and_validate() {
  parse_args "$@" || exit 1
  validate_required_options || exit 1
}
