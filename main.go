package main

import (
	"fmt"
	"mehf/yle-bot/db"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type BlogPosting struct {
	ID             string `json:"@id"`
	LiveBlogUpdate struct {
		Type          string `json:"@type"`
		DatePublished string `json:"datePublished"`
		ArticleBody   string `json:"articleBody"`
		URL           string `json:"url"`
		Author        string `json:"author"`
		Headline      string `json:"headline"`
	} `json:"liveBlogUpdate"`
}

func main() {
	godotenv.Load(".env")
	go db.Connect()

	dg, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	go fetcher(dg)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!data" {
		data := GetYleNews()

		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       data.LiveBlogUpdate.Headline,
			Description: data.LiveBlogUpdate.ArticleBody,
			URL:         "https://yle.fi/a/74-20008814",
		})
	}
}
