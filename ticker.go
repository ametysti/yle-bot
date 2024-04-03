package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mehf/yle-bot/db"
	"net/http"
	"os"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
)

func fetcher(dg *discordgo.Session) {

	for range time.Tick(time.Second * 10) {
		data := GetYleNews()

		publishedDate := db.GetRecentID()

		heartbeatUrl := os.Getenv("HEARTBEAT_URL")

		if strings.HasPrefix(heartbeatUrl, "https://") {
			go http.Head(heartbeatUrl)
		}

		threadConfig := &discordgo.ThreadStart{
			Name:                data.LiveBlogUpdate.Headline,
			Type:                5,
			AutoArchiveDuration: 0,
		}

		message := &discordgo.MessageSend{
			Content: data.Content,
		}

		if data.LiveBlogUpdate.Image.URL != "" {
			imageData, _ := GetImageDataFromURL(data.LiveBlogUpdate.Image.URL)

			reader := bytes.NewReader(imageData)

			file := &discordgo.File{
				Name:        "kuva.png",
				ContentType: "image/png",
				Reader:      reader,
			}

			message.Files = append(message.Files, file)
		}

		//dg.ForumThreadStartComplex("1224756452896542910", threadConfig, message)

		if publishedDate != data.LiveBlogUpdate.DatePublished {
			db.SaveID(data.LiveBlogUpdate.DatePublished)

			dg.ForumThreadStartComplex("1224729567072227502", threadConfig, message)
		}
	}
}

func GetYleNews() BlogPosting {
	filterTags := md.Rule{
		Filter: []string{"img", "iframe", "picture"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			return md.String("")
		},
	}
	converter := md.NewConverter("", true, nil)
	converter.AddRules(filterTags)

	url := "https://yle.fi/a/74-20008814"

	// Create a new collector
	c := colly.NewCollector()

	var jsonData BlogPosting

	foundJSON := false
	foundHTML := false

	c.OnHTML("div.post-content script[type='application/ld+json']", func(e *colly.HTMLElement) {
		if !foundJSON {
			if err := json.Unmarshal([]byte(e.Text), &jsonData); err != nil {
				fmt.Println("Failed to parse JSON data:", err)
			}

			foundJSON = true
		}
	})

	c.OnHTML("div.post-content", func(e *colly.HTMLElement) {
		if !foundHTML {
			html, err := e.DOM.Html()

			if err != nil {
				fmt.Println("Failed to get html content", err)
			}

			mdText, err := converter.ConvertString(html)

			// remove title from actual content because its already in the title lmao
			// Not too sure if the Yle "short news" has sections that may use h2,
			// so as a precaution just delete the section ONLY if the article name matches
			jsonData.LiveBlogUpdate.Headline = strings.TrimSpace(jsonData.LiveBlogUpdate.Headline)
			mdText = strings.ReplaceAll(mdText, fmt.Sprintf("## %s", jsonData.LiveBlogUpdate.Headline), "")

			if err != nil {
				fmt.Println("Failed to convert html to md", err)
			}

			jsonData.Content = mdText

			foundHTML = true
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

func GetImageDataFromURL(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching the image: %v", err)
	}
	defer response.Body.Close()

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading image data: %v", err)
	}

	return imageData, nil
}
