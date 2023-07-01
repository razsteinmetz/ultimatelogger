package ultimatelogger

import (
	"fmt"
	"github.com/gregdel/pushover"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"sync"
	"time"
)

// PushoverHook sends log via Pushover (https://pushover.net/)
// we can have a burst of up to MAX_BURST of messages sent without any inbetween delays
// there after we delay by MUTE_DELAY between messages.  Once MUTE_DELAY has been reached between messages a new Burst can happen
// we only use async mode to send message as not to hold the main thread

const (
	MAX_BURST  = 5
	MUTE_DELAY = 15 * time.Second
)

var pushMutex sync.Mutex
var wg sync.WaitGroup

type PushoverHook struct {
	muteDelay         time.Duration
	maxBurst          int
	msgCount          int
	lastMsgSentAt     time.Time
	minLevel          logrus.Level
	formatter         logrus.Formatter
	pushOverApp       *pushover.Pushover
	PushOverRecipient *pushover.Recipient
	FixedFont         bool
}

// NewPushoverHook init & returns a new PushoverHook with a defined max level and a formatter
func NewPushoverHook(pushoverUserToken, pushoverAPIToken string, maxLevel logrus.Level, formatter logrus.Formatter, fixedFont bool) (*PushoverHook, error) {
	var err error
	p := PushoverHook{
		muteDelay: MUTE_DELAY,
		maxBurst:  MAX_BURST,
		msgCount:  0,
		minLevel:  maxLevel,
		formatter: formatter,
		FixedFont: fixedFont,
	}
	if formatter == nil {
		p.formatter = &prefixed.TextFormatter{}
	}
	p.pushOverApp = pushover.New(pushoverAPIToken)
	p.PushOverRecipient = pushover.NewRecipient(pushoverUserToken)
	return &p, err
}

// Levels returns the available logging levels.
func (hook *PushoverHook) Levels() []logrus.Level {
	return logrus.AllLevels[:hook.minLevel+1]
}

// Fire is called when a log event is fired.
func (hook *PushoverHook) Fire(entry *logrus.Entry) error {
	msg, err := hook.formatter.Format(entry)
	if err != nil {
		logrus.Error("Cant format message to pushover")
		return err
	}
	go hook.SendMessage(msg, entry)
	if entry.Level <= logrus.FatalLevel {
		time.Sleep(1 * time.Second)
		WaitPush()
	}
	return nil
}

// SendMessage call this as a goroutine to send a message
// note that it will sleep if burst is exceeded.
// todo: make sure that only to allow another burst after certain time period or quite
func (hook *PushoverHook) SendMessage(msg []byte, entry *logrus.Entry) {
	wg.Add(1)
	pushMutex.Lock()
	defer pushMutex.Unlock()
	defer wg.Done()
	if time.Since(hook.lastMsgSentAt) >= hook.muteDelay {
		hook.msgCount = 0
	} else {
		hook.msgCount++
	}
	pmsg := &pushover.Message{
		Message: string(msg), // "<b><h1>Message Body</h1></b> <br> <i>Message Body in italic</i>",
		//Title:       "Message Title",
		Priority:    logToPush(entry.Level),
		Timestamp:   entry.Time.Unix(),
		Expire:      time.Hour,
		CallbackURL: "",
		Retry:       300 * time.Second,
		//DeviceName:  hook.deviceName,
		//Sound:       pushover.SoundBike,
		HTML:      !hook.FixedFont,
		Monospace: hook.FixedFont,
	}

	_, err := hook.pushOverApp.SendMessage(pmsg, hook.PushOverRecipient)
	if err != nil {
		fmt.Println("Error sending to pushover :", err.Error())
		time.Sleep(100 * time.Millisecond)
	}
	hook.lastMsgSentAt = time.Now()
	if hook.msgCount >= hook.maxBurst {
		//fmt.Println("sleeping mute after msgcount is ", hook.msgCount)
		time.Sleep(hook.muteDelay)
		hook.msgCount = 0
	}
}

// convert logrus level to pushover levels
func logToPush(level logrus.Level) (p int) {
	switch level {
	case logrus.PanicLevel:
		return 2
	case logrus.FatalLevel:
		return 1
	case logrus.ErrorLevel:
		return 0
	case logrus.WarnLevel:
		return -1
	case logrus.InfoLevel:
		return -2
	default:
		return -2
	}
}

// WaitPush - wait for all messages to complete - use before program end!
func WaitPush() {
	wg.Wait()
}
