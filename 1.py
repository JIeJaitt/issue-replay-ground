#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import uuid
import websocket
import _thread as thread
import sys

URL = ("ws://operationapi-sit.fosunhanig.com/operation-support-api/cs/v1/ws?"
       "session_id=561c3c6b-615a-4abd-9f68-7d7c11833870&token=test")

# URL =  "ws://operationapi-sit.fotechwealth.com.local/operation-support-api/cs/v1/ws" + \
#        "?session_id=561c3c6b-615a-4abd-9f68-7d7c11833870&token=test"

def on_open(ws):
    print("【INFO】WebSocket 已连接")
    # 随便发一条心跳，可按需删掉
    ws.send("hello")

def on_message(ws, message):
    print("【RCV】", message)

def on_error(ws, error):
    print("【ERR】", error)

def on_close(ws, close_status_code, close_msg):
    print(f"【CLOSE】code={close_status_code}  reason={close_msg}")

if __name__ == '__main__':
    headers = {"X-Request-ID": str(uuid.uuid4())}
    ws = websocket.WebSocketApp(URL,
                                header=headers,
                                on_open=on_open,
                                on_message=on_message,
                                on_error=on_error,
                                on_close=on_close)

    # 允许 Ctrl-C 退出
    try:
        ws.run_forever()
    except KeyboardInterrupt:
        print("\n【INFO】用户中断，关闭连接")
        ws.close()
        sys.exit(0)