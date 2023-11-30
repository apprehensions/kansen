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

type UserIDs []discord.UserID

func (uids *UserIDs) String() string {
	var suids []string
	for _, uid := range *uids {
		suids = append(suids, uid.String())
	}
	return strings.Join(suids, ",")
}

func (uids *UserIDs) Set(value string) error {
	id, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	*uids = append(*uids, discord.UserID(id))
	return nil
}

func main() {
	var (
		uids  UserIDs
		token string
	)
	flag.StringVar(&token, "token", "", "Discord Token")
	flag.Var(&uids, "user", "User ID(s) to favor")
	flag.Parse()

	s := state.New(token)
	s.AddIntents(gateway.IntentGuilds | gateway.IntentGuildMessages | gateway.IntentDirectMessages)

	if err := s.Open(context.TODO()); err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	me, err := s.Me()
	if err != nil {
		log.Fatal(err)
	}

	s.AddHandler(func(m *gateway.MessageCreateEvent) {
		if m.Message.Author.ID == me.ID {
			return
		}

		if len(uids) > 0 {
			for _, uid := range uids {
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

	log.Println("Connected, waiting for users", uids, "to send file paths")

	select {}
}
