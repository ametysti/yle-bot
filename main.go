package main

import (
	"fmt"
	"mehf/yle-bot/db"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		Image         struct {
			URL string `json:"url"`
		} `json:"image"`
	} `json:"liveBlogUpdate"`
	Content string
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

	if m.Content == "!yle-data" {
		data := GetYleNews()

		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       data.LiveBlogUpdate.Headline,
			Description: data.LiveBlogUpdate.ArticleBody,
			URL:         "https://yle.fi/a/74-20008814",
		})
	}

	if m.Content == "!status" {
		start := time.Now()
		resp, _ := http.Get("https://yle.fi/a/74-20008814")
		respTime := time.Since(start).Milliseconds()

		embed := &discordgo.MessageEmbed{}

		embed.URL = "https://yle.fi/a/74-20008814"

		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%dms", respTime),
		}

		if resp.StatusCode != 200 {
			embed.Title = "❌ Ei saada yhteyttä"
			embed.Description = "Emme nähtävästi saaneet yhteyttä Ylen sivuille."
			embed.Color = 15548997
		}

		if resp.StatusCode == 200 && respTime > 1000 {
			embed.Title = "ℹ️ Yhteys saatu"
			embed.Description = "Saimme yhteyden Ylen sivuille mutta yhdistäminen oli hidasta."
			embed.Color = 16776960
		}

		if resp.StatusCode == 200 && respTime < 1000 {
			embed.Title = "✅ Yhteys saatu"
			embed.Description = "Saimme yhteyden Ylen sivuille"
			embed.Color = 5763719
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}
}
