package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Proxy 结构体表示Docker配置中的代理设置。
type Proxy struct {
	Proxies struct {
		Default struct {
			HttpProxy  string `json:"httpProxy"`
			HttpsProxy string `json:"httpsProxy"`
			NoProxy    string `json:"noProxy"`
		} `json:"default"`
	} `json:"proxies"`
}

// NewProxy 通过给定的设置创建一个新的Proxy。
func NewProxy(httpProxy, httpsProxy, noProxy string) *Proxy {
	return &Proxy{
		Proxies: struct {
			Default struct {
				HttpProxy  string `json:"httpProxy"`
				HttpsProxy string `json:"httpsProxy"`
				NoProxy    string `json:"noProxy"`
			} `json:"default"`
		}{
			Default: struct {
				HttpProxy  string `json:"httpProxy"`
				HttpsProxy string `json:"httpsProxy"`
				NoProxy    string `json:"noProxy"`
			}{
				HttpProxy:  httpProxy,
				HttpsProxy: httpsProxy,
				NoProxy:    noProxy,
			},
		},
	}
}

// ReadJsonFile 读取文件中的JSON数据，并输出格式化后的JSON。
func ReadJsonFile(filePath string) ([]byte, error) {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件错误: %v", err)
	}

	var obj interface{}
	if err := json.Unmarshal(fileData, &obj); err != nil {
		return nil, fmt.Errorf("解码JSON错误: %v", err)
	}

	formattedJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("编排JSON错误: %v", err)
	}

	return formattedJSON, nil
}

// WriteToFile 将JSON数据写入文件。
func WriteToFile(filePath string, data interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("打开文件错误: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("将JSON写入文件错误: %v", err)
	}

	return nil
}

// BackupFile 创建文件的备份，并管理备份的保留。
func BackupFile(srcFile string, backupDir string, maxBackups int) error {
	// 确保备份目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	baseName := filepath.Base(srcFile)
	// 在备份文件名中变更时间戳格式
	backupFile := filepath.Join(backupDir, baseName+"."+time.Now().Format("2006-01-02 15:04:05"))
	err := CopyFile(srcFile, backupFile)
	if err != nil {
		return fmt.Errorf("复制文件到备份出错: %v", err)
	}

	// 列出文件并仅保留最新的'maxBackups'个文件
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("读取备份目录出错: %v", err)
	}

	// 筛选并收集备份文件
	var backupFiles []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		matched, _ := filepath.Match(baseName+".*", f.Name())
		if matched {
			backupFiles = append(backupFiles, f.Name())
		}
	}

	// 如果备份的文件数量超过了最大值，删除最老的文件
	if len(backupFiles) > maxBackups {
		sort.Slice(backupFiles, func(i, j int) bool {
			// 按包含时间戳的文件名排序
			return backupFiles[i] < backupFiles[j]
		})

		for _, f := range backupFiles[:len(backupFiles)-maxBackups] {
			err := os.Remove(filepath.Join(backupDir, f))
			if err != nil {
				return fmt.Errorf("删除旧备份文件失败: %v", err)
			}
		}
	}

	return nil
}

// CopyFile 将文件从源路径复制到目标路径。
func CopyFile(src, dest string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, input, 0644)
}

func main() {
	// 定义命令行参数
	httpProxyPtr := flag.String("httpProxy", "http://127.0.0.1:7890", "HTTP代理地址。")
	httpsProxyPtr := flag.String("httpsProxy", "http://127.0.0.1:7890", "HTTPS代理地址。")
	noProxyPtr := flag.String("noProxy", "localhost,127.0.0.1,.daocloud.io", "无代理设置。")
	onProxy := flag.Int("onProxy", -1, "代理设置 0: 关闭，1: 开启")

	flag.Parse()

	// 验证'onProxy'标记
	if *onProxy != 0 && *onProxy != 1 {
		fmt.Println("onProxy 参数错误, 请使用 0 或 1")
		return
	}

	// 获取用户的家目录并准备文件路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("获取家目录失败:", err)
		return
	}
	userDockerConfigPath := filepath.Join(homeDir, ".docker/config.json")
	// 判断文件是否存在, 如果不存在就创建空json文件
	if _, err = os.Stat(userDockerConfigPath); os.IsNotExist(err) {
		if err = WriteToFile(userDockerConfigPath, map[string]interface{}{}); err != nil {
			fmt.Println(err)
			return
		}
	}
	backupDir := filepath.Join(homeDir, ".docker/config_back")

	// 使用提供的标记创建新的代理
	uProxy := NewProxy(*httpProxyPtr, *httpsProxyPtr, *noProxyPtr)
	isOpen := map[int]string{0: "关闭", 1: "开启"}[*onProxy]
	fmt.Printf("配置如下:\nhttpProxy: %s\nhttpsProxy: %s\nnoProxy: %s\n代理开关: %s\n\n", uProxy.Proxies.Default.HttpProxy, uProxy.Proxies.Default.HttpsProxy, uProxy.Proxies.Default.NoProxy, isOpen)

	// 在修改Docker配置之前执行备份过程
	if err := BackupFile(userDockerConfigPath, backupDir, 5); err != nil {
		fmt.Println(err)
		return
	}

	// 检索当前的Docker配置
	currentConfig, err := ReadJsonFile(userDockerConfigPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v 原始内容\n%s", userDockerConfigPath, string(currentConfig))

	// 从文件解码当前配置
	var configData map[string]interface{}
	if err := json.Unmarshal(currentConfig, &configData); err != nil {
		fmt.Println("解码JSON错误:", err)
		return
	}

	// 根据'onProxy'标记启用或禁用代理
	if *onProxy == 1 {
		configData["proxies"] = uProxy.Proxies
	} else if *onProxy == 0 {
		delete(configData, "proxies")
	}

	// 将修改后的配置写回文件
	if err := WriteToFile(userDockerConfigPath, configData); err != nil {
		fmt.Println(err)
		return
	}

	// 检索并打印修改后的配置
	modifiedConfig, err := ReadJsonFile(userDockerConfigPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("\n\n%v 修改后内容\n%s", userDockerConfigPath, modifiedConfig)
}
