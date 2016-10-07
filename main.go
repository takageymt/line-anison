package main

import (       
  "fmt"
  "os"
  "strconv"
  "net/http"
  "github.com/line/line-bot-sdk-go/linebot"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var bot *linebot.Client
var db *sql.DB

func main() {
  var err error
  
  configProxy()

  bot, err = lineClient()
  if err != nil {
    fmt.Println(err)
  }

	err = connectDB()
	if err != nil {
		fmt.Println(err)
	} else {
		defer db.Close()
	}

  listen()
}

func listen() {
  port := os.Getenv("PORT")
  http.HandleFunc("/callback", handleRequest)
  http.ListenAndServe(":"+port, nil)    
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
  fmt.Println(req.Body)

  received, err := bot.ParseRequest(req)
  if err != nil {
		fmt.Println(err)
		return
  }
	
  for _, result := range received.Results {
    content := result.Content()
    if isTextContent(content) {
      text, err := content.TextContent()
      if err != nil {
        fmt.Println(err)
				continue		
      }
      fmt.Println(text.Text)

			message, err := searchQuery(text.Text)
			if err != nil {
				fmt.Println(err)
				continue
			}

      err = reply(message, content.From);			
      if err != nil {
        fmt.Println(err)
      }	
  
    }
  }
}

func isTextContent(content *linebot.ReceivedContent) bool {
  return content != nil && content.IsMessage && content.ContentType == linebot.ContentTypeText
}

func reply(text string, userID string) error {
	fmt.Println(text)
  _, err := bot.SendText([]string{userID}, text)
  return err
}

func configProxy() {
  fixieURL := os.Getenv("FIXIE_URL")
  os.Setenv("HTTP_PROXY", fixieURL)
  os.Setenv("HTTPS_PROXY", fixieURL)
}

func lineClient() (*linebot.Client, error) {
  lineChannelID, err := strconv.Atoi(os.Getenv("LINE_CHANNEL_ID"))
  if err != nil {
    return nil, err
  }
  lineChannelSecret := os.Getenv("LINE_CHANNEL_SECRET")
  lineMID := os.Getenv("LINE_MID")
  bot, err := linebot.NewClient(int64(lineChannelID), lineChannelSecret, lineMID)
  return bot, err
}

