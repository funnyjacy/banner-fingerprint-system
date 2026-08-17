package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type BannerInput struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Banner string `json:"banner"`
}

type FingerprintResult struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Version    string  `json:"version"`
	OsHint     string  `json:"os_hint"`
	Confidence float64 `json:"confidence"`
}

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	inputFile := os.Getenv("INPUT_FILE")
	if inputFile == "" {
		inputFile = "/data/input.json"
	}

	log.Printf("Reading input from: %s", inputFile)
	log.Printf("Server URL: %s", serverURL)

	// 等待服务器启动
	if err := waitForServer(serverURL, 30*time.Second); err != nil {
		log.Fatalf("Server not available: %v", err)
	}

	// 读取输入文件
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	var inputs []BannerInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		log.Fatalf("Failed to parse input JSON: %v", err)
	}

	log.Printf("Loaded %d banner records", len(inputs))

	// 发送请求到服务器
	results, err := sendFingerprintRequest(serverURL, inputs)
	if err != nil {
		log.Fatalf("Failed to get fingerprint results: %v", err)
	}

	// 输出结果
	fmt.Println("\n=== Banner 指纹识别结果 ===\n")
	for _, result := range results {
		fmt.Printf("IP: %s  Port: %d\n", result.IP, result.Port)
		fmt.Printf("  协议: %s\n", result.Protocol)
		fmt.Printf("  产品: %s\n", result.Product)
		fmt.Printf("  版本: %s\n", result.Version)
		if result.OsHint != "" {
			fmt.Printf("  操作系统提示: %s\n", result.OsHint)
		}
		fmt.Printf("  置信度: %.2f\n", result.Confidence)
		fmt.Println()
	}

	// 输出 JSON 格式
	jsonOutput, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}

	fmt.Println("=== JSON 输出 ===")
	fmt.Println(string(jsonOutput))
}

func waitForServer(serverURL string, timeout time.Duration) error {
	healthURL := serverURL + "/health"
	client := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(timeout)

	log.Printf("Waiting for server to be ready...")

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Printf("Server is ready")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

func sendFingerprintRequest(serverURL string, inputs []BannerInput) ([]FingerprintResult, error) {
	endpoint := serverURL + "/fingerprint"

	jsonData, err := json.Marshal(inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []FingerprintResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}
