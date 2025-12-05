package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// 并发安全的 writer
type safeWriter struct {
	mtx sync.Mutex
	w   http.ResponseWriter
}

func (sw *safeWriter) Write(p []byte) (int, error) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()
	return sw.w.Write(p)
}

// 把远端流拷贝到本地流，按行转发（SSE 场景）
func copyStream(model string, url string, dst io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(dst, "【%s 错误】%v\n", model, err)
		return
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		// 可选：加上模型前缀，方便前端区分
		fmt.Fprintf(dst, "[%s] %s\n", model, line)
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(dst, "【%s 读取出错】%v\n", model, err)
	}
}

// 接口处理器
func streamTwoModels(w http.ResponseWriter, r *http.Request) {
	// 设置分块传输
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sw := &safeWriter{w: w}
	var wg sync.WaitGroup

	// 假设本地起了两个 mock 流服务
	modelA := "http://localhost:8001/stream"
	modelB := "http://localhost:8002/stream"

	wg.Add(2)
	go copyStream("modelA", modelA, sw, &wg)
	go copyStream("modelB", modelB, sw, &wg)

	// 等待两边全部结束
	wg.Wait()
	flusher.Flush()
}

func main() {
	// 第一段 mock，监听 :8001
	go func() {
		mux1 := http.NewServeMux()
		mux1.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			flusher := w.(http.Flusher)
			for i := 0; i < 10; i++ {
				fmt.Fprintf(w, "data: hello %d\n\n", i)
				flusher.Flush()
				time.Sleep(200 * time.Millisecond)
			}
		})
		log.Println(http.ListenAndServe(":8001", mux1))
	}()

	// 第二段 mock，监听 :8002
	go func() {
		mux2 := http.NewServeMux()
		mux2.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			flusher := w.(http.Flusher)
			for i := 0; i < 10; i++ {
				fmt.Fprintf(w, "data: world %d\n\n", i)
				flusher.Flush()
				time.Sleep(200 * time.Millisecond)
			}
		})
		log.Println(http.ListenAndServe(":8002", mux2))
	}()

	// 真正的业务接口
	http.HandleFunc("/merge", streamTwoModels)
	log.Println("listen :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
