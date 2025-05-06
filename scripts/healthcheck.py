#!/usr/bin/env python3
import http.client
import os
import sys
import socket
import json
from urllib.parse import urlparse


def health_check():
    """
    对推理网关执行健康检查
    返回：成功时为 0，失败时为非 0
    """
    # 从环境变量获取主机和端口，或使用默认值
    host = os.environ.get("HEALTH_CHECK_HOST", "localhost")
    port = int(os.environ.get("PORT", "8080"))

    timeout = int(os.environ.get("HEALTH_CHECK_TIMEOUT", "5"))
    path = "/ping"

    try:
        # 设置连接超时
        conn = http.client.HTTPConnection(host, port, timeout=timeout)
        # 发送请求
        conn.request("GET", path)
        # 获取响应
        response = conn.getresponse()

        # 检查状态码
        if response.status != 200:
            print(f"健康检查失败：状态码 {response.status}")
            return 1

        # 读取响应内容
        body = response.read().decode("utf-8")

        # 解析 JSON 并检查响应内容
        try:
            data = json.loads(body)
            if data.get("message") != "pong":
                print("健康检查失败：响应内容不符合预期")
                return 1
        except json.JSONDecodeError:
            print("健康检查失败：无法解析 JSON 响应")
            return 1

        print("健康检查通过")
        return 0

    except socket.timeout:
        print(f"健康检查失败：连接超时 ({timeout}秒)")
        return 1
    except ConnectionRefusedError:
        print(f"健康检查失败：连接被拒绝 {host}:{port}")
        return 1
    except Exception as e:
        print(f"健康检查失败：{e}")
        return 1
    finally:
        if "conn" in locals():
            conn.close()


if __name__ == "__main__":
    sys.exit(health_check())
