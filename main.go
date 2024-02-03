package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
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
	proxy := &Proxy{}
	proxy.Proxies.Default.HttpProxy = httpProxy
	proxy.Proxies.Default.HttpsProxy = httpsProxy
	proxy.Proxies.Default.NoProxy = noProxy
	return proxy
}

type DockerConfig struct {
	Proxy      Proxy
	MapConfig  map[string]interface{}
	ConfigPath string
	BackupDir  string
}

func NewDockerConfig(userProxy *Proxy) (*DockerConfig, error) {
	dockerConfig := &DockerConfig{}
	var err error

	dockerConfig.Proxy = *userProxy
	if err = dockerConfig.InitPath(); err != nil {
		return dockerConfig, err
	}
	if dockerConfig.MapConfig, err = dockerConfig.ReadConfigFile(); err != nil {
		return dockerConfig, err
	}

	return dockerConfig, nil
}

// InitPath 获取用户的家目录并准备文件路径
func (d *DockerConfig) InitPath() error {
	// 获取用户的家目录并准备文件路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.New(err.Error())
	}
	d.ConfigPath = filepath.Join(homeDir, ".docker/config.json")
	d.BackupDir = filepath.Join(homeDir, ".docker/config_back")
	return nil
}

// ReadConfigFile 读取Docker配置文件
func (d *DockerConfig) ReadConfigFile() (map[string]interface{}, error) {
	var configData map[string]interface{}
	fileData, err := ioutil.ReadFile(d.ConfigPath)
	if err != nil {
		if _, err = os.Stat(d.ConfigPath); os.IsNotExist(err) {
			// 写入一个空json文件
			fileData = []byte("{}")
			if err = ioutil.WriteFile(d.ConfigPath, fileData, 0644); err != nil {
				return configData, errors.New(err.Error())
			}
		} else {
			return configData, errors.New(err.Error())
		}
	}
	if err = json.Unmarshal(fileData, &configData); err != nil {
		return configData, errors.New(err.Error())
	}
	fmt.Printf("原始内容\n%s\n", string(fileData))
	return configData, nil
}

// UpdateConfig 更新Docker配置文件
func (d *DockerConfig) UpdateConfig(onProxy *int) error {
	if err := BackupFile(d.ConfigPath, d.BackupDir, 5); err != nil {
		return errors.New(err.Error())
	}
	if *onProxy == 1 {
		d.MapConfig["proxies"] = d.Proxy
	} else if *onProxy == 0 {
		delete(d.MapConfig, "proxies")
	}
	return nil
}

// WriteConfigFile 写入Docker配置文件
// WriteConfigFile 写入Docker配置文件
func (d *DockerConfig) WriteConfigFile() error {
	jsonData := Map2SJson(d.MapConfig)
	fmt.Printf("修改后\n%s\n", jsonData)

	file, err := os.OpenFile(d.ConfigPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 直接写入JSON字符串到文件中
	if _, err = file.WriteString(jsonData); err != nil {
		return errors.New(err.Error())
	}
	return nil
}

// Map2SJson 将map转换为格式化的JSON字符串
func Map2SJson(m map[string]interface{}) string {
	s, _ := json.MarshalIndent(m, "", "    ")
	return string(s)
}

// BackupFile 创建文件的备份，并管理备份的保留。
func BackupFile(srcFile string, backupDir string, maxBackups int) error {
	// 确保备份目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return errors.New(err.Error())
	}

	baseName := filepath.Base(srcFile)
	// 在备份文件名中变更时间戳格式
	backupFile := filepath.Join(backupDir, baseName+"."+time.Now().Format("2006-01-02 15:04:05"))
	err := CopyFile(srcFile, backupFile)
	if err != nil {
		return errors.New(err.Error())
	}

	// 列出文件并仅保留最新的'maxBackups'个文件
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return errors.New(err.Error())
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
			err = os.Remove(filepath.Join(backupDir, f))
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}

	return nil
}

// CopyFile 将文件从源路径复制到目标路径。
func CopyFile(src, dest string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return errors.New(err.Error())
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
	uProxy := NewProxy(*httpProxyPtr, *httpsProxyPtr, *noProxyPtr)
	// 使用提供的标记创建新的代理
	isOpen := map[int]string{0: "关闭", 1: "开启"}[*onProxy]
	fmt.Printf("配置如下:\nhttpProxy: %s\nhttpsProxy: %s\nnoProxy: %s\n代理开关: %s\n\n", uProxy.Proxies.Default.HttpProxy, uProxy.Proxies.Default.HttpsProxy, uProxy.Proxies.Default.NoProxy, isOpen)

	dockerConfig, err := NewDockerConfig(uProxy)
	if err != nil {
		fmt.Printf("error NewDockerConfig: %+v", err)
		return
	}

	err = dockerConfig.UpdateConfig(onProxy)
	if err != nil {
		fmt.Printf("error UpdateConfig: %+v", err)
		return
	}

	err = dockerConfig.WriteConfigFile()
	if err != nil {
		fmt.Printf("error WriteConfigFile: %+v", err)
		return
	}
}
