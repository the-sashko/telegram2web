package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
)

func main() {
	type PhotoObject struct {
		ID     string `json:"file_id"`
		Size   int    `json:"file_size"`
		Path   string `json:"file_path"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	type ChatObject struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Username string `json:"username"`
		Type     string `json:"type"`
	}
	type ChannelPostObject struct {
		ID      int           `json:"message_id"`
		Chat    ChatObject    `json:"chat"`
		Date    int           `json:"date"`
		Text    string        `json:"text"`
		Photos  []PhotoObject `json:"photo"`
		Caption string        `json:"caption"`
	}
	type EditedChannelPostObject struct {
		ID       int           `json:"message_id"`
		Chat     ChatObject    `json:"chat"`
		Date     int           `json:"date"`
		EditDate int           `json:"edit_date"`
		Text     string        `json:"text"`
		Photos   []PhotoObject `json:"photo"`
		Caption  string        `json:"caption"`
	}
	type APIResponseMessage struct {
		ID                int                     `json:"update_id"`
		ChannelPost       ChannelPostObject       `json:"channel_post"`
		EditedChannelPost EditedChannelPostObject `json:"edited_channel_post"`
	}
	type APIResponse struct {
		Status bool                 `json:"ok"`
		List   []APIResponseMessage `json:"result"`
	}
	type EntryObject struct {
		ID    int
		Title string
		Text  string
		File  string
		Date  int
		UPD   int
	}
	var Resp APIResponse
	var Posts []*EntryObject

	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		os.Mkdir("tmp", os.ModePerm)
	}

	BotAPIKey := "***"
	ChatID := "@chatid"
	HTTPresponse, err := http.Get("https://api.telegram.org/bot" + BotAPIKey + "/getUpdates?chat_id=" + ChatID)
	if err == nil {
		defer HTTPresponse.Body.Close()
		JSONBlob, err := ioutil.ReadAll(HTTPresponse.Body)
		if err == nil {
			err := json.Unmarshal(JSONBlob, &Resp)
			if err != nil {
				fmt.Println("error:", err)
			}
			//fmt.Println(Resp)
			if Resp.Status {
				db, err := bolt.Open("data/data.db", 0600, nil)
				if err != nil {
					log.Fatal(err)
				}

				db.Update(func(tx *bolt.Tx) error {
					dbBucket, err := tx.CreateBucketIfNotExists([]byte("Posts"))
					if err != nil {
						return fmt.Errorf("create bucket: %s", err)
					}
					dbBucket.Put([]byte("id"), []byte("42"))
					return nil
				})

				defer db.Close()
				for _, Message := range Resp.List {
					entry := new(EntryObject)
					if Message.ChannelPost.ID > 0 {
						if len(Message.ChannelPost.Text) > 0 {
							entry.Text = Message.ChannelPost.Text
						}
						if len(Message.ChannelPost.Caption) > 0 {
							entry.Text = Message.ChannelPost.Caption
						}
						if len(Message.ChannelPost.Photos) > 0 {
							photoSize := 0
							for _, Photo := range Message.ChannelPost.Photos {
								if photoSize < Photo.Size {
									photoSize = Photo.Size
									entry.File = Photo.ID
								}
							}
						}
						if len(entry.Text) > 0 || len(entry.File) > 0 {
							entry.ID = Message.ChannelPost.ID
							entry.Date = Message.ChannelPost.Date
							entry.UPD = Message.ChannelPost.Date
							Posts = append(Posts, entry)
							//fmt.Println(reflect.TypeOf(Posts))
						}
					}
				}
				for _, Message := range Resp.List {
					entry := new(EntryObject)
					if Message.EditedChannelPost.ID > 0 {
						if len(Message.EditedChannelPost.Text) > 0 {
							entry.Text = Message.EditedChannelPost.Text
						}
						if len(Message.EditedChannelPost.Caption) > 0 {
							entry.Text = Message.EditedChannelPost.Caption
						}
						if len(Message.EditedChannelPost.Photos) > 0 {
							photoSize := 0
							for _, Photo := range Message.EditedChannelPost.Photos {
								if photoSize < Photo.Size {
									photoSize = Photo.Size
									entry.File = Photo.ID
								}
							}
						}
						if len(entry.Text) > 0 || len(entry.File) > 0 {
							entry.ID = Message.EditedChannelPost.ID
							entry.Date = Message.EditedChannelPost.Date
							entry.UPD = Message.EditedChannelPost.EditDate
						}
					}
				}
				fmt.Println(Posts[1])
			}
		}
	}
}
