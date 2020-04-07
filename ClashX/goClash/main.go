package main // import "github.com/yichengchen/clashX/ClashX"
import (
	"C"
	"encoding/json"
	"github.com/Dreamacro/clash/config"
	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub/executor"
	"github.com/Dreamacro/clash/hub/route"
	"github.com/phayes/freeport"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

func isAddrValid(addr string) bool {
	if addr != "" {
		comps := strings.Split(addr, ":")
		v := comps[len(comps)-1]
		if port, err := strconv.Atoi(v); err == nil {
			if port > 0 && port < 65535 && checkPortAvailable(port, false) {
				return true
			}
		}
	}
	return false
}

func checkPortAvailable(port int, lan bool) bool {
	var addr string
	if port < 1 || port > 65534 {
		return false
	}
	if lan {
		addr = ":"
	} else {
		addr = "127.0.0.1:"
	}
	l, err := net.Listen("tcp", addr+strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

func parseConfig(checkPort bool) (*config.Config, error) {
	configFile := filepath.Join(constant.Path.HomeDir(), constant.Path.Config())
	constant.SetConfig(configFile)

	cfg, err := executor.Parse()
	if err != nil {
		return nil, err
	}
	if checkPort {
		if !isAddrValid(cfg.General.ExternalController) {
			port, err := freeport.GetFreePort()
			if err != nil {
				return nil, err
			}
			cfg.General.ExternalController = "127.0.0.1:" + strconv.Itoa(port)
			cfg.General.Secret = ""
		}
		lan := cfg.General.AllowLan

		if !checkPortAvailable(cfg.General.Port, lan) {
			if port, err := freeport.GetFreePort(); err == nil {
				cfg.General.Port = port
			}
		}

		if !checkPortAvailable(cfg.General.SocksPort, lan) {
			if port, err := freeport.GetFreePort(); err == nil {
				cfg.General.SocksPort = port
			}
		}
	}

	go route.Start(cfg.General.ExternalController, cfg.General.Secret)

	executor.ApplyConfig(cfg, true)
	return cfg, nil
}

//export verifyClashConfig
func verifyClashConfig(content *C.char) *C.char {

	b := []byte(C.GoString(content))
	cfg, err := executor.ParseWithBytes(b)
	if err != nil {
		return C.CString(err.Error())
	}

	if len(cfg.Proxies) < 1 {
		return C.CString("No proxy found in config")
	}
	return C.CString("success")
}

//export run
func run(checkConfig bool) *C.char {
	cfg, err := parseConfig(checkConfig)
	if err != nil {
		return C.CString(err.Error())
	}

	portInfo := map[string]string{
		"externalController": cfg.General.ExternalController,
		"secret":             cfg.General.Secret,
	}

	jsonString, err := json.Marshal(portInfo)
	if err != nil {
		return C.CString(err.Error())
	}

	return C.CString(string(jsonString))
}

//export setUIPath
func setUIPath(path *C.char) {
	route.SetUIPath(C.GoString(path))
}

//export clashUpdateConfig
func clashUpdateConfig(path *C.char) *C.char {
	cfg, err := executor.ParseWithPath(C.GoString(path))
	if err != nil {
		return C.CString(err.Error())
	}
	executor.ApplyConfig(cfg, false)
	return C.CString("success")
}

//export clashGetConfigs
func clashGetConfigs() *C.char {
	general := executor.GetGeneral()
	jsonString, err := json.Marshal(general)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(string(jsonString))
}

func main() {
}
