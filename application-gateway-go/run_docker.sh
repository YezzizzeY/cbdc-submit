#!/bin/bash

# 启动 Docker 容器并执行命令
output=$(docker run --network 2pc-network -ti opencbdc-tx-twophase /bin/bash -c "
    ./build/src/uhs/client/client-cli 2pc-compose.cfg mempool0.dat wallet0.dat send 30 usd1qq227pnghgxl8svuusnsyhn8a0nysxfgez37emau8m79ql79mj7c769xync
")

# 检查命令执行是否成功
if [ $? -ne 0 ]; then
    echo "Docker 命令执行失败。"
    exit 1
fi

# 保留输出内容到文件
echo "$output" > output.txt

# 输出内容到终端
echo "$output"

