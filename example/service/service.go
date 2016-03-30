package service

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"

	"daemon.go"
	"daemon.go/example/server"
)

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "example", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "example", log.Ldate|log.Ltime)
}

type DaemonService struct {
	Id      string
	IsStart bool
	Server  *server.Server
}

func Manage(service *daemon.Service, mode int) (string, error) {

	usage := `Usage: 
					go run [-options] [argument]
						 
			options:
					-id 	  sevice id. only for arguments ：start. example -id=1 
					-port     daemon port. for every arguments. example -port=9988
			argument:
					start     service start
					stop      service stop
					status    service status
	`

	port := flag.Int("port", -1, "daemon listen port")
	id := flag.String("id", "", "service id")
	flag.Parse()
	if flag.NArg() != 1 {
		return usage, errors.New("param error")
	}
	commands := flag.Args()
	command := commands[0]
	stdlog.Println("id:", *id, ".port:", *port, ".argument:", command)

	if *port > 1 {
		switch command {
		case "start":
			if *id == "" {
				return usage, errors.New("id error")
			}
			return service.Start(*id, int32(*port), mode)
		case "stop":
			return service.Stop(int32(*port))
		case "status":
			return service.Status(int32(*port))
		default:
			return usage, nil
		}
	}
	return usage, nil
}

func (this *DaemonService) Install(id string, mode int) (string, error) {
	this.Id = id
	this.Server = &server.Server{}
	this.Server.Install(mode)
	return "Install success", nil
}

func (this *DaemonService) Start() (string, error) {
	this.IsStart = true
	this.Server.Start()
	return "DaemonService was started! port:" + strconv.Itoa(this.Server.Port), nil
}
func (this *DaemonService) Stop() (string, error) {
	this.Server.Stop()
	this.IsStart = false
	return "Success", nil
}
func (this *DaemonService) Status() (bool, error) {
	return this.IsStart, nil
}
func (this *DaemonService) GetId() string {
	return this.Id
}
