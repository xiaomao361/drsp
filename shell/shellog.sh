#!/bin/bash

# date: 2018-03-15
# author: zhouwei
# email: xiaomao361@163.com
# func: log level echo

info() {
	echo " Info   $(date +'%Y-%m-%d %H:%M:%S'): $1"
}

success() {
	echo -e "\033[32m Success \033[0m $(date +'%Y-%m-%d %H:%M:%S'): $1"
}

error() {
	echo -e "\033[31m Error \033[0m $(date +'%Y-%m-%d %H:%M:%S'): $1"
}

warn() {
	echo -e "\033[33m Warn \033[0m $(date +'%Y-%m-%d %H:%M:%S'): $1"
}
