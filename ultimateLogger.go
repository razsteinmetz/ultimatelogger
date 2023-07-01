package ultimatelogger

import (
	_ "embed"
	"fmt"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"io"
	"log"
	"os"
)

/*
UltimateLogger - a versitle logger that can log to console, file, telegram, pushover and is configurable via a toml file.
*/

var ulogger *logrus.Logger
var defaultText prefixed.TextFormatter

// create the default config if it does not exist in the client direcotry
// note the file will be created in the program using the package.
//
//go:embed "defaultConfig.toml"
var DEFAULTCONFIG []byte

//go:embed "template.toml"
var TEMPLATECONFIG []byte

const defaultConfigPath = "defaultConfig.toml"
const templatePath = "ultimatelogger.toml"

//	func CreateDefaultConfig() error {
//		//fmt.Println(string(DEFAULTCONFIG))
//		if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
//			err := os.WriteFile(defaultConfigPath, DEFAULTCONFIG, 0666)
//			if err != nil {
//				fmt.Println("Error creating default config file:", err)
//				return err
//			}
//		}
//		return nil
//	}
//
// create a config file if its does not exist
func CreateConfig(fpath string, content *[]byte) error {
	//fmt.Println(string(DEFAULTCONFIG))
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		err := os.WriteFile(fpath, *content, 0666)
		if err != nil {
			fmt.Printf("Error creating %s file: %s", fpath, err)
			return err
		}
	}
	return nil
}

func load_default_config() {
	viper.SetConfigType("toml") // REQUIRED if the config file does not have the extension in the name
	//viper.SafeWriteConfigAs("./defaultConfig.toml")
	viper.SetConfigName("defaultConfig.toml") // name of config file (without extension)
	viper.SetConfigType("toml")               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error reading default config file: %w", err))
	}
	//fmt.Printf("%+v\n", viper.AllSettings())
	return
}

// load the configration from a file.  Note that usually the default config is loaded first
// and then the user file is loaded and merged in.
func load_config_from_file(configFileName string) {
	viper.SetConfigName(configFileName) // name of config file (without extension)
	viper.SetConfigType("toml")         // REQUIRED if the config file does not have the extension in the name
	//viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.AddConfigPath(".")     // optionally look for config in the working directory
	err := viper.MergeInConfig() // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		fmt.Println("Error in reading user config, is it missing?", err.Error())
	}

	return
}

func get_level_string(lvl string, defaultLevel logrus.Level) logrus.Level {
	if lvl == "" {
		return defaultLevel
	}
	l, err := logrus.ParseLevel(lvl)
	if err != nil {
		fmt.Printf("Error parsing level %s: %v\n", lvl, err)
		return defaultLevel
	}
	return l
}

// get the level from the viper key
func get_level(viperkey string, defaultLevel logrus.Level) logrus.Level {
	level := viper.GetString(viperkey + ".level")
	if level == "" {
		return defaultLevel
	}
	l, err := logrus.ParseLevel(level)
	if err != nil {
		fmt.Printf("Error parsing level %s: %v\n", level, err)
		return defaultLevel
	}
	return l
}

// get_formatter load the default formatter and then see if there are any overrides in the specific section
// if so update it.  returns a new instance of a formatter
func get_formatter(viperkey string, defultText prefixed.TextFormatter) logrus.Formatter {
	var jsonFormatter logrus.JSONFormatter
	var textFormatter prefixed.TextFormatter
	jsonFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	format := viper.GetString(viperkey + ".Format")
	// if the format is json, then use the json formatter with the default settings
	if format == "json" {
		return &jsonFormatter
	} else {
		// create a copy of the default text formatter
		textFormatter = defultText
	}
	// look for overrides
	if viper.IsSet(viperkey + ".text") {
		err := viper.UnmarshalKey(viperkey+".text", &textFormatter)
		if err != nil {
			fmt.Printf("Error unmarshal %s.text: %s\n", viperkey, err.Error())
		}
	}
	return &textFormatter
}

// create console main logger
func log_console() {
	if !viper.GetBool("console") {
		ulogger.Out = io.Discard
		return
	}
	format := get_formatter("ConsoleConfig", defaultText)
	ulogger.Formatter = format
	level := get_level("ConsoleConfig", logrus.InfoLevel)
	//lvl := viper.GetString("ConsoleConfig.Level")
	//if lvl == "" {
	//	lvl = "info"
	//}
	//level, err := logrus.ParseLevel(lvl)
	//if err != nil {
	//	fmt.Println("Error parsing level:", err.Error())
	//	level = logrus.InfoLevel
	//}
	ulogger.SetLevel(level)
	return
}

// handle the rotating log hook
func log_rotating() {
	if !viper.GetBool("rotatefile") {
		return // not enabled
	}
	var cfg RotateFileConfig
	err := viper.UnmarshalKey("RotateFileConfig", &cfg)
	if err != nil {
		fmt.Println("Error unmarshal RotateFileConfig:", err.Error())
		return
	}
	//lg := cfg.LogFile
	//lvl := get_level_string(cfg.Level, logrus.InfoLevel)
	//
	//lvl, err := logrus.ParseLevel(cfg.Level)
	//if err != nil {
	//	lvl = logrus.InfoLevel
	//}
	//format := get_formatter("RotateFileConfig", defaultText)
	hook, err := lumberjackrus.NewHook(

		&cfg.LogFile,
		get_level_string(cfg.Level, logrus.InfoLevel),
		get_formatter("RotateFileConfig", defaultText),
		&lumberjackrus.LogFileOpts{})

	fmt.Println("adding hook")
	ulogger.AddHook(hook)
}

func log_pushover() {
	if !viper.GetBool("pushover") {
		return // not enabled
	}
	//hook, err := NewPushoverHook("PUSH_OVER_USER_TOKEN", "PUSH_OVER_API_TOKEN")
	lvl, err := logrus.ParseLevel(viper.GetString("PushoverConfig.Level"))
	if err != nil {
		lvl = logrus.InfoLevel
	}
	format := get_formatter("PushoverConfig", defaultText)
	uk := viper.GetString("PushoverConfig.UserKey")
	ak := viper.GetString("PushoverConfig.APIKey")
	ff := viper.GetBool("PushoverConfig.FixedFont")
	hook, err := NewPushoverHook(uk, ak, lvl, format, ff)

	if err != nil {
		log.Panic("Error creating PushoverHook:", err.Error())
		return
	}
	//ff := viper.GetBool("PushoverConfig.FixedFont")
	//hook := logrusPushover.NewPushoverHook(uk, ak, true, lvl, "", ff)
	ulogger.AddHook(hook)
}

func log_telegram() {
	if !viper.GetBool("telegram") {
		return // not enabled
	}
	//hook, err := NewPushoverHook("PUSH_OVER_USER_TOKEN", "PUSH_OVER_API_TOKEN")
	lvl, err := logrus.ParseLevel(viper.GetString("TelegramConfig.Level"))
	if err != nil {
		lvl = logrus.InfoLevel
	}
	format := get_formatter("TelegramConfig", defaultText)
	chatid := viper.GetInt64("TelegramConfig.ChatID")
	ak := viper.GetString("TelegramConfig.APIKey")

	hook := NewTelegramHook(ak, chatid, lvl, format)
	ulogger.AddHook(hook)
}

// UltimateLogger returns a pointer to the global logger to be used.
// it takes the configration from the toml files and sets up the logger
func UltimateLogger() *logrus.Logger {
	err := CreateConfig(defaultConfigPath, &DEFAULTCONFIG)
	if err != nil {
		log.Panic("Error creating default config:", err.Error())
	}
	err = CreateConfig(templatePath, &TEMPLATECONFIG)
	if err != nil {
		log.Panic("Error creating template config:", err.Error())
	}
	load_default_config()
	load_config_from_file("ultimateLogger.toml")
	// load the default text formatter configuration to a global variable for use in other formatters
	err = viper.UnmarshalKey("text", &defaultText)
	if err != nil {
		fmt.Println("Unmarshal Error:", err.Error())
	}
	ulogger = logrus.New()
	// create each hook
	log_console()
	log_rotating()
	log_pushover()
	log_telegram()
	return ulogger
	////for i := 0; i < 12; i++ {
	////	ulogger.Warning("testingtestingtestingtestingtestingtestingtestingtestingtestingtestingtestingtestingtestingtestingtestingtesting", i)
	////}
	//ulogger.Info("testing")
	//ulogger.Debug("testing")
	//ulogger.Warning("testing")
	//WaitPush()
}
