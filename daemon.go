// daemon
package daemon

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type Daemon interface {
	Install(id string, mode int) (string, error)

	Start() (string, error)

	Stop() (string, error)

	Status() (bool, error)

	GetId() string
}

const (
	Mode_Product    = 1
	Mode_Standalone = 2
	Mode_Test       = 3
	Mode_Dev        = 4
)

var stdlog, errlog *log.Logger

type Service struct {
	DaemonService Daemon

	conn net.Conn
	//系统信号的channel
	interrupt chan os.Signal
	//传递返回结果的channel
	resultChan chan *Result
	//tcp回写完成的信号channel
	writeOk chan string
}
type Result struct {
	Id  string `json:"Id"`
	Ok  bool   `json:"Ok"`
	Err bool   `json:"Err"`
}

//id 当前服务的id port是命令监听的port
//mode 参数 是根据环境不同而加载不同的配置文件
func (service *Service) Start(id string, port int32, mode int) (string, error) {
	stdlog.Println("Service install......")
	str, err := service.DaemonService.Install(id, mode)
	if err != nil {
		return str, err
	}
	stdlog.Println("Service install success!")
	stdlog.Println("Service start......")
	str1, err1 := service.DaemonService.Start()
	if err1 != nil {
		return str1, err1
	}
	stdlog.Println(str1, " id:", service.DaemonService.GetId())

	service.interrupt = make(chan os.Signal, 1)
	signal.Notify(service.interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	stdlog.Println("Daemon listener start ......")
	portStr := ":" + strconv.Itoa(int(port))
	listener, err := net.Listen("tcp", portStr)
	if err != nil {
		return "Possibly was a problem with the port binding", err
	}
	stdlog.Println("Daemon listener start success! port:", port)
	listen := make(chan net.Conn, 100)
	service.resultChan = make(chan *Result)
	service.writeOk = make(chan string)
	go service.acceptConnection(listener, listen)

	for {
		select {
		case conn := <-listen:
			go service.handleClient(conn)
		case killSignal := <-service.interrupt:
			str, err := service.DaemonService.Stop()
			var ret *Result
			if err != nil {
				errlog.Println(str, "\nError: ", err)
				ret = &Result{
					Id:  service.DaemonService.GetId(),
					Ok:  false,
					Err: true,
				}
			} else {
				ret = &Result{
					Id:  service.DaemonService.GetId(),
					Ok:  true,
					Err: false,
				}
			}

			if killSignal == syscall.SIGTERM {
				//发送结果 给handleClient
				service.resultChan <- ret
				select {
				//等待handleClient的信号
				case ok := <-service.writeOk:
					stdlog.Println(ok)
				}
			}

			stdlog.Println("Got signal:", killSignal)
			stdlog.Println("Stoping listening on ", listener.Addr())
			listener.Close()
			if killSignal == os.Interrupt {
				return "Daemon was interruped by system signal", nil
			} else if killSignal == syscall.SIGTERM {
				return "Daemon was sigterm by stop command", nil
			}
			return "Daemon was killed", nil
		}
	}

	return "Service and daemon listener start success!", nil
}
func (service *Service) Stop(port int32) (string, error) {
	stdlog.Println("Service stop......")
	str, err := service.clientStart(port)
	if err != nil {
		return str, err
	}
	stdlog.Println(str)
	stdlog.Println("Send stop command to daemon listener......")
	service.conn.Write([]byte("stop"))
	for {
		result := make([]byte, 1024)
		read_len, err := service.conn.Read(result)
		if read_len == 0 || err != nil {
			return "Command : status .Data is null", errors.New("Daemon listen reponse is null")
		}
		buf := result[:read_len]
		stdlog.Println("buf:", string(buf))
		ret := &Result{}
		json.Unmarshal(buf, ret)
		if ret.Ok {
			return "Service is stopped!", nil
		} else {
			return "Command : stop .", errors.New("Service is not running!")
		}
	}
}
func (service *Service) Status(port int32) (string, error) {
	stdlog.Println("Get service status......")
	str, err := service.clientStart(port)
	if err != nil {
		return str, err
	}
	stdlog.Println(str)
	stdlog.Println("Send status command to daemon listener......")
	service.conn.Write([]byte("status"))
	for {
		result := make([]byte, 1024)
		read_len, err := service.conn.Read(result)
		if read_len == 0 || err != nil {
			return "Command : status .Data is null", errors.New("Daemon listen reponse is null")
		}
		buf := result[:read_len]
		stdlog.Println("buf:", string(buf))
		ret := &Result{}
		json.Unmarshal(buf, ret)
		stdlog.Println("ok:", ret.Ok, "ret", ret)
		if ret.Ok {
			stdlog.Println("Get Service status success!")
			return "Service is running!", nil
		} else {
			return "Command : status .", errors.New("Service not running")
		}
	}
}

func (service *Service) read(conn net.Conn) (string, error) {

	return "Get Service status success!", nil
}

func (service *Service) clientStart(port int32) (string, error) {
	var err error
	portStr := ":" + strconv.Itoa(int(port))
	stdlog.Println("Connect to port" + portStr)
	service.conn, err = net.Dial("tcp", portStr)
	if err != nil {
		return "Daemon listener is not runnig port" + portStr, err
	}
	return "Connect success! port" + portStr, nil
}
func (service *Service) acceptConnection(listener net.Listener, listen chan<- net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		listen <- conn
	}
}

func (service *Service) handleClient(client net.Conn) {
	for {
		result := make([]byte, 1024)
		read_len, err := client.Read(result)
		if read_len == 0 || err != nil {
			return
		}
		command := string(result[:read_len])

		var ret *Result
		switch command {
		case "stop":
			//发送停止信号
			service.interrupt <- syscall.SIGTERM
			select {
			//等待start方法的回写结果
			case ret = <-service.resultChan:

				buf1, _ := json.Marshal(ret)
				client.Write(buf1)
			}
			//发送回写成功信号
			service.writeOk <- "data write ok!"
			break
		case "status":
			status, _ := service.DaemonService.Status()
			stdlog.Println("status:", status)
			if status {
				ret = &Result{
					Id:  service.DaemonService.GetId(),
					Ok:  true,
					Err: false,
				}
			} else {
				ret = &Result{
					Id:  service.DaemonService.GetId(),
					Ok:  false,
					Err: false,
				}
			}
			buf1, _ := json.Marshal(ret)
			client.Write(buf1)
			break
		default:
			ret = &Result{
				Id:  service.DaemonService.GetId(),
				Ok:  false,
				Err: false,
			}
			buf1, _ := json.Marshal(ret)
			client.Write(buf1)
			break
		}

	}
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}
func (service *Service) Standlaone(id string, port int32) (string, error) {
	return service.Start(id, port, Mode_Standalone)
}
