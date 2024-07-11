import subprocess

def run_docker_command():
    # 启动 Docker 容器并执行命令
    command = (
        'docker run --network 2pc-network -i opencbdc-tx-twophase /bin/bash -c '
        '"./build/src/uhs/client/client-cli 2pc-compose.cfg mempool0.dat wallet0.dat mint 10 5"'
    )
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    
    # 检查命令执行是否成功
    if result.returncode != 0:
        print("Docker 命令执行失败。输出如下：")
        print(result.stderr)
        return
    
    # 保留输出内容到文件
    with open('output.txt', 'w') as file:
        file.write(result.stdout)
    
    # 输出内容到终端
    print(result.stdout)

if __name__ == "__main__":
    run_docker_command()

