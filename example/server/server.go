// server
package server

import (
	"fmt"

	"daemon.go"
	"github.com/Terry-Mao/goconf"
)

const (
	Standlone = "resources/standlone/"
	Release   = "resources/release/"
	File      = "app.conf"
)

type Server struct {
	Port int
	Mode string
}
type MyConfig struct {
	Mode string `goconf:"base:mode"`
	Port int    `goconf:"group1:port"`
}

func (this *Server) Start() {
	fmt.Println("**************服务开始启动****************")
	fmt.Println(" mode:", this.Mode, " port:", this.Port)
	fmt.Println("**************服务启动完毕****************")
}
func (this *Server) Stop() {
	fmt.Println("**************服务开始关闭****************")
	fmt.Println(" mode:", this.Mode, " port:", this.Port)
	fmt.Println("**************服务关闭完毕****************")
}
func (this *Server) Install(mode int) {
	fmt.Println("**************服务初始化****************")
	fmt.Println("mode:", mode)
	fmt.Println("**************读取配置文件****************")
	myconf, _ := initConfig(mode)
	this.Mode = myconf.Mode
	this.Port = myconf.Port
	fmt.Println(" mode:", this.Mode, " port:", this.Port)
	fmt.Println("**************服务初始化完毕****************")
}
func initConfig(mode int) (*MyConfig, error) {
	file := ""
	if mode == daemon.Mode_Standalone {
		file = Standlone + File
	} else {
		file = Release + File
	}
	gconf := goconf.New()
	if err := gconf.Parse(file); err != nil {
		fmt.Println(file)
		return nil, err
	}
	myConfig := &MyConfig{}
	gconf.Unmarshal(myConfig)
	return myConfig, nil
}
