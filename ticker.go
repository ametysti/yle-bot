package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mehf/yle-bot/db"
	"mehf/yle-bot/prom"
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
	for range time.Tick(time.Second * 60) {
		latency := dg.HeartbeatLatency()

		if latency.Milliseconds() > 0 {
			prom.BotLatency.Set(latency.Seconds())
		}

		data := GetYleNews()

		recentId := db.GetRecentID()

		heartbeatUrl := os.Getenv("HEARTBEAT_URL")

		if strings.HasPrefix(heartbeatUrl, "https://") {
			// internal monitoring ensuring data is received successfully
			go http.Head(heartbeatUrl)
		}

		threadConfig := &discordgo.ThreadStart{
			Name:                data.Title,
			Type:                5,
			AutoArchiveDuration: 0,
		}

		message := &discordgo.MessageSend{
			Content: data.Content,
		}

		if data.Image.URL != "" {
			imageData, _ := GetImageDataFromURL(data.Image.URL)

			reader := bytes.NewReader(imageData)

			file := &discordgo.File{
				Name:        "kuva.png",
				ContentType: "image/png",
				Reader:      reader,
			}

			message.Files = append(message.Files, file)
		}

		//dg.ForumThreadStartComplex("1224756452896542910", threadConfig, message)

		if recentId != data.Title {
			db.SaveID(data.Title)

			forumMessage, _ := dg.ForumThreadStartComplex(os.Getenv("NEWS_CHANNEL_ID"), threadConfig, message)

			if len(data.Image.Description) != 0 {
				dg.ChannelMessageSend(forumMessage.ID, "Tagit: "+data.Image.Keywords)
			}
		}
	}
}

func GetYleNews() BlogPosting {
	filterTags := md.Rule{
		Filter: []string{"img", "iframe", "picture", "a"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			yleDomain := "https://yle.fi"

			if selec.Is("a") {
				href, exists := selec.Attr("href")
				if exists && strings.HasPrefix(href, "/a/") {
					newHref := yleDomain + href
					selec.SetAttr("href", newHref)
				}
				return md.String("[" + selec.Text() + "](<" + selec.AttrOr("href", "") + ">)")
			}

			if selec.Is("img") || selec.Is("iframe") || selec.Is("picture") {
				return md.String("")
			}

			return md.String(selec.Text())
		},
	}
	converter := md.NewConverter("", true, nil)
	converter.AddRules(filterTags)

	url := "https://yle.fi/uutiset"

	c := colly.NewCollector()

	var jsonData BlogPosting
	var LDJsonData LDJson

	foundArticle := false

	c.OnHTML("article.yle__article", func(e *colly.HTMLElement) {
		if foundArticle {
			return
		}
		foundArticle = true

		e.DOM.Find("h1.yle__article__heading").Each(func(_ int, s *goquery.Selection) {
			jsonData.Title = s.Text()
		})

		e.DOM.Find("figure.yle__article__figure script[type='application/ld+json']").Each(func(_ int, s *goquery.Selection) {
			if err := json.Unmarshal([]byte(s.Text()), &LDJsonData); err != nil {
				fmt.Println("Failed to parse JSON data:", err)
			}
		})

		e.DOM.Find("section.yle__article__content").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()

			if err != nil {
				fmt.Println("Failed to get html content", err)
			}

			mdText, err := converter.ConvertString(html)

			if err != nil {
				fmt.Println("Failed to convert html to md", err)
			}

			jsonData.Content = mdText
		})
	})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			fmt.Printf("Response failed with %d", r.StatusCode)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	start := time.Now()
	err := c.Visit(url)
	respTime := time.Since(start).Seconds()

	if err != nil {
		fmt.Println("Failed to visit URL:", err)
	}

	prom.YleLatency.Observe(respTime)

	jsonData.Image = LDJsonData.Image

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
