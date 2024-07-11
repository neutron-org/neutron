package main

import (
	"fmt"
	dbm "github.com/cosmos/cosmos-db"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cosmossdk.io/log"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NOTE: The functions below are copy-pasted from cosmos-sdk@v0.50.5 (see https://github.com/cosmos/cosmos-sdk/blob/v0.50.5/server/util.go#L102)
// Reason for this is that we want to remove cosmos-sdk override of commit_timeout to 5s (https://github.com/cosmos/cosmos-sdk/blob/v0.50.5/server/util.go#L252-L254)

// InterceptConfigsPreRunHandler is identical to InterceptConfigsAndCreateContext
// except it also sets the server context on the command and the server logger.
func InterceptConfigsPreRunHandler(cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}, cmtConfig *cmtcfg.Config) error {
	serverCtx, err := InterceptConfigsAndCreateContext(cmd, customAppConfigTemplate, customAppConfig, cmtConfig)
	if err != nil {
		return err
	}

	// overwrite default server logger
	logger, err := CreateSDKLogger(serverCtx, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	serverCtx.Logger = logger.With(log.ModuleKey, "server")

	// set server context
	return SetCmdServerContext(cmd, serverCtx)
}

// InterceptConfigsAndCreateContext performs a pre-run function for the root daemon
// application command. It will create a Viper literal and a default server
// Context. The server CometBFT configuration will either be read and parsed
// or created and saved to disk, where the server Context is updated to reflect
// the CometBFT configuration. It takes custom app config template and config
// settings to create a custom CometBFT configuration. If the custom template
// is empty, it uses default-template provided by the server. The Viper literal
// is used to read and parse the application configuration. Command handlers can
// fetch the server Context to get the CometBFT configuration or to get access
// to Viper.
func InterceptConfigsAndCreateContext(cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}, cmtConfig *cmtcfg.Config) (*server.Context, error) {
	serverCtx := server.NewDefaultContext()

	// Get the executable name and configure the viper instance so that environmental
	// variables are checked based off that name. The underscore character is used
	// as a separator.
	executableName, err := os.Executable()
	if err != nil {
		return nil, err
	}

	basename := path.Base(executableName)

	// configure the viper instance
	if err := serverCtx.Viper.BindPFlags(cmd.Flags()); err != nil {
		return nil, err
	}
	if err := serverCtx.Viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return nil, err
	}

	serverCtx.Viper.SetEnvPrefix(basename)
	serverCtx.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	serverCtx.Viper.AutomaticEnv()

	// intercept configuration files, using both Viper instances separately
	config, err := interceptConfigs(serverCtx.Viper, customAppConfigTemplate, customAppConfig, cmtConfig)
	if err != nil {
		return nil, err
	}

	// return value is a CometBFT configuration object
	serverCtx.Config = config
	if err = bindFlags(basename, cmd, serverCtx.Viper); err != nil {
		return nil, err
	}

	return serverCtx, nil
}

// CreateSDKLogger creates a the default SDK logger.
// It reads the log level and format from the server context.
func CreateSDKLogger(ctx *server.Context, out io.Writer) (log.Logger, error) {
	var opts []log.Option
	if ctx.Viper.GetString(flags.FlagLogFormat) == flags.OutputFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}
	opts = append(opts,
		log.ColorOption(!ctx.Viper.GetBool(flags.FlagLogNoColor)),
		// We use CometBFT flag (cmtcli.TraceFlag) for trace logging.
		log.TraceOption(ctx.Viper.GetBool(server.FlagTrace)))

	// check and set filter level or keys for the logger if any
	logLvlStr := ctx.Viper.GetString(flags.FlagLogLevel)
	if logLvlStr == "" {
		return log.NewLogger(out, opts...), nil
	}

	logLvl, err := zerolog.ParseLevel(logLvlStr)
	switch {
	case err != nil:
		// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
		filterFunc, err := log.ParseLogLevel(logLvlStr)
		if err != nil {
			return nil, err
		}

		opts = append(opts, log.FilterOption(filterFunc))
	default:
		opts = append(opts, log.LevelOption(logLvl))
	}

	return log.NewLogger(out, opts...), nil
}

// SetCmdServerContext sets a command's Context value to the provided argument.
// If the context has not been set, set the given context as the default.
func SetCmdServerContext(cmd *cobra.Command, serverCtx *server.Context) error {
	v := cmd.Context().Value(server.ServerContextKey)
	if v == nil {
		v = serverCtx
	}

	serverCtxPtr := v.(*server.Context)
	*serverCtxPtr = *serverCtx

	return nil
}

// interceptConfigs parses and updates a CometBFT configuration file or
// creates a new one and saves it. It also parses and saves the application
// configuration file. The CometBFT configuration file is parsed given a root
// Viper object, whereas the application is parsed with the private package-aware
// viperCfg object.
func interceptConfigs(rootViper *viper.Viper, customAppTemplate string, customConfig interface{}, cmtConfig *cmtcfg.Config) (*cmtcfg.Config, error) {
	rootDir := rootViper.GetString(flags.FlagHome)
	configPath := filepath.Join(rootDir, "config")
	cmtCfgFile := filepath.Join(configPath, "config.toml")

	conf := cmtConfig

	switch _, err := os.Stat(cmtCfgFile); {
	case os.IsNotExist(err):
		cmtcfg.EnsureRoot(rootDir)

		if err = conf.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("error in config file: %w", err)
		}

		defaultCometCfg := cmtcfg.DefaultConfig()
		// The SDK is opinionated about those comet values, so we set them here.
		// We verify first that the user has not changed them for not overriding them.
		if conf.RPC.PprofListenAddress == defaultCometCfg.RPC.PprofListenAddress {
			conf.RPC.PprofListenAddress = "localhost:6060"
		}

		cmtcfg.WriteConfigFile(cmtCfgFile, conf)

	case err != nil:
		return nil, err

	default:
		rootViper.SetConfigType("toml")
		rootViper.SetConfigName("config")
		rootViper.AddConfigPath(configPath)

		if err := rootViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read in %s: %w", cmtCfgFile, err)
		}
	}

	// Read into the configuration whatever data the viper instance has for it.
	// This may come from the configuration file above but also any of the other
	// sources viper uses.
	if err := rootViper.Unmarshal(conf); err != nil {
		return nil, err
	}

	conf.SetRoot(rootDir)

	appCfgFilePath := filepath.Join(configPath, "app.toml")
	if _, err := os.Stat(appCfgFilePath); os.IsNotExist(err) {
		if customAppTemplate != "" {
			config.SetConfigTemplate(customAppTemplate)

			if err = rootViper.Unmarshal(&customConfig); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", appCfgFilePath, err)
			}

			config.WriteConfigFile(appCfgFilePath, customConfig)
		} else {
			appConf, err := config.ParseConfig(rootViper)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", appCfgFilePath, err)
			}

			config.WriteConfigFile(appCfgFilePath, appConf)
		}
	}

	rootViper.SetConfigType("toml")
	rootViper.SetConfigName("app")
	rootViper.AddConfigPath(configPath)

	if err := rootViper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge configuration: %w", err)
	}

	return conf, nil
}

func bindFlags(basename string, cmd *cobra.Command, v *viper.Viper) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("bindFlags failed: %v", r)
		}
	}()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		err = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", basename, strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))))
		if err != nil {
			panic(err)
		}

		err = v.BindPFlag(f.Name, f)
		if err != nil {
			panic(err)
		}

		// Apply the viper config value to the flag when the flag is not set and
		// viper has a value.
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				panic(err)
			}
		}
	})

	return err
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.WriteCloser, err error) {
	if traceWriterFile == "" {
		return
	}
	return os.OpenFile(
		traceWriterFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0o666,
	)
}
