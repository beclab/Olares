#!/usr/bin/env bash

source ./common.sh

config_system() {
    echo "config system"
    return
    local ntpdate hwclock

    # kernel printk log level
    ensure_success $sh_c 'sysctl -w kernel.printk="3 3 1 7"'

    # ntp sync
    ntpdate=$(command -v ntpdate)
    hwclock=$(command -v hwclock)

    printf '#!/bin/sh\n\n%s -b -u pool.ntp.org && %s -w\n\nexit 0\n' "$ntpdate" "$hwclock" > cron.ntpdate
    ensure_success $sh_c '/bin/sh cron.ntpdate'
    ensure_success $sh_c 'cat cron.ntpdate > /etc/cron.daily/ntpdate && chmod 0700 /etc/cron.daily/ntpdate'
    ensure_success rm -f cron.ntpdate
}

config_system