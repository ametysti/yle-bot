package main

import (
	"encoding/json"
	"fmt"
	"mehf/yle-bot/db"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
)

func fetcher(dg *discordgo.Session) {
	for range time.Tick(time.Minute * 5) {
		data := GetYleNews()

		publishedDate := db.GetRecentID()

		heartbeatUrl := os.Getenv("HEARTBEAT_URL")

		if strings.HasPrefix(heartbeatUrl, "https://") {
			go http.Head(heartbeatUrl)
		}

		data.LiveBlogUpdate.ArticleBody = strings.ReplaceAll(data.LiveBlogUpdate.ArticleBody, "   ", "\n\n")
		data.LiveBlogUpdate.ArticleBody = strings.ReplaceAll(data.LiveBlogUpdate.ArticleBody, "  ", "\n\n")

		if publishedDate != data.LiveBlogUpdate.DatePublished {
			db.SaveID(data.LiveBlogUpdate.DatePublished)

			dg.ForumThreadStart("1224729567072227502", data.LiveBlogUpdate.Headline, 0, data.LiveBlogUpdate.ArticleBody)
		}
	}
}

func GetYleNews() BlogPosting {
	url := "https://yle.fi/a/74-20008814"

	// Create a new collector
	c := colly.NewCollector()

	var jsonData BlogPosting

	foundJSON := false

	c.OnHTML("div.post-content script[type='application/ld+json']", func(e *colly.HTMLElement) {
		if !foundJSON {
			if err := json.Unmarshal([]byte(e.Text), &jsonData); err != nil {
				fmt.Println("Failed to parse JSON data:", err)
			}
			foundJSON = true
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	if err := c.Visit(url); err != nil {
		fmt.Println("Failed to visit URL:", err)
	}

	fmt.Printf("JSON Data: %+v\n", jsonData)

	return jsonData
}
