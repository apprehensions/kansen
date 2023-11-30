package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

func main() {
	token := flag.String("token", "", "Discord Token")
	users := flag.String("users", "", "User ID(s) to favor; seperated by commas")
	flag.Parse()

	s := state.New(*token)
	s.AddIntents(gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentDirectMessages)

	if err := s.Open(context.TODO()); err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	me, err := s.Me()
	if err != nil {
		log.Fatal(err)
	}

	userids := func() (uids []discord.UserID) {
		suids := strings.Split(*users, ",")
		for _, suid := range suids {
			uid, err := strconv.Atoi(suid)
			if err != nil {
				log.Fatal(err)
			}
			uids = append(uids, discord.UserID(uid))
		}
		return
	}()

	s.AddHandler(func(m *gateway.MessageCreateEvent) {
		if m.Message.Author.ID == me.ID {
			return
		}

		if len(userids) > 0 {
			for _, uid := range userids {
				if m.Message.Author.ID != uid {
					return
				}
			}
		}

		log.Println("User", m.Message.Author.ID, "Requested image", m.Message.Content)

		f, err := os.Open(m.Message.Content)
		if err != nil {
			log.Println(err)
			_, err := s.SendMessageReply(m.Message.ChannelID, err.Error(), m.Message.ID)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		defer f.Close()

		_, err = s.SendMessageComplex(m.Message.ChannelID, api.SendMessageData{
			Files: []sendpart.File{{
				Name:   f.Name(),
				Reader: f,
			}},
		})
		if err != nil {
			log.Fatal(err)
		}
	})

	log.Println("Connected, waiting for users", userids, "to send file paths")

	select {}
}
