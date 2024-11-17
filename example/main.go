package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/choirulanwar/batchexecute"
)

type NewsItem struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Publisher string `json:"publisher,omitempty"`
}

func GetNews(topic, language, country string) ([]NewsItem, error) {
	params := batchexecute.Params{
		Host: "news.google.com",
		App:  "DotsSplashUi",
		Rpcs: []batchexecute.RpcRequest{
			{
				ID: "Qxytce",
				Args: []string{
					encodeArgs([]interface{}{
						"gtireq",
						[]interface{}{
							[]interface{}{language, country, []string{"FINANCE_TOP_INDICES", "WEB_TEST_1_0_0"}, nil, nil, 1, 1, fmt.Sprintf("%s:%s", country, language), nil, 420, nil, nil, nil, nil, nil, 0},
							language, country, 1, []int{2, 4, 8}, 1, 1, "686348916", 0, 0, nil, 0,
						},
						topic,
					}),
				},
			},
			{
				ID: "EAfJqe",
				Args: []string{
					encodeArgs([]interface{}{
						"gtsreq",
						[]interface{}{
							[]interface{}{language, country, []string{"FINANCE_TOP_INDICES", "WEB_TEST_1_0_0"}, nil, nil, 1, 1, fmt.Sprintf("%s:%s", country, language), nil, 420, nil, nil, nil, nil, nil, 0},
							language, country, 1, []int{2, 4, 8}, 1, 1, "686348916", 0, 0, nil, 0,
						},
						nil,
						topic,
						"",
					}),
				},
			},
		},
	}

	urlObj, body, err := batchexecute.PreparedBatchExecute(params)
	if err != nil {
		fmt.Println("Error preparing batch execution request:", err)
		return nil, err
	}

	headers := http.Header{}
	headers.Set("content-type", "application/x-www-form-urlencoded;charset=utf-8")

	_, respBody, err := batchexecute.Post(urlObj, headers, body)
	if err != nil {
		fmt.Println("Error sending POST request:", err)
		return nil, err
	}

	results, err := batchexecute.Decode(respBody, "", false, []string{})
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return nil, err
	}

	jsonResults, err := json.Marshal(results)
	if err != nil {
		fmt.Println("Error converting results to JSON:", err)
		return nil, err
	}

	parsed, err := parseJSON(string(jsonResults))
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func parseJSON(jsonString string) ([]NewsItem, error) {
	var rawData []map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &rawData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	var newsItems []NewsItem

	for _, item := range rawData {
		data := item["data"].([]interface{})
		if len(data) > 1 {

			mainArray, ok := data[1].([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid JSON structure: main array not found")
			}

			for _, entry := range mainArray {
				entryArray, ok := entry.([]interface{})
				if !ok {
					continue
				}

				for _, subEntry := range entryArray {
					subEntryArray, ok := subEntry.([]interface{})
					if !ok {
						continue
					}

					for _, subSubEntry := range subEntryArray {
						subSubEntryArray, ok := subSubEntry.([]interface{})
						if !ok {
							continue
						}

						for _, finalEntry := range subSubEntryArray {
							finalEntryArray, ok := finalEntry.([]interface{})
							if !ok {
								continue
							}

							if len(finalEntryArray) < 4 {
								continue
							}

							newsArray, ok := finalEntryArray[3].([]interface{})
							if !ok {
								continue
							}

							for _, newsEntry := range newsArray {
								newsEntryArray, ok := newsEntry.([]interface{})
								if !ok {
									continue
								}

								if len(newsEntryArray) < 11 {
									continue
								}

								title, ok := newsEntryArray[2].(string)
								if !ok {
									continue
								}

								url, ok := newsEntryArray[6].(string)
								if !ok {
									continue
								}

								var thumbnail string
								if newsEntryArray[8] != nil {
									thumbnailArray, ok := newsEntryArray[8].([]interface{})
									if ok && len(thumbnailArray) > 0 {
										thumbnailSubArray, ok := thumbnailArray[0].([]interface{})
										if ok && len(thumbnailSubArray) > 13 {
											thumbnail, _ = thumbnailSubArray[13].(string)
										}
									}
								}

								var publisher string
								if len(newsEntryArray[10].([]interface{})) > 2 {
									publisher, _ = newsEntryArray[10].([]interface{})[2].(string)
								}

								newsItems = append(newsItems, NewsItem{
									Title:     title,
									URL:       url,
									Thumbnail: thumbnail,
									Publisher: publisher,
								})
							}
						}
					}
				}
			}
		}
	}

	return newsItems, nil
}

func encodeArgs(args []interface{}) string {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}
	return string(argsJSON)
}

func main() {
	// Get the news data
	jsonResults, err := GetNews("CAAqJggKIiBDQkFTRWdvSUwyMHZNRGx1YlY4U0FtbGtHZ0pKUkNnQVAB", "en", "US")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the JSON string
	fmt.Println(jsonResults)
}
