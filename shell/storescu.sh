#!/bin/bash

# date: 2017-08-11
# author: zhouwei
# email: xiaomao361@163.com
# modify: 2018-03-15 modify to check a path dicom and send
#         2018-03-16 change the hole architecture

HOME=/home/pi/drsp
LOG=$HOME/drsp.log
CHECK=$HOME/dicom
TMP=$HOME/tmp
BIN=$HOME/dcm4che/bin
server_path=zhouwei@192.168.1.47:11112

source $HOME/shellog.sh

size_check() {
	info "now in func size_check"
	file_size=$(cd $CHECK && du | sed -n '$p' | awk '{print $1}')
	info "file size:"$file_size
	sleep 30
	compare_size=$(cd $CHECK && du | sed -n '$p' | awk '{print $1}')
	info "compare size:"$compare_size
	if [ "$file_size" == "$compare_size" ]; then
		# wait for 15s, if file size no changed, start storescp
		# use this func to check the images if has receive over
		info "file equle compare prepare storescu dicom images"
		mv $CHECK/* $TMP && scu || error "move images error"
		backup_check
		# take a look during files uploading, if there are others files sending or sended already
	else
		info "file unequle compare keep receiveing dicom images"
		size_check
		# need to take a look of the mem
		# while the files is sending, make sure it will not break from judge.
	fi
}

backup_check() {
	info "now in func backup_check"
	file_size=$(cd $CHECK && du | sed -n '$p' | awk '{print $1}')
	if [ "$file_size" -le 80 ]; then
		info "file size not change while sending..."
	else
		echo $(date +'%Y-%m-%d %H:%M:%S')
		warn "file changed while sending, back to the size check func"
		size_check
	fi
}

scu() {
	info "now in func scu"
	cd $HOME/tmp
	warn "moveing the dicom images from dicom path to  tmp path"
	$BIN/storescu -c $server_path $TMP
	if [ $? -eq 0 ]; then
		success "sending images success"
	else
		error "sending images error"
	fi
	warn "delete tmp images"
	rm -rf $HOME/tmp/*
}

while true; do
	echo "----start:$(date +'%Y-%m-%d %H:%M:%S')---"
	init_size=$(cd $CHECK && du | sed -n '$p' | awk '{print $1}')
	info "init_size: "$init_size
	if [ $init_size -ge 80 ]; then
		# empty directory size is 24 , so use 80 to judge if has receive dicom images
		info "directory size > 80, receiveing images, go to size check func"
		size_check
	else
		echo "no patient yet keep watching"
		sleep 5
	fi
	echo -e "----end:$(date +'%Y-%m-%d %H:%M:%S')-----\n"
done
