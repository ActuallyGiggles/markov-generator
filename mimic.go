package main

import (
	"fmt"
	"strings"
	"time"
)

func Mimic(C chan Message) {
messageRange:
	for c := range C {
		u := c.Channel
		m := c.Message

		if updatingEmotes {
			continue
		}

		for i := 0; i < len(Users); i++ {
			user := &Users[i]
			if user.Name == u {

				if user.Busy || !user.IsLive {
					continue messageRange
				}

				e := ParseEmote(m)

				exists := false
				for i := 0; i < len(user.Emotes); i++ {
					emote := &user.Emotes[i]
					if e == emote.Name {
						exists = true
						emote.Value++

						printDetectedEmotes(user)
					}

					if emote.Value >= Config.MessageThreshold {
						go Respond(user, emote.Name)
						user.Messages = 0
						user.Emotes = nil
						continue messageRange
					}
				}

				if !exists && e != "" {
					entry := Emote{
						Name:  e,
						Value: 1,
					}
					user.Emotes = append(user.Emotes, entry)

					printDetectedEmotes(user)
				}

				user.Messages++

				if user.Messages > Config.MessageSample {
					user.Messages = 0
					user.Emotes = nil
				}
			}
		}
	}
}

func printDetectedEmotes(user *User) {
	fmt.Printf("Emotes detected in %s:\n", user.Name)
	for _, emoticon := range user.Emotes {
		fmt.Printf("\t%s: %d/%d\n", emoticon.Name, emoticon.Value, Config.MessageThreshold)
	}
}

func ParseEmote(message string) (eJ string) {
	s := strings.Split(message, " ")

	var eS []string
loop1:
	for _, w := range s {
		for _, emote := range ChannelEmotes {
			if w == emote {
				for _, blacked := range Config.BlacklistEmotes {
					if strings.ToLower(blacked) == strings.ToLower(emote) {
						return ""
					}
				}
				eS = append(eS, emote)
				continue loop1
			}
		}
		for _, emote := range GlobalEmotes {
			if w == emote {
				for _, blacked := range Config.BlacklistEmotes {
					if strings.ToLower(blacked) == strings.ToLower(emote) {
						return ""
					}
				}
				eS = append(eS, emote)
				continue loop1
			}
		}
	}

	eJ = strings.Join(eS, " ")

	return eJ
}

func Respond(u *User, message string) {
	u.Busy = true
	rS := RandomNumber(2, 10)
	fmt.Printf("Saying %s in %s's chat in %d seconds.\n", message, u.Name, rS)
	time.Sleep(time.Duration(rS) * time.Second)
	Say(u.Name, message)
	clearTerminal()
	fmt.Printf("Said %s in %s's chat.\n", message, u.Name)

	if Config.IntervalMin == Config.IntervalMax {
		fmt.Println("Waiting", Config.IntervalMin, "minutes to start detecting again...")
		time.Sleep(time.Duration(Config.IntervalMin) * time.Minute)
		u.Busy = false
	} else {
		r := RandomNumber(Config.IntervalMin, Config.IntervalMax)
		fmt.Println("Waiting", r, "minutes to start detecting again...")
		time.Sleep(time.Duration(r) * time.Minute)
		u.Busy = false
	}
}
