package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Name       string   `toml:"name"`
	Token      string   `toml:"token"`
	Targets    []Target `toml:"targets"`
	DeleteTime int64    `toml:"deletetime"`
	Join       string   `toml:"join"`
	Move       string   `toml:"move"`
	Leave      string   `toml:"leave"`
}

type Target struct {
	Category    string `toml:"category"`
	TextChannel string `toml:"sendto"`
}

type ReserveDelete struct {
	Session   *discordgo.Session
	ChannelId string
	MessageId string
	LimitUnix int64
}

var (
	cfgDic  map[string]Config
	waitDel []ReserveDelete
	stopper = make(chan bool)
	power   = make(chan bool)
)

func main() {
	cfgDic = map[string]Config{}
	logfile, err := os.OpenFile("./voyagemaster.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Can't open logfile:", err.Error())
	}
	defer logfile.Close()

	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	log.Printf("============> VoyageMaster sail out!: %v\n", os.Args)

	if len(os.Args) > 1 {
		if os.Args[1] == "setting" {
			// setting tool
			setting()
		} else {
			log.Printf("Unknown arguments: %v\n", os.Args[1:])
		}
		return
	}
	cfgs, err := getConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, cfg := range cfgs {
		go cfg.watch()
		cfgDic["Bot "+cfg.Token] = cfg
	}
	go deleteLine(power)

	<-stopper
}

// Interactive setting tools
func setting() {
	log.Println("Move to setting mode")
	scanner := bufio.NewScanner(os.Stdin)

	var cfgs []Config
	cfgs, err := getConfig()
	if err == nil {
		fmt.Println("The setting file is already there. Do you want to remake it? If not, it will be added to the existing settings.")
		yorn := ""
		for yorn != "yes" && yorn != "no" {
			fmt.Print("yes(new settings)/no(add to the existing settings):")
			scanner.Scan()
			yorn = scanner.Text()
		}
		if yorn == "yes" {
			cfgs = make([]Config, 0)
		}
	}

	cont := "yes"
	for cont == "yes" {
		var cfg Config

		// Bot-name
		for cfg.Name == "" {
			fmt.Print("Bot name:")
			scanner.Scan()
			cfg.Name = scanner.Text()
		}

		// Token
		for cfg.Token == "" {
			fmt.Print("Discord Bot Token:")
			scanner.Scan()
			cfg.Token = scanner.Text()
		}

		// Targets
		cont = "yes"
		for cont == "yes" {
			var target Target
			for target.Category == "" {
				fmt.Print("Target category id(contain voice channels to be monitored):")
				scanner.Scan()
				target.Category = scanner.Text()
			}
			for target.TextChannel == "" {
				fmt.Print("Text channel id(to send notifications to):")
				scanner.Scan()
				target.TextChannel = scanner.Text()
			}
			cfg.Targets = append(cfg.Targets, target)

			cont = ""
			fmt.Println("Do you want to add other targets? Please enter `yes` or `no`.")
			for cont != "yes" && cont != "no" {
				fmt.Print("yes/no:")
				scanner.Scan()
				cont = scanner.Text()
			}
		}

		// Delete time
		for cfg.DeleteTime <= 0 {
			fmt.Print("Time to delete notification(natural number only(>0, integer)):")
			scanner.Scan()
			dt, err := strconv.Atoi(scanner.Text())
			if err == nil {
				cfg.DeleteTime = int64(dt)
			}
		}

		// Join Message
		fmt.Println("Please enter Join notification message template(`{user}`->username, `{channel}`->channel to join)")
		for cfg.Join == "" {
			fmt.Print("Join:")
			scanner.Scan()
			cfg.Join = scanner.Text()
		}

		// Move Message
		fmt.Println("Please enter Move notification message template(`{user}`->username, `{before}`->channel to leave, `{after}`->channel to join)")
		for cfg.Move == "" {
			fmt.Print("Move:")
			scanner.Scan()
			cfg.Move = scanner.Text()
		}

		// Leave Message
		fmt.Println("Please enter Leave notification message template(`{user}`->username, `{channel}`->channel to leave)")
		for cfg.Leave == "" {
			fmt.Print("Leave:")
			scanner.Scan()
			cfg.Leave = scanner.Text()
		}

		cfgs = append(cfgs, cfg)
		cont = ""
		fmt.Println("Do you want to add other setting? Please enter `yes` or `no`.")
		for cont != "yes" && cont != "no" {
			fmt.Print("yes/no:")
			scanner.Scan()
			cont = scanner.Text()
		}
	}

	// Write file
	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(struct {
		Bot []Config `toml:"bot"`
	}{
		Bot: cfgs,
	})

	if err != nil {
		log.Fatal("Toml encode error:", err)
	}
	err = ioutil.WriteFile("config.toml", buf.Bytes(), 0777)
	if err != nil {
		log.Fatal("File write error:", err)
	}
	log.Println("Create setting file: config.toml")
}

// Goroutines

func (cfg Config) watch() {
	disc, err := discordgo.New()
	disc.Token = "Bot " + cfg.Token
	if err != nil {
		log.Fatalf("[%s]Can't login: %s\n", cfg.Name, err.Error())
	}

	disc.AddHandler(onVoiceStateUpdate)
	err = disc.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer disc.Close()

	log.Printf("[%s]Starting bot is successfully.", cfg.Name)
	<-power
}

func deleteLine(power chan bool) {
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
					sess := delMsg.Session
					err := sess.ChannelMessageDelete(delMsg.ChannelId, delMsg.MessageId)
					if err != nil {
						log.Printf("[%s]MessageDeleteError: %s\n", cfgDic[sess.Identify.Token].Name, err.Error())
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// Callback

func onVoiceStateUpdate(s *discordgo.Session, aft *discordgo.VoiceStateUpdate) {
	token := s.Identify.Token
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
			log.Printf("[%s]%s: %s -> %s", cfgDic[token].Name, user.Username, befch.Name, aftch.Name)
			for _, target := range cfgDic[token].Targets {
				if befch.ParentID == target.Category || aftch.ParentID == target.Category {
					// find target
					sendNotify(s, target.TextChannel, makeNotifyMessage(token, user.Username, befch.Name, aftch.Name))
					return
				}
			}
		} else {
			befch := getChannel(s, bef.ChannelID)
			for _, target := range cfgDic[token].Targets {
				if befch.ParentID == target.Category {
					// find target
					sendNotify(s, target.TextChannel, makeNotifyMessage(token, user.Username, befch.Name, ""))
					return
				}
			}
		}
	} else if validAft {
		aftch := getChannel(s, aft.ChannelID)
		for _, target := range cfgDic[token].Targets {
			if aftch.ParentID == target.Category {
				// find target
				sendNotify(s, target.TextChannel, makeNotifyMessage(token, user.Username, "", aftch.Name))
				return
			}
		}
	}
}

// Some tools for discordgo

func getChannel(s *discordgo.Session, id string) *discordgo.Channel {
	st, err := s.Channel(id)
	if err != nil {
		log.Fatalf("[%s]Can't get channel: %s\n", cfgDic[s.Identify.Token].Name, err.Error())
	}
	return st
}

func getUser(s *discordgo.Session, id string) *discordgo.User {
	us, err := s.User(id)
	if err != nil {
		log.Fatalf("[%s]Can't get user: %s\n", cfgDic[s.Identify.Token].Name, err.Error())
	}
	return us
}

func sendNotify(s *discordgo.Session, channelID string, msg string) {
	token := s.Identify.Token
	log.Printf("[%s]Send %s %s", cfgDic[token].Name, channelID, msg)
	message, err := s.ChannelMessageSend(channelID, msg)

	waitDel = append(waitDel, ReserveDelete{
		Session:   s,
		ChannelId: channelID,
		MessageId: message.ID,
		LimitUnix: time.Now().Unix() + cfgDic[token].DeleteTime,
	})

	if err != nil {
		log.Printf("[%s]MessageSendError[ChannelID:%s] %v¥n", cfgDic[token].Name, channelID, err)
	}
}

func makeNotifyMessage(token string, user string, bef string, aft string) string {
	if bef == "" {
		if aft == "" {
			log.Fatalf("[%s]Can't create message: both bef and aft is nothing\n", cfgDic[token].Name)
		}
		// Join notify
		return strings.Replace(
			strings.Replace(
				cfgDic[token].Join,
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
				cfgDic[token].Leave,
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
				cfgDic[token].Move,
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

func getConfig() ([]Config, error) {
	var cfgs struct {
		Bot []Config `toml:"bot"`
	}
	_, err := toml.DecodeFile("config.toml", &cfgs)
	if err != nil {
		return nil, err
	}
	return cfgs.Bot, nil
}
