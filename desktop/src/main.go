package main

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"time"
	"bufio"
	"os"
	"strings"
	"encoding/json"
	"path/filepath"

	"github.com/atotto/clipboard"
)

var dt = time.Now()

func main() {
	http.HandleFunc("/set_clipboard", setClipboard)
	http.HandleFunc("/get_clipboard", getClipboard)
	config, lerr := loadConfigFile("lemon_push.conf")
	if lerr != nil {
        fmt.Println("加载配置lemon_push.conf失败:", lerr)
        return
    }
	port := ":"+config["port"] // 监听端口

	fmt.Println(dt.Format("2006-01-02 15:04:05"), "  服务端已启动")
	getLocalIP()
	fmt.Println(dt.Format("2006-01-02 15:04:05"), "  服务端监听端口:", config["port"])

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
	}
}

func getLocalIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(dt.Format("2006-01-02 15:04:05"), "  本机IP:", ipnet.IP.String())
			}
		}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

func setClipboard(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	code := values.Get("text")
	clipboard.WriteAll(code)
	fmt.Println("客户端 "+r.RemoteAddr+" 设置剪切板："+code)
	p := regexp.MustCompile(`https?://[^\s]+/[^/]+`)
	if p.MatchString(code) {
		matches := p.FindAllString(code, -1)
		for _, match := range matches {
			fmt.Printf("%s  启动浏览器打开链接：%s\n", dt.Format("2006-01-02 15:04:05"), match)
			openBrowser(match)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	sendStr := "ok"
	resp["data"] = sendStr
	resp["code"] = "0"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return
	}
	w.Write(jsonResp)
	return
}

func getClipboard(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	text, _ := clipboard.ReadAll()
	fmt.Println("客户端 "+r.RemoteAddr+" 获取剪切板："+text)
	resp["data"] = text
	resp["code"] = "0"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return
	}
	w.Write(jsonResp)
	return
}
func loadConfigFile(filename string) (map[string]string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execPath)
	filePath := filepath.Join(execDir, filename)

	file, err := os.Open(filePath)
	if err != nil {
		// 文件不存在，使用默认配置并创建文件
		config := make(map[string]string)
		config["port"] = "14756"

		// 创建文件并写入默认配置
		file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		for key, value := range config {
			_, err := writer.WriteString(key + "=" + value + "\n")
			if err != nil {
				return nil, err
			}
		}
		writer.Flush()

		return config, nil
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}


/* func loadConfigFile(filename string) (map[string]string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execPath)
	filePath := filepath.Join(execDir, filename)
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    config := make(map[string]string)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])
            config[key] = value
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return config, nil
} */