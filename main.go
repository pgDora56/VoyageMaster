package main

import (
	"log"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token      string         `toml:"token"`
	Targets    []Target       `toml:"targets"`
	DeleteTime int64          `toml:"deletetime"`
	Template   NotifyTemplate `toml:"templates"`
}

type Target struct {
	Category    string `toml:"category"`
	TextChannel string `toml:"sendto"`
}

type NotifyTemplate struct {
	Join  string `toml:"join"`
	Move  string `toml:"move"`
	Leave string `toml:"leave"`
}

type ReserveDelete struct {
	ChannelId string
	MessageId string
	LimitUnix int64
}

var (
	cfg     Config
	waitDel []ReserveDelete
	stopper = make(chan bool)
)

func main() {
	cfg = getConfig()
	disc, err := discordgo.New()
	disc.Token = "Bot " + cfg.Token
	if err != nil {
		log.Fatal("Can't login", err)
	}

	disc.AddHandler(onVoiceStateUpdate)
	err = disc.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer disc.Close()

	power := make(chan bool)
	go deleteLine(power, disc)

	log.Println("Starting bot is successfully.")
	<-stopper
}

func getConfig() (cfg Config) {
	_, err := toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func deleteLine(power chan bool, s *discordgo.Session) {
	waitDel = make([]ReserveDelete, 0)
	for {
		select {
		case <-power:
			return
		default:
			if len(waitDel) != 0 {
				nowUnix := time.Now().Unix()
				if waitDel[0].LimitUnix < nowUnix {
					delMsg := waitDel[0]
					waitDel = waitDel[1:]
					err := s.ChannelMessageDelete(delMsg.ChannelId, delMsg.MessageId)
					if err != nil {
						log.Println("MessageDeleteError:", err)
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// Callback

func onVoiceStateUpdate(s *discordgo.Session, aft *discordgo.VoiceStateUpdate) {
	user := getUser(s, aft.UserID)
	bef := aft.BeforeUpdate
	validBef := bef != nil
	validAft := aft.ChannelID != ""

	if validBef {
		if validAft {
			if bef.ChannelID == aft.ChannelID {
				// Don't move
				return
			}

			// Move
			befch := getChannel(s, bef.ChannelID)
			aftch := getChannel(s, aft.ChannelID)
			log.Printf("%s: %s -> %s", user.Username, befch.Name, aftch.Name)
			for _, target := range cfg.Targets {
				if befch.ParentID == target.Category || aftch.ParentID == target.Category {
					// find target
					sendNotify(s, target.TextChannel, makeNotifyMessage(user.Username, befch.Name, aftch.Name))
					return
				}
			}
		} else {
			befch := getChannel(s, bef.ChannelID)
			for _, target := range cfg.Targets {
				if befch.ParentID == target.Category {
					// find target
					sendNotify(s, target.TextChannel, makeNotifyMessage(user.Username, befch.Name, ""))
					return
				}
			}
			log.Printf("%s: Leave from %s", user.Username, befch.Name)
		}
	} else if validAft {
		aftch := getChannel(s, aft.ChannelID)
		for _, target := range cfg.Targets {
			if aftch.ParentID == target.Category {
				// find target
				sendNotify(s, target.TextChannel, makeNotifyMessage(user.Username, "", aftch.Name))
				return
			}
		}
	}
}

// Some tools for discordgo

func getChannel(s *discordgo.Session, id string) *discordgo.Channel {
	st, err := s.Channel(id)
	if err != nil {
		log.Fatal("Cant get channel:", err)
	}
	return st
}

func getUser(s *discordgo.Session, id string) *discordgo.User {
	us, err := s.User(id)
	if err != nil {
		log.Fatal("Cant get user:", err)
	}
	return us
}

func sendNotify(s *discordgo.Session, channelID string, msg string) {
	log.Println("Send", channelID, msg)
	message, err := s.ChannelMessageSend(channelID, msg)

	waitDel = append(waitDel, ReserveDelete{
		ChannelId: channelID,
		MessageId: message.ID,
		LimitUnix: time.Now().Unix() + cfg.DeleteTime,
	})

	if err != nil {
		log.Printf("MessageSendError[ChannelID:%s] %v¥n", channelID, err)
	}
}

func makeNotifyMessage(user string, bef string, aft string) string {
	if bef == "" {
		if aft == "" {
			log.Fatal("Can't create message: both bef and aft is nothing")
		}
		// Join notify
		return strings.Replace(
			strings.Replace(
				cfg.Template.Join,
				"{user}",
				user,
				-1,
			),
			"{channel}",
			aft,
			-1,
		)
	}
	if aft == "" {
		// Leave notify
		return strings.Replace(
			strings.Replace(
				cfg.Template.Leave,
				"{user}",
				user,
				-1,
			),
			"{channel}",
			bef,
			-1,
		)
	}

	// Move Notify
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				cfg.Template.Move,
				"{user}",
				user,
				-1,
			),
			"{before}",
			bef,
			-1,
		),
		"{after}",
		aft,
		-1,
	)
}
