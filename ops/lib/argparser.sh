#!/bin/sh

# --------------------------------------------------------
# Minimal argument parser shell script (POSIX compliant)
#
# Usage:
#   source argparser.sh
#
# Define options with define_option "key" "default_value" "required|optional"
# Define flags with define_flag "flag_name"
# Parse args with parse_args_and_validate "$@" || exit 1
# Retrieve option value: get_option "key"
# Check flag presence: has_flag "flag"
# Supports:
#   --key value
#   key=value
#   --flag
#   --help
# --------------------------------------------------------

set -eo pipefail

OPTIONS_DEFAULTS=
OPTIONS_VALUES=
OPTIONS_REQUIRED=
FLAGS=

define_option() {
   key=$1
   val=$2
   req=${3:-optional}
   key_var=$(echo "$key" | tr '-' '_')
   eval OPTIONS_DEFAULTS_$key_var=\$val
   eval OPTIONS_VALUES_$key_var=\$val
   eval OPTIONS_REQUIRED_$key_var=\$req
   eval OPTIONS_KEYS_$key_var=\$key
}

define_flag() {
  flag=$1
  flag_var=$(echo "$flag" | tr '-' '_')
  eval FLAGS_$flag_var=0
}

get_option() {
   key=$1
   key_var=$(echo "$key" | tr '-' '_')
   eval "echo \"\$OPTIONS_VALUES_$key_var\""
}

has_flag() {
   flag=$1
   flag_var=$(echo "$flag" | tr '-' '_')
   eval "[ \"\$FLAGS_$flag_var\" = 1 ]"
}

print_usage() {
   echo "Usage: $0 [options]"
   echo
   echo "Options:"
   set | grep '^OPTIONS_DEFAULTS_' | while IFS= read -r line; do
      varname=${line%%=*}
      key=${varname#OPTIONS_DEFAULTS_}
      eval "def=\${$varname}"
      eval "req=\${OPTIONS_REQUIRED_$key}"
      printf "  --%-15s %s%s\n" "$key" \
         "$( [ "$req" = "required" ] && echo "(required) " )" \
         "$( [ -n "$def" ] && echo "default: $def" )"
   done
   echo
   echo "Flags:"
   set | grep '^FLAGS_' | while IFS= read -r line; do
      varname=${line%%=*}
      flag=${varname#FLAGS_}
      printf "  --%s\n" "$flag"
   done
   echo "  --help            Display this help and exit"
}

error_msg() {
   key_var=$1
   msg=$2
   eval key_orig=\${OPTIONS_KEYS_$key_var:-$key_var}
   echo "[argparser] ❌   Error: $msg: --$key_orig" >&2
   echo
   print_usage
}

parse_args() {
   while [ $# -gt 0 ]; do
      arg=$1
      shift

      if [ "$arg" = "--help" ]; then
         print_usage
         return 2
      fi

      case "$arg" in
         --*=*)
            key="${arg%%=*}"
            key="${key#--}"
            value="${arg#*=}"
            key_var=$(echo "$key" | tr '-' '_')
            if eval "[ \"\${OPTIONS_DEFAULTS_$key_var+x}\" ]"; then
               eval OPTIONS_VALUES_$key_var=\$value
            else
               error_msg "$key" "Unknown option"
               return 1
            fi
            ;;
         --*)
            key="${arg#--}"
            key_var=$(echo "$key" | tr '-' '_')
            if eval "[ \"\${FLAGS_$key_var+x}\" ]"; then
               eval FLAGS_$key_var=1
            elif eval "[ \"\${OPTIONS_DEFAULTS_$key_var+x}\" ]"; then
               if [ $# -eq 0 ]; then
                  error_msg "$key" "Missing value for option"
                  return 1
               fi
               val=$1
               shift
               eval OPTIONS_VALUES_$key_var=\$val
            else
               error_msg "$key" "Unknown option"
               return 1
            fi
            ;;
         *=*)
            key="${arg%%=*}"
            value="${arg#*=}"
            key_var=$(echo "$key" | tr '-' '_')
            if eval "[ \"\${OPTIONS_DEFAULTS_$key_var+x}\" ]"; then
               eval OPTIONS_VALUES_$key_var=\$value
            else
               error_msg "$key" "Unknown option"
               return 1
            fi
            ;;
         *)
            error_msg "$key" "Unknown argument"
            return 1
            ;;
      esac
   done
   return 0
}

validate_required_options() {
   missing=0
   errors=""
   for varname in $(set | grep '^OPTIONS_REQUIRED_' | cut -d= -f1); do
      req=$(eval "echo \${$varname}")
      if [ "$req" = "required" ]; then
         key=${varname#OPTIONS_REQUIRED_}
         val=$(eval "echo \${OPTIONS_VALUES_$key}")
         if [ -z "$val" ]; then
            eval key_orig=\${OPTIONS_KEYS_$key:-$key}
            errors="${errors}[argparser] ❌   Error: Missing required option: --${key_orig}\n"
            missing=1
         fi
      fi
   done
   if [ "$missing" -eq 1 ]; then
      printf "$errors" >&2
      echo >&2
      print_usage >&2
      return 1
   fi
   return 0
}

parse_args_and_validate() {
   parse_args "$@" || return $?
   validate_required_options || return 1
   return 0
}