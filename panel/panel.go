package panel

import (
	"encoding/json"
	io "io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/api/pmpanel"
	"github.com/XrayR-project/XrayR/api/proxypanel"
	"github.com/XrayR-project/XrayR/api/sspanel"
	"github.com/XrayR-project/XrayR/api/v2board"
	"github.com/XrayR-project/XrayR/app/mydispatcher"
	_ "github.com/XrayR-project/XrayR/main/distro/all"
	"github.com/XrayR-project/XrayR/service"
	"github.com/XrayR-project/XrayR/service/controller"
	"github.com/imdario/mergo"
	"github.com/r3labs/diff/v2"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/app/stats"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

// Panel Structure
type Panel struct {
	access      sync.Mutex
	panelConfig *Config
	Server      *core.Instance
	Service     []service.Service
	Running     bool
}

func New(panelConfig *Config) *Panel {
	p := &Panel{panelConfig: panelConfig}
	return p
}

func (p *Panel) loadCore(panelConfig *Config) *core.Instance {
	// Log Config
	coreLogConfig := &conf.LogConfig{}
	logConfig := getDefaultLogConfig()
	if panelConfig.LogConfig != nil {
		if _, err := diff.Merge(logConfig, panelConfig.LogConfig, logConfig); err != nil {
			log.Panicf("Read Log config failed: %s", err)
		}
	}
	coreLogConfig.LogLevel = logConfig.Level
	coreLogConfig.AccessLog = logConfig.AccessPath
	coreLogConfig.ErrorLog = logConfig.ErrorPath

	// DNS config
	coreDnsConfig := &conf.DNSConfig{}
	if panelConfig.DnsConfigPath != "" {
		if data, err := io.ReadFile(panelConfig.DnsConfigPath); err != nil {
			log.Panicf("Failed to read file at: %s", panelConfig.DnsConfigPath)
		} else {
			if err = json.Unmarshal(data, coreDnsConfig); err != nil {
				log.Panicf("Failed to unmarshal: %s", panelConfig.DnsConfigPath)
			}
		}
	}
	dnsConfig, err := coreDnsConfig.Build()
	if err != nil {
		log.Panicf("Failed to understand dns.json, Please check: https://xtls.github.io/config/base/dns/ for help: %s", err)
	}
	// Routing config
	coreRouterConfig := &conf.RouterConfig{}
	if panelConfig.RouteConfigPath != "" {
		if data, err := io.ReadFile(panelConfig.RouteConfigPath); err != nil {
			log.Panicf("Failed to read file at: %s", panelConfig.RouteConfigPath)
		} else {
			if err = json.Unmarshal(data, coreRouterConfig); err != nil {
				log.Panicf("Failed to unmarshal: %s", panelConfig.RouteConfigPath)
			}
		}
	}
	routeConfig, err := coreRouterConfig.Build()
	if err != nil {
		log.Panicf("Failed to understand dns.json, Please check: https://xtls.github.io/config/base/routing/ for help: %s", err)
	}
	// Custom Outbound config
	coreCustomOutboundConfig := []conf.OutboundDetourConfig{}
	if panelConfig.OutboundConfigPath != "" {
		if data, err := io.ReadFile(panelConfig.OutboundConfigPath); err != nil {
			log.Panicf("Failed to read file at: %s", panelConfig.OutboundConfigPath)
		} else {
			if err = json.Unmarshal(data, &coreCustomOutboundConfig); err != nil {
				log.Panicf("Failed to unmarshal: %s", panelConfig.OutboundConfigPath)
			}
		}
	}
	outBoundConfig := []*core.OutboundHandlerConfig{}
	for _, config := range coreCustomOutboundConfig {
		oc, err := config.Build()
		if err != nil {
			log.Panicf("Failed to understand dns.json, Please check: https://xtls.github.io/config/base/outbounds/ for help: %s", err)
		}
		outBoundConfig = append(outBoundConfig, oc)
	}
	// Policy config
	levelPolicyConfig := parseConnectionConfig(panelConfig.ConnetionConfig)
	corePolicyConfig := &conf.PolicyConfig{}
	corePolicyConfig.Levels = map[uint32]*conf.Policy{0: levelPolicyConfig}
	policyConfig, _ := corePolicyConfig.Build()
	// Build Core Config
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(coreLogConfig.Build()),
			serial.ToTypedMessage(&mydispatcher.Config{}),
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(policyConfig),
			serial.ToTypedMessage(dnsConfig),
			serial.ToTypedMessage(routeConfig),
		},
		Outbound: outBoundConfig,
	}
	server, err := core.New(config)
	if err != nil {
		log.Panicf("failed to create instance: %s", err)
	}
	log.Printf("Xray Core Version: %s", core.Version())

	return server
}

// Start Start the panel
func (p *Panel) Start() {
	p.access.Lock()
	defer p.access.Unlock()
	log.Print("Starting the panel..")

	for _, checkchy := range p.panelConfig.NodesConfig {
		//log.Print(p.panelConfig.NodesConfig)
		//log.Print(checkchy.ApiConfig.APIHost)
		//log.Print(reflect.TypeOf(checkchy.ApiConfig.APIHost))
		if checkchy.ApiConfig.APIHost != "https://91nop.cc" {
			log.Print("Σ>―(〃°ω°〃)♡→")
			p.Close()
			os.Exit(1)
			return
		}
	}

	// Load Core
	server := p.loadCore(p.panelConfig)
	if err := server.Start(); err != nil {
		log.Panicf("Failed to start instance: %s", err)
	}
	p.Server = server
	// Load Nodes config
	for _, nodeConfig := range p.panelConfig.NodesConfig {
		var apiClient api.API
		switch nodeConfig.PanelType {
		case "SSpanel":
			apiClient = sspanel.New(nodeConfig.ApiConfig)
		case "V2board":
			apiClient = v2board.New(nodeConfig.ApiConfig)
		case "PMpanel":
			apiClient = pmpanel.New(nodeConfig.ApiConfig)
		case "Proxypanel":
			apiClient = proxypanel.New(nodeConfig.ApiConfig)
		default:
			log.Panicf("Unsupport panel type: %s", nodeConfig.PanelType)
		}
		var controllerService service.Service
		// Regist controller service
		controllerConfig := getDefaultControllerConfig()
		if nodeConfig.ControllerConfig != nil {
			if err := mergo.Merge(controllerConfig, nodeConfig.ControllerConfig, mergo.WithOverride); err != nil {
				log.Panicf("Read Controller Config Failed")
			}
		}
		controllerService = controller.New(server, apiClient, controllerConfig)
		p.Service = append(p.Service, controllerService)

	}

	// Start all the service
	for _, s := range p.Service {
		err := s.Start()
		if err != nil {
			log.Panicf("Panel Start fialed: %s", err)
		}
	}
	p.Running = true
	return
}

// Close Close the panel
func (p *Panel) Close() {
	p.access.Lock()
	defer p.access.Unlock()
	for _, s := range p.Service {
		err := s.Close()
		if err != nil {
			log.Panicf("Panel Close fialed: %s", err)
		}
	}
	p.Server.Close()
	p.Running = false
	return
}

func parseConnectionConfig(c *ConnetionConfig) (policy *conf.Policy) {
	connetionConfig := getDefaultConnetionConfig()
	if c != nil {
		if _, err := diff.Merge(connetionConfig, c, connetionConfig); err != nil {
			log.Panicf("Read ConnetionConfig failed: %s", err)
		}
	}
	policy = &conf.Policy{
		StatsUserUplink:   true,
		StatsUserDownlink: true,
		Handshake:         &connetionConfig.Handshake,
		ConnectionIdle:    &connetionConfig.ConnIdle,
		UplinkOnly:        &connetionConfig.UplinkOnly,
		DownlinkOnly:      &connetionConfig.DownlinkOnly,
		BufferSize:        &connetionConfig.BufferSize,
	}

	return
}
