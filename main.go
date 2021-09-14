package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

var (
	Token   = "Bot TOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKEN"
	stopper = make(chan bool)
)

func main() {
	disc, err := discordgo.New()
	disc.Token = Token
	if err != nil {
		log.Fatal("Can't login", err)
	}

	disc.AddHandler(onVoiceStateUpdate)
	err = disc.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer disc.Close()

	log.Println("Starting bot is successfully.")
	<-stopper
	return
}

func onVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	bef := v.BeforeUpdate
	if bef != nil {
		log.Printf("Before: %v %v %v %v", bef.UserID, bef.SessionID, getChannelName(s, bef.ChannelID), bef.GuildID)
	}
	log.Printf("After : %v %v %v %v", v.UserID, v.SessionID, getChannelName(s, v.ChannelID), v.GuildID)
}

func getChannelName(s *discordgo.Session, id string) string {
	st, err := s.Channel(id)
	if err != nil {
		log.Fatal("Cant get channel name:", err)
	}
	return st.Name
}
