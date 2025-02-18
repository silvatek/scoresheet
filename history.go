package main

import (
	"context"
	"fmt"
	"strings"
)

type HistoryItem struct {
	ItemType string
	ItemCode string
	Summary  string
	UrlPath  string
}

func GetHistory(ctx context.Context, cookieValue string) []HistoryItem {
	var ids []string
	var history []HistoryItem
	if cookieValue != "" {
		ids = strings.Split(cookieValue, " ")
		for _, id := range ids {
			if strings.TrimSpace(id) == "" {
				continue
			}
			parts := strings.Split(id, ":")
			var itemType string
			if len(parts) == 2 {
				itemType = parts[0]
				id = parts[1]
			} else {
				itemType = "Game"
			}
			var item HistoryItem
			if strings.ToLower(itemType) == "game" {
				game := dataStore.getGame(ctx, id)
				item = HistoryItem{
					ItemType: itemType,
					ItemCode: game.ID,
					Summary:  game.Title,
					UrlPath:  "/game/" + game.ID,
				}
			} else if strings.ToLower(itemType) == "list" {
				list := dataStore.getList(ctx, id)
				item = HistoryItem{
					ItemType: itemType,
					ItemCode: list.ID,
					Summary:  list.Name,
					UrlPath:  "/list/" + list.ID,
				}
			}

			if item.Summary != "" {
				history = append(history, item)
			}
		}
	}
	return history
}

func AddToHistory(newValue string, existingList string) string {
	return newValue + " " + strings.Trim(strings.ReplaceAll(existingList, newValue, " "), " ")
}

func HistoryString(items []HistoryItem) string {
	text := ""
	for _, item := range items {
		text += fmt.Sprintf("%s:%s ", item.ItemType, item.ItemCode)
	}
	return strings.TrimSpace(text)
}
