#! /bin/bash

SAR_DEV_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
SAR_NOT_CALL_INSTALL=1 source "$SAR_DEV_DIR/install.sh"

SAR_TMP_SCRIPT_1=/tmp/SARTMP$random$random$random
_runtime_env_script_gen "$SAR_DEV_DIR" > $SAR_TMP_SCRIPT_1
source $SAR_TMP_SCRIPT_1
rm -f $SAR_TMP_SCRIPT_1