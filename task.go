package main

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/krol44/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

type Task struct {
	App           *App
	Message       *tgbotapi.Message
	Files         []string
	MessageEditID int
	UserFromDB    User
	Torrent       struct {
		Name            string
		Process         *torrent.Torrent
		Progress        int64
		Uploaded        int64
		TorrentProgress int64
		TorrentUploaded int64
	}
}

func (t Task) Send(ct tgbotapi.Chattable) tgbotapi.Message {
	mess, err := t.App.Bot.Send(ct)
	if err != nil {
		log.Traceln(err)
	}

	return mess
}

func (t *Task) Alloc(typeDl string) {
	// creating edit message
	messStat := t.Send(tgbotapi.NewMessage(t.Message.Chat.ID, "🍀 Download is starting soon..."))
	t.MessageEditID = messStat.MessageID

	for {
		// global queue
		qn, _ := t.App.ChatsWork.m.Load(t.Message.MessageID)
		if qn.(int) < config.MaxTasks {
			break
		}

		t.Send(tgbotapi.NewEditMessageText(t.Message.Chat.ID, t.MessageEditID,
			fmt.Sprintf("🍀 Download is starting soon...\n\n🚦 Your queue: %d", qn.(int)-config.MaxTasks+1)))

		time.Sleep(4 * time.Second)
	}

	if t.UserFromDB.Premium == 0 {
		t.Send(tgbotapi.NewSticker(t.Message.Chat.ID,
			tgbotapi.FileID("CAACAgIAAxkBAAIEW2OcfHb7yPa6z59rHlFiTTUTkA3XAAJ-GQACHiDBS43V6msCr8MXKwQ")))

		messPremium := tgbotapi.NewMessage(t.Message.Chat.ID,
			`‼️ You don't have a donation for us, only the first 5 minutes video is available and torrent in the zip archive don't available too

		<a href="url-donate">Help us, subscribe and service will be more fantastical</a> 🔥

		(Write your telegram username in the body message. After donation, you will access 30 days)`)
		messPremium.ParseMode = tgbotapi.ModeHTML
		t.Send(messPremium)
	}

	// log
	t.App.SendLogToChannel(t.Message.From.ID, "mess", fmt.Sprintf("start download "+typeDl))
}

func (t Task) Cleaner() {
	t.App.LockForRemove.Add(1)

	log.Info("Folders cleaning...")

	pathConvert := config.DirBot + "/storage"
	dirs, _ := os.ReadDir(pathConvert)

	for _, val := range dirs {
		err := os.RemoveAll(pathConvert + "/" + val.Name())
		if err != nil {
			log.Error(err)
		}
	}

	if config.IsDev == false {
		for _, pathWay := range t.Files {
			pathDir := path.Dir(pathWay)
			pathRemove := config.DirBot + "/torrent-client/" + path.Dir(pathWay)

			// todo change or remove
			if pathDir == "." {
				pathRemove = config.DirBot + "/torrent-client/" + pathWay
			}

			_, err := os.Stat(pathRemove)
			if err != nil {
				continue
			}

			err = os.RemoveAll(pathRemove)
			if err != nil {
				log.Error(err)
			}
		}
	}

	t.App.LockForRemove.Done()
}

func (Task) IsAllowFormatForConvert(pathWay string) bool {
	for _, ext := range config.AllowVideoFormats {
		if ext == path.Ext(pathWay) {
			return true
		}
	}

	return false
}

func (Task) UniqueId(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	use := now.UnixNano() % 0x100000
	return fmt.Sprintf("%s-%08x%05x", prefix, sec, use)
}
