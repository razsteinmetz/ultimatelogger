package ultimatelogger

import (
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
)

// import (
//
//	"github.com/orandin/lumberjackrus"
//	"github.com/sirupsen/logrus"
//	prefixed "github.com/x-cray/logrus-prefixed-formatter"
//
// )
//
//	type LogConfig struct {
//		Telegram         bool             `toml:"telegram"`
//		Pushover         bool             `toml:"pushover"`
//		Console          bool             `toml:"console"`
//		Rotatefile       bool             `toml:"rotatefile"`
//		RotateFileConfig RotateFileConfig `toml:"RotateFileConfig"`
//		PushoverConfig   PushoverConfig   `toml:"PushoverConfig"`
//		TelegramConfig   TelegramConfig   `toml:"TelegramConfig"`
//		ConsoleConfig    ConsoleConfig    `toml:"ConsoleConfig"`
//		Text             Text             `toml:"text"` // this is the text formatter (same for all) - right now the only type we use
//	}
type RotateFileConfig struct {
	lumberjackrus.LogFile `mapstructure:",squash"` // squashes the struct into this one
	//Filename              string
	Level     string `toml:"Level"`
	Formatter string `toml:"Formatter"`
	formatter logrus.Formatter
}

//
//type PushoverConfig struct {
//	APIKey    string `toml:"APIKey"`
//	UserKey   string `toml:"UserKey"`
//	Level     string `toml:"Level"`
//	Formatter string `toml:"Formatter"`
//	FixedFont bool   `toml:"FixedFont"`
//	formatter logrus.Formatter
//}
//type TelegramConfig struct {
//	APIKey    string `toml:"APIKey"`
//	ChatID    int64  `toml:"ChatID"`
//	Level     string `toml:"Level"`
//	Formatter string `toml:"Formatter"`
//	formatter logrus.Formatter
//}
//type ConsoleConfig struct {
//	Level     string `toml:"Level"`
//	Formatter string `toml:"Formatter"`
//	formatter logrus.Formatter
//}
//
//type Text struct {
//	Config prefixed.TextFormatter
//}
