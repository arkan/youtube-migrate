package yt_migrate

import (
	"log"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

// Client ...
type Client struct {
	service *youtube.Service
}

// New ...
func New() (*Client, error) {
	httpClient, err := getHTTPClient()
	if err != nil {
		return nil, err
	}

	service, err := youtube.New(httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{service: service}, nil
}

// GetSubscriptions returns the list of all the subscriptions.
func (c *Client) GetSubscriptions() ([]string, error) {
	log.Printf("[GetSubscriptions]")
	ret := []string{}

	call := c.service.Subscriptions.List([]string{"snippet", "contentDetails"})
	call = call.Mine(true)
	call = call.MaxResults(50)

	for {
		log.Printf("[GetSubscriptions]: len(%d)", len(ret))

		response, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, i := range response.Items {
			ret = append(ret, i.Snippet.ResourceId.ChannelId)
		}

		if response.NextPageToken == "" {
			break
		}

		call = call.PageToken(response.NextPageToken)
	}

	return ret, nil
}

// AddSubscription subscribes to a given channelID.
func (c *Client) AddSubscription(channelID string) error {
	log.Printf("[AddSubscription] ChannelID: %q", channelID)

	sub := &youtube.Subscription{
		Snippet: &youtube.SubscriptionSnippet{
			ResourceId: &youtube.ResourceId{
				Kind:      "youtube#channel",
				ChannelId: channelID,
			},
		},
	}
	callInsert := c.service.Subscriptions.Insert([]string{"snippet"}, sub)
	_, err := callInsert.Do()
	if err != nil {
		gerr, ok := err.(*googleapi.Error)
		if ok {
			if gerr.Code == 400 && gerr.Message == "The subscription that you are trying to create already exists." {
				return nil
			}
		}

		return err
	}

	return nil
}
