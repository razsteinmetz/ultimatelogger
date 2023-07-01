package ultimatelogger

// handle the telegram hook
// adpated from "github.com/Hu13er/telegrus"

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// telegram url
const sendMessageRequest = "https://api.telegram.org/bot%s/sendMessage?chat_id=%d&text=%s"
const CHANNEL_SIZE = 128

// holds telegram bot info
type telegramBot struct {
	botToken string
	chatID   int64
	queue    chan string // queue of messages to send
	cancel   chan bool   // cancel channel
}

// create a new telegram bot
func newTelegramBot(botToken string, chatID int64) *telegramBot {
	bot := &telegramBot{
		botToken: botToken,
		chatID:   chatID,
		queue:    make(chan string, CHANNEL_SIZE),
		cancel:   make(chan bool),
	}
	// the bot runs in the background all the time (until a cancel is sent or program exit)
	go bot.flush()
	return bot
}

func (tb *telegramBot) SendMsg(text string) {
	tb.queue <- text
}

func (tb *telegramBot) Cancel() {
	tb.cancel <- true
}

func (tb *telegramBot) flush() {
	for {
		select {
		case txt := <-tb.queue:
			query := fmt.Sprintf(sendMessageRequest,
				url.QueryEscape(tb.botToken),
				tb.chatID,
				url.QueryEscape(txt))

			resp, err := http.Get(query)
			if err != nil {
				log.Println("Error sending message:", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				log.Println("Response status code:", resp.Status)
				continue
			}
		case <-tb.cancel:
			return
		}
	}
}

var (
	TextFormatter = &logrus.TextFormatter{DisableColors: true}
	JSONFormatter = &logrus.JSONFormatter{}
)

type TelegramHook struct {
	bot       *telegramBot
	MinLevel  logrus.Level
	formatter logrus.Formatter
	mutex     sync.Mutex // to prevent concurrent access to the bot
}

func NewTelegramHook(botToken string, chatID int64, minLevel logrus.Level, formatter logrus.Formatter) logrus.Hook {
	return &TelegramHook{
		bot:       newTelegramBot(botToken, chatID),
		MinLevel:  minLevel,
		formatter: formatter,
	}
}

func (h *TelegramHook) Fire(entry *logrus.Entry) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	buf, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	h.bot.SendMsg(string(buf))
	if entry.Level <= logrus.FatalLevel {
		time.Sleep(1 * time.Second)
		h.bot.Cancel()
	}
	return nil
}

func (h *TelegramHook) Levels() []logrus.Level {
	return logrus.AllLevels[:h.MinLevel+1]
}
