LOG_DATE_FORMAT='+%Y/%m/%d %H:%M:%S'

loginfo() {
    echo -ne '\033[1;32m'                           1>&2
    echo -n "[`date "$LOG_DATE_FORMAT"`] INFO "     1>&2
    echo $*                                         1>&2
    echo -ne '\033[0m'                              1>&2
}

__escaped_date_sed() {
    date "$LOG_DATE_FORMAT" | sed -E 's/\//\\\//g'
}

loginfo_stream() {
    echo -ne '\033[1;32m'                           1>&2
    sed -E 's/^(.*)$/['"`__escaped_date_sed`"'] INFO \1/g' 1>&2
    echo -ne '\033[0m'                              1>&2
}

logerror() {
    echo -ne '\033[1;31m'                           1>&2
    echo -n "[`date "$LOG_DATE_FORMAT"`] ERROR "    1>&2
    echo $*                                         1>&2
    echo -ne '\033[0m'                              1>&2
}

logwarn() {
    echo -ne '\033[1;33m'                           1>&2
    echo -n "[`date "$LOG_DATE_FORMAT"`] WARN "     1>&2
    echo $*                                         1>&2
    echo -ne '\033[0m'                              1>&2
}

log_exec() {
    loginfo exec $*
    $*
    return $?
}
