#!/bin/bash
###
 # @Author: your name
 # @Date: 2022-03-10 09:45:43
 # @LastEditTime: 2022-03-10 13:32:58
 # @LastEditors: Please set LastEditors
 # @Description: 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 # @FilePath: /flv/build.sh
### 
WORKPATH=$(dirname $0)
EXAMPLEPATH=${WORKPATH}/example

#tinygo glue code
TINYGLUECODEPATH=$(tinygo env TINYGOROOT)/targets/wasm_exec.js

#glue code
GOGLUECODEPATH=$(go env GOROOT)/misc/wasm/wasm_exec.js

GLUECODEPATH=$TINYGLUECODEPATH

if [[ ! -f "$TINYGLUECODEPATH" ]]; then
GLUECODEPATH=$GOGLUECODEPATH

fi

if [[ ! -f "$EXAMPLEPATH/wasm_exrc.js" ]]; then
cp $GLUECODEPATH $EXAMPLEPATH
fi

if [[ ! -f "$TINYGLUECODEPATH" ]]; then
tinygo build -o ${EXAMPLEPATH}/flv.wasm ${WORKPATH}/main.go
else 
GOOS=js GOARCH=wasm go build -o ${EXAMPLEPATH}/flv.wasm ${WORKPATH}/main.go
fi


