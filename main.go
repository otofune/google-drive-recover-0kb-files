package main

import (
	"fmt"
	"log"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func main() {
	oauth, err := getOAuth2Config()
	if err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}
	client, err := getClient(oauth)
	if err != nil {
		log.Fatalf("Unable to setup client: %v", err)
	}

	d, err := drive.New(client)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fileField := googleapi.Field(strings.Join([]string{
		"kind",
		"id",
		"name",
		"modifiedTime",
		"mimeType",
		"fileExtension", // binary でしか有効でないらしいので判定に使う
		"quotaBytesUsed",
		"headRevisionId",
	}, ","))
	fields := []googleapi.Field{
		"kind",
		"nextPageToken",
		"incompleteSearch",
		"files(" + fileField + ")",
	}
	q := `(not mimeType contains 'application/vnd.google-apps') and trashed = false and 'me' in owners`

	nextPageToken := ""
INFINITY_READ:
	for i := 0; i == 0 || nextPageToken != ""; i++ {
		fmt.Printf("# Listing %d\n", i)
		list, err := d.Files.List().Fields(fields...).OrderBy("quotaBytesUsed asc").Q(q).Spaces("drive").PageSize(100).PageToken(nextPageToken).Do()
		if err != nil {
			log.Fatalf("ファイル検索に失敗: %v", err)
		}
		if list.IncompleteSearch {
			fmt.Println("Incomplete Search, break")
			break INFINITY_READ
		}
		nextPageToken = list.NextPageToken
		for _, f := range list.Files {
			// f := f
			if f.QuotaBytesUsed != 0 {
				fmt.Println("Non empty file encounted, break")
				break INFINITY_READ
			}

			fmt.Printf("# File %s(id: %s) is empty. rev: %s\n%s\n", f.Name, f.Id, f.HeadRevisionId, f.Id)
			if err := d.Revisions.Delete(f.Id, f.HeadRevisionId).Do(); err != nil {
				log.Fatalf("revision(%s, %s) の削除に失敗した: %+v", f.Id, f.HeadRevisionId, err)
			}
		}
	}
}
