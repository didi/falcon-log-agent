#!/bin/bash
workspace=$(cd $(dirname $0) && pwd -P)
cd $workspace

## const
module=log-agent
app=falcon-$module
godepsrc=$workspace/deps

cfg=dev.cfg
gitversion=.gitversion
control=./control.sh
thrift=rpc.thrift
pem=vault.pem
supervisor_conf=supervisor.conf.append
#deploymeta=./deploy-meta
#dockerfile=./Dockerfile

## function
function build() {
	# 设置golang环境变量
    echo -e "`go version`"
    export GOPATH=$workspace:$godepsrc
    echo $GOPATH


	local go16="/usr/local/go1.6.2"
	if [ -d "$go16" ]; then
	   export GOROOT="$go16"
	   export PATH=$GOROOT/bin:$PATH
	fi

    # 进行编译
    go build -o $app main.go
    local sc=$?
    if [ $sc -ne 0 ];then
    	## 编译失败, 退出码为 非0
        echo "$app build error"
        exit $sc
    else
        echo -n "$app build ok, vsn="
        gitversion
    fi
}

function pack() {
    #cd ./output
    tar zcvf falcon-log-agent.tar.gz control falcon-log-agent cfg/dev.cfg
}

## internals
function gitversion() {
    git log -1 --pretty=%h > $gitversion
    local gv=`cat $gitversion`
    echo "$gv"
}


##########################################
## main
## 其中, 
## 		1.进行编译
##		2.生成部署包output
##########################################

# 1.进行编译
build
# 2.打包
pack

# 编译成功
echo -e "build done"
exit 0
