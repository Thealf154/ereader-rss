package rssparser

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParseRSS(t *testing.T) {
	xmlFile, err := os.Open("./assets/newyorker.xml")
	if err != nil {
		t.Fatalf("error opening file: %s", err)
	}
	defer xmlFile.Close()

	xmlData, err := io.ReadAll(xmlFile)
	if err != nil {
		t.Fatalf("error reading file: %s", err)
	}

	rss, err := ReadRSS(&xmlData)
	if err != nil {
		t.Fatalf("error parsing file: %s", err)
	}

	// Print the channel information
	fmt.Printf("Channel Title: %s\n", rss.Channel.Title)
	fmt.Printf("Channel Link: %s\n", rss.Channel.Link)
	fmt.Printf("Channel Description: %s\n", rss.Channel.Description)
	fmt.Printf("Last Build Date: %s\n", rss.Channel.LastBuildDate)

	// Print each item in the channel
	for _, item := range rss.Channel.Items {
		fmt.Printf("\nItem Title: %s\n", item.Title)
		fmt.Printf("Item Description: %s\n", item.Description)
		fmt.Printf("Item Link: %s\n", item.Link)
		fmt.Printf("Item GUID: %s\n", item.Guid)
		fmt.Printf("Item PubDate: %s\n", item.PubDate)
		fmt.Printf("Item Category: %s\n", item.Category)
	}

}
