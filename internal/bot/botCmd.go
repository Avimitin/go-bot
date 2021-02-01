package bot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Avimitin/go-bot/internal/net"
	bapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var botCMD = command{
	"start":     start,
	"ping":      ping,
	"dump":      dump,
	"kick":      kick,
	"shutup":    shutUp,
	"disshutup": disShutUp,
}

func cmdArgv(msg *bapi.Message) []string {
	args := strings.Fields(msg.Text)
	if len(args) > 1 {
		args = args[1:]
		return args
	}
	return nil
}

func start(m *bapi.Message) error {
	username := m.From.UserName
	if username == "" {
		username = m.From.FirstName
	}
	userID := m.From.ID
	userLink := fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, userID, username)
	_, err := sendP("Hi, "+userLink, m.Chat.ID, "HTML")
	if err != nil {
		return errF("start", err, "send fail")
	}
	return nil
}

func ping(m *bapi.Message) error {
	now := time.Now()
	msg, err := sendT("pong!", m.Chat.ID)
	if err != nil {
		return errF("ping", err, "send fail")
	}
	current := time.Now()
	delay := current.Sub(now).Milliseconds()
	text := fmt.Sprintf("bot 与 Telegram 服务器的延迟大约为 %d 毫秒", delay)
	_, err = editT(text, m.Chat.ID, msg.MessageID)
	if err != nil {
		return errF("ping", err, "edit fail")
	}
	return nil
}

func dump(m *bapi.Message) error {
	var text = "<b>Message Information</b>\n" +
		"=== <b>CHAT</b> ===\n" +
		"<b>ID:</b> <code>%v</code>\n" +
		"<b>TYPE:</b> <code>%v</code>\n" +
		"<b>USERNAME:</b> <code>%v</code>\n" +
		"=== <b>USER</b> ===\n" +
		"<b>ID:</b> <code>%v</code>\n" +
		"<b>USERNAME:</b> <code>%v</code>\n" +
		"<b>NICKNAME:</b> <code>%v %v</code>\n" +
		"<b>LANGUAGE:</b> <code>%v</code>\n" +
		"=== <b>MSG</b> ===\n" +
		"<b>ID:</b> <code>%v</code>\n" +
		"<b>TEXT:</b> %v"

	if reply := m.ReplyToMessage; reply != nil {
		text = fmt.Sprintf(text,
			reply.Chat.ID, reply.Chat.Type, reply.Chat.UserName,
			reply.From.ID, reply.From.UserName, reply.From.FirstName, reply.From.LastName, reply.From.LanguageCode,
			reply.MessageID, reply.Text)
	} else {
		text = fmt.Sprintf(text,
			m.Chat.ID, m.Chat.Type, m.Chat.UserName,
			m.From.ID, m.From.UserName, m.From.FirstName, m.From.LastName, m.From.LanguageCode,
			m.MessageID, m.Text)
	}

	_, err := sendP(text, m.Chat.ID, "HTML")
	if err != nil {
		return errF("dump", err, "send fail")
	}
	return nil
}

func kick(m *bapi.Message) error {
	is, err := isAdmin(m.From.ID, m.Chat)
	if err != nil {
		errMsg := "fail to get user permission"
		if _, err = sendT(errMsg, m.Chat.ID); err != nil {
			return errF("kick", err, "fail to send error notify")
		}
		return errF("kick", err, errMsg)
	}
	// if command caller are not admin
	if !is {
		_, err = sendT("YOU ARE NOT ADMIN! DONT TOUCH THIS COMMAND!", m.Chat.ID)
		if err != nil {
			return errF("kick", err, "fail to send kick alert")
		}
		return nil
	}

	if m.ReplyToMessage == nil {
		_, err = sendT("You should reply to a user to kick him.", m.Chat.ID)
		if err != nil {
			return errF("kick", err, "fail to send usage")
		}
		return nil
	}

	userToKick := m.ReplyToMessage.From.ID
	err = kickUser(userToKick, m.Chat.ID, time.Now().Unix()+1)
	if err != nil {
		return errF("kick", err, "fail to kick user")
	}

	_, err = sendT("user has been kick forever", m.Chat.ID)
	if err != nil {
		return errF("kick", err, "fail to send usage")
	}
	return nil
}

func punishNoPermissionUser(m *bapi.Message) error {
	var err error
	respMsg, serr := sendT("generating....", m.Chat.ID)
	if serr != nil {
		return errF("shutUp", err, "fail to send generating msg")
	}

	var minLimit, maxLimit int64 = 60, 300
	rand.Seed(time.Now().Unix())
	randTime := rand.Int63n(maxLimit-minLimit) + minLimit
	err = editUserPermissions(m.From.ID, m.Chat.ID, time.Now().Unix()+randTime, false)
	if err != nil {
		if _, err = sendT("fail to limit user:"+err.Error(), m.Chat.ID); err != nil {
			return errF("shutUp", err, "fail to send error message")
		}
		return errF("shutUp", err, "fail to limit user")
	}

	respMsg, err = editT(
		fmt.Sprintf("Boom, you get a %d mins ban", randTime), m.Chat.ID, respMsg.MessageID)
	if err != nil {
		return errF("shutUp", err, "fail to edit message")
	}
	return nil
}

func shutUp(m *bapi.Message) error {
	is, err := isAdmin(m.From.ID, m.Chat)
	if err != nil {
		if _, err = sendT("fail to fetch admins, please try again later",
			m.Chat.ID); err != nil {
			return errF("shutUp", err, "fail to send error notify")
		}
		return errF("shutUp", err, "fail to fetch admins")
	}
	// if user is not admin
	if !is {
		return punishNoPermissionUser(m)
	}

	if m.ReplyToMessage == nil {
		if _, err = sendT("reply to a user to use this command", m.Chat.ID); err != nil {
			return errF("shutUp", err, "fail to send notify")
		}
	}

	err = editUserPermissions(m.From.ID, m.Chat.ID, time.Now().Unix()+1, false)
	if err != nil {
		if _, err = sendT("fail to limit user: "+err.Error(), m.Chat.ID); err != nil {
			return errF("shutUp", err, "fail to send error message")
		}
		return errF("shutUp", err, "fail to limit user")
	}
	if _, err := sendT("user has been forever muted", m.Chat.ID); err != nil {
		return errF("shutUp", err, "fail to send successful ban message")
	}
	return nil
}

func disShutUp(m *bapi.Message) error {
	is, err := isAdmin(m.From.ID, m.Chat)
	if err != nil {
		_, err = sendT("fail to fetch admin list", m.Chat.ID)
		if err != nil {
			return errF("disShutUp", err, "")
		}
		return err
	}
	if !is {
		return punishNoPermissionUser(m)
	}

	if m.ReplyToMessage == nil {
		_, err = sendT("Reply to a user to recover his permission", m.Chat.ID)
		if err != nil {
			return errF("disShutUp", err, "")

		}
		return nil
	}
	err = editUserPermissions(m.From.ID, m.Chat.ID, 0, true)
	if err != nil {
		sendT("recover user priviledge: "+err.Error(), m.Chat.ID)
		return errF("disShutUp", err, "")
	}
	_, err = sendT("User has recovered", m.Chat.ID)
	if err != nil {
		return errF("disShutUp", err, "")
	}
	return nil
}

func weather(m *bapi.Message) error {
	argv := cmdArgv(m)
	if argv == nil {
		_, err := sendT("Gib me a city name", m.Chat.ID)
		if err != nil {
			return errF("weather", err, "")
		}
		return nil
	}

	respMsg, err := sendT("requesting API server...", m.Chat.ID)
	if err != nil {
		return errF("weather", err, "")
	}

	city := argv[0]
	text, err := getWeatherContext(city)
	if err != nil {
		_, err = editT("fetch weather failed: "+err.Error(), m.Chat.ID, respMsg.MessageID)
		if err != nil {
			return errF("weather", err, "edit failed")
		}
	}
	_, err = editP(text, m.Chat.ID, respMsg.MessageID, "HTML")
	if err != nil {
		return errF("weather", err, "edit failed")
	}
	return nil
}

func getWeatherContext(city string) (string, error) {
	url := fmt.Sprintf("https://wttr.in/%s?format=%l的天气:+%c+温度:%t+湿度:%h+降雨量:%p", city)
	resp, err := net.Get(url)
	if err != nil {
		return "", errF("getWeatherContext", err, "get city "+city)
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, fmt.Sprintf("https://wttr.in/%s.png", city), resp), nil
}
