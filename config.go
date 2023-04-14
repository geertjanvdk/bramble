package bramble

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLogLevel = log.DebugLevel

	defaultPollIntervalString = "10s"

	defaultPortGateway = 8082
	defaultPortPrivate = 8083
	defaultPortMetrics = 9099

	defaultMaxServiceResponseSize = 1024 * 1024
)

var cfgLog = log.New()

func init() {
	cfgLog.SetLevel(log.InfoLevel)
	cfgLog.SetFormatter(&log.JSONFormatter{TimestampFormat: time.RFC3339Nano})
}

var (
	defaultAddressGateway = "0.0.0.0:" + strconv.Itoa(defaultPortGateway)
	defaultAddressPrivate = "0.0.0.0:" + strconv.Itoa(defaultPortPrivate)
	defaultAddressMetrics = "0.0.0.0:" + strconv.Itoa(defaultPortMetrics)
)

const envBrambleLogLevel = "BRAMBLE_LOG_LEVEL"

var Version = "dev"

// PluginConfig contains the configuration for the named plugin
type PluginConfig struct {
	Name   string
	Config json.RawMessage
}

// Config contains the gateway configuration
type Config struct {
	IdFieldName            string    `json:"id-field-name"`
	IdFieldType            string    `json:"id-field-type"`
	GatewayListenAddress   string    `json:"gateway-address"`
	DisableIntrospection   bool      `json:"disable-introspection"`
	MetricsListenAddress   string    `json:"metrics-address"`
	PrivateListenAddress   string    `json:"private-address"`
	GatewayPort            int       `json:"gateway-port"`
	MetricsPort            int       `json:"metrics-port"`
	PrivatePort            int       `json:"private-port"`
	Services               []string  `json:"services"`
	LogLevel               log.Level `json:"loglevel"`
	PollInterval           string    `json:"poll-interval"`
	PollIntervalDuration   time.Duration
	MaxRequestsPerQuery    int64 `json:"max-requests-per-query"`
	MaxServiceResponseSize int64 `json:"max-service-response-size"`
	Plugins                []PluginConfig
	// Config extensions that can be shared among plugins
	Extensions map[string]json.RawMessage
	// HTTP client to customize for downstream services query
	QueryHTTPClient *http.Client

	plugins          []Plugin
	executableSchema *ExecutableSchema
	watcher          *fsnotify.Watcher
	configFiles      []string
	linkedFiles      []string
}

func newConfig() *Config {
	return &Config{
		GatewayPort:            defaultPortGateway,
		PrivatePort:            defaultPortPrivate,
		MetricsPort:            defaultPortMetrics,
		LogLevel:               defaultLogLevel,
		PollInterval:           defaultPollIntervalString,
		MaxRequestsPerQuery:    50,
		MaxServiceResponseSize: defaultMaxServiceResponseSize,
	}
}

func (c *Config) addrOrPort(addr string, port int) string {
	if addr != "" {
		return addr
	}
	return fmt.Sprintf(":%d", port)
}

// GatewayAddress returns the host:port string of the gateway
func (c *Config) GatewayAddress() string {
	return c.addrOrPort(c.GatewayListenAddress, c.GatewayPort)
}

// PrivateAddress returns the address for private port
func (c *Config) PrivateAddress() string {
	return c.addrOrPort(c.PrivateListenAddress, c.PrivatePort)
}

func (c *Config) PrivateHttpAddress(path string) string {
	if c.PrivateListenAddress == "" {
		return fmt.Sprintf("http://localhost:%d/%s", c.PrivatePort, path)
	}
	return fmt.Sprintf("http://%s/%s", c.PrivateListenAddress, path)
}

// MetricAddress returns the address for the metric port
func (c *Config) MetricAddress() string {
	return c.addrOrPort(c.MetricsListenAddress, c.MetricsPort)
}

func (c *Config) handleConfigFile(f string) error {
	fp, err := os.Open(f)
	if err != nil {
		return err
	}
	defer func() { _ = fp.Close() }()

	if err := json.NewDecoder(fp).Decode(&c); err != nil {
		return fmt.Errorf("error decoding config file %q: %w", f, err)
	}
	return nil
}

// Load loads or reloads all the config files.
func (c *Config) Load() error {
	return c.load(false)
}

// Reload reloads all the config files.
func (c *Config) Reload() error {
	return c.load(true)
}

func (c *Config) load(isReload bool) error {
	prevLogLevel := c.LogLevel

	c.Extensions = nil
	// concatenate plugins from all the config files
	var plugins []PluginConfig
	for _, configFile := range c.configFiles {
		if err := c.handleConfigFile(configFile); err != nil {
			return err
		}
		plugins = append(plugins, c.Plugins...)
	}
	c.Plugins = plugins

	if strings.TrimSpace(c.IdFieldName) != "" {
		IdFieldName = c.IdFieldName
		IdFieldType = c.IdFieldType
	}

	logLevel, ok := os.LookupEnv(envBrambleLogLevel)
	if ok {
		if level, err := log.ParseLevel(logLevel); err == nil {
			c.LogLevel = level
		} else {
			cfgLog.WithField("loglevel", logLevel).Warn("invalid loglevel; using default")
			c.LogLevel = defaultLogLevel
		}
	}

	if isReload && prevLogLevel != c.LogLevel {
		cfgLog.WithField("from", prevLogLevel.String()).
			WithField("to", c.LogLevel.String()).
			Info("log level has changed")
	}
	log.SetLevel(c.LogLevel)

	var err error
	c.PollIntervalDuration, err = time.ParseDuration(c.PollInterval)
	if err != nil {
		return fmt.Errorf("invalid poll interval: %w", err)
	}

	services, err := c.buildServiceList()
	if err != nil {
		return err
	}
	c.Services = services

	c.plugins = c.ConfigurePlugins()

	return nil
}

func (c *Config) buildServiceList() ([]string, error) {
	serviceSet := map[string]bool{}
	for _, service := range c.Services {
		serviceSet[service] = true
	}
	for _, service := range strings.Fields(os.Getenv("BRAMBLE_SERVICE_LIST")) {
		serviceSet[service] = true
	}
	for _, plugin := range c.plugins {
		ok, path := plugin.GraphqlQueryPath()
		if ok {
			service := c.PrivateHttpAddress(path)
			serviceSet[service] = true
		}
	}
	services := []string{}
	for service := range serviceSet {
		services = append(services, service)
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("no services found in BRAMBLE_SERVICE_LIST or %s", c.configFiles)
	}
	return services, nil
}

// Watch starts watching the config files for change.
func (c *Config) Watch() {
	for {
		select {
		case err := <-c.watcher.Errors:
			log.WithError(err).Error("config watch error")
		case e := <-c.watcher.Events:
			shouldUpdate := false
			for i := range c.configFiles {
				log.WithFields(log.Fields{"event": e, "files": c.configFiles, "links": c.linkedFiles}).Debug("received config file event")
				// we want to reload the config if:
				// - the config file was updated, or
				// - the config file is a symlink and was changed (k8s config map update)
				if filepath.Clean(e.Name) == c.configFiles[i] && (e.Op == fsnotify.Write || e.Op == fsnotify.Create) {
					shouldUpdate = true
					break
				}
				currentFile, _ := filepath.EvalSymlinks(c.configFiles[i])
				if c.linkedFiles[i] != "" && c.linkedFiles[i] != currentFile {
					c.linkedFiles[i] = currentFile
					shouldUpdate = true
					break
				}
			}

			if !shouldUpdate {
				log.Debug("no configuration file reload needed")
				continue
			}

			if e.Op != fsnotify.Write && e.Op != fsnotify.Create {
				log.Debug("ignoring non write/create event")
				continue
			}

			err := c.Reload()
			if err != nil {
				cfgLog.WithError(err).Error("watcher failed reloading config")
			}
			cfgLog.WithField("services", c.Services).Info(c.LogLevel, "watcher reloaded configuration")
			err = c.executableSchema.UpdateServiceList(c.Services)
			if err != nil {
				cfgLog.WithError(err).Error("watcher failed updating services")
			}
			cfgLog.WithField("services", c.Services).Info(c.LogLevel, "watcher updated services")
		}
	}
}

// GetConfig returns operational config for the gateway
func GetConfig(configFiles []string) (*Config, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("could not create watcher: %w", err)
	}
	var linkedFiles []string
	for _, configFile := range configFiles {
		// watch the directory, else we'll lose the watch if the file is relinked
		err = watcher.Add(filepath.Dir(configFile))
		if err != nil {
			return nil, fmt.Errorf("error add file to watcher: %w", err)
		}
		linkedFile, _ := filepath.EvalSymlinks(configFile)
		linkedFiles = append(linkedFiles, linkedFile)
	}

	cfg := newConfig()
	cfg.watcher = watcher
	cfg.configFiles = configFiles
	cfg.linkedFiles = linkedFiles

	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ConfigurePlugins calls the Configure method on each plugin.
func (c *Config) ConfigurePlugins() []Plugin {
	var enabledPlugins []Plugin
	for _, pl := range c.Plugins {
		p, ok := RegisteredPlugins()[pl.Name]
		if !ok {
			log.Warnf("plugin %q not found", pl.Name)
			continue
		}
		err := p.Configure(c, pl.Config)
		if err != nil {
			log.WithError(err).Fatalf("error unmarshalling config for plugin %q: %s", pl.Name, err)
		}
		enabledPlugins = append(enabledPlugins, p)
	}

	return enabledPlugins
}

// Init initializes the config and does an initial fetch of the services.
func (c *Config) Init() error {
	var err error
	c.Services, err = c.buildServiceList()
	if err != nil {
		return fmt.Errorf("error building service list: %w", err)
	}

	var services []*Service
	for _, s := range c.Services {
		services = append(services, NewService(s))
	}

	queryClientOptions := []ClientOpt{WithMaxResponseSize(c.MaxServiceResponseSize), WithUserAgent(GenerateUserAgent("query"))}
	if c.QueryHTTPClient != nil {
		queryClientOptions = append(queryClientOptions, WithHTTPClient(c.QueryHTTPClient))
	}
	queryClient := NewClientWithPlugins(c.plugins, queryClientOptions...)
	es := NewExecutableSchema(c.plugins, c.MaxRequestsPerQuery, queryClient, services...)
	err = es.UpdateSchema(true)
	if err != nil {
		return err
	}

	c.executableSchema = es

	var pluginsNames []string
	for _, plugin := range c.plugins {
		plugin.Init(c.executableSchema)
		pluginsNames = append(pluginsNames, plugin.ID())
	}

	if len(pluginsNames) > 0 {
		cfgLog.WithField("plugins", pluginsNames).Info("plugins enabled")
	}

	return nil
}

type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}
