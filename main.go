package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	PhotoApi = "https://api.pexels.com/v1"
	VideoApi = "https://api.pexels.com/videos"
)

type Client struct {
	Token          string
	hc             http.Client
	RemainingTimes int32
}

func NewClient(token string) *Client {
	client := http.Client{}
	return &Client{Token: token, hc: client}
}

type SearchResult struct {
	Page         int32   `json:"page"`
	PerPage      int32   `json:"per_page"`
	TotalResults int32   `json:"total_results"`
	NextPage     string  `json:"next_page"`
	Photos       []Photo `json:"photos"`
}

type CuratedResult struct {
	Page     int32   `json:"page"`
	PerPage  int32   `json:"per_page"`
	NextPage string  `json:"next_page"`
	Photos   []Photo `json:"photos"`
}

type Photo struct {
	Id              int32       `json:"id"`
	Width           int32       `json:"width"`
	Height          int32       `json:"height"`
	Url             string      `json:"url"`
	Photographer    string      `json:"photographer"`
	PhotographerUrl string      `json:"photographer_url"`
	Src             PhotoSource `json:"src"`
}

type PhotoSource struct {
	Original  string `json:"original"`
	Large     string `json:"large"`
	Large2x   string `json:"large2x"`
	Medium    string `json:"medium"`
	Small     string `json:"small"`
	Portrait  string `json:"portrait"`
	Square    string `json:"square"`
	Landscape string `json:"landscape"`
	Tiny      string `json:"tiny"`
}

type VideoSearchResult struct {
	Page         int32   `json:"page"`
	PerPage      int32   `json:"per_page"`
	TotalResults int32   `json:"total_results"`
	NextPage     string  `json:"next_page"`
	Videos       []Video `json:"videos"`
}

type Video struct {
	Id            int32           `json:"id"`
	Width         int32           `json:"width"`
	Height        int32           `json:"height"`
	Url           string          `json:"url"`
	Image         string          `json:"image"`
	FullRes       interface{}     `json:"full_res"`
	Duration      float64         `json:"duration"`
	VideoFiles    []VideoFiles    `json:"video_files"`
	VideoPictures []VideoPictures `json:"video_pictures"`
}

type PopularVideos struct {
	Page         int32   `json:"page"`
	PerPage      int32   `json:"per_page"`
	TotalResults int32   `json:"total_results"`
	Url          string  `json:"url"`
	Videos       []Video `json:"videos"`
}

type VideoFiles struct {
	Id       int32  `json:"id"`
	Quality  string `json:"quality"`
	FileType string `json:"file_type"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
	Link     string `json:"link"`
}

type VideoPictures struct {
	Id      int32  `json:"id"`
	Picture string `json:"picture"`
	Nr      int32  `json:"nr"`
}

func (c *Client) SearchPhotos(query string, perPage, page int) (*SearchResult, error) {
	url := fmt.Sprintf("%s/search?query=%s&per_page=%d&page=%d", PhotoApi, query, perPage, page)
	resp, err := c.requestDoWithAuth("GET", url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SearchResult

	err = json.Unmarshal(data, &result)
	return &result, err
}

func (c *Client) CuratedPhotos(perPage, page int) (*CuratedResult, error) {
	url := fmt.Sprintf("%s/curated?per_page=%d&page=%d", PhotoApi, perPage, page)
	resp, err := c.requestDoWithAuth("GET", url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CuratedResult

	err = json.Unmarshal(data, &result)
	return &result, err
}

func (c *Client) requestDoWithAuth(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.Token)
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	times, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining"))
	if err != nil {
		return nil, err
	} else {
		c.RemainingTimes = int32(times)
	}

	return resp, nil
}

func (c *Client) GetPhoto(id int32) (*Photo, error) {
	url := fmt.Sprintf("%s/photos/%d", PhotoApi, id)
	resp, err := c.requestDoWithAuth("GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	var photo Photo
	err = json.Unmarshal(data, &photo)
	return &photo, err
}

func (c *Client) GetRandomPhoto() (*Photo, error) {
	rand.Seed(time.Now().UnixNano())
	page := rand.Intn(100)

	result, err := c.CuratedPhotos(1, page)
	if err != nil {
		return nil, err
	}

	if len(result.Photos) == 0 {
		return nil, fmt.Errorf("No photo found")
	}

	return &result.Photos[0], nil
}

func (c *Client) SearchVideo(query string, perPage, page int) (*VideoSearchResult, error) {
	url := fmt.Sprintf("%s/search?query=%s&per_page=%d&page=%d", VideoApi, query, perPage, page)
	resp, err := c.requestDoWithAuth("GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result VideoSearchResult
	err = json.Unmarshal(data, &result)
	return &result, err
}

func (c *Client) PopularVideo(perPage, page int) (*PopularVideos, error) {
	url := fmt.Sprintf("%s/popular?per_page=%d&page=%d", VideoApi, perPage, page)
	resp, err := c.requestDoWithAuth("GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result PopularVideos
	err = json.Unmarshal(data, &result)
	return &result, err
}

func (c *Client) GetRandomVideo() (*Video, error) {
	rand.Seed(time.Now().UnixNano())
	page := rand.Intn(100)

	result, err := c.PopularVideo(1, page)
	if err != nil {
		return nil, err
	}

	if len(result.Videos) == 0 {
		return nil, fmt.Errorf("No video found")
	}

	return &result.Videos[0], nil
}

func (c *Client) GetRemainingRequestInThisMonth() int32 {
	return c.RemainingTimes
}

func main() {
	os.Setenv("PexelsToken", "YOUR_PEXELS_API_KEY")
	
	var TOKEN = os.Getenv("PexelsToken")
	var c = NewClient(TOKEN)

	http.HandleFunc("/api/random-photo", func(w http.ResponseWriter, r *http.Request) {
		photo, err := c.GetRandomPhoto()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(photo)
	})

	http.HandleFunc("/api/random-video", func(w http.ResponseWriter, r *http.Request) {
		video, err := c.GetRandomVideo()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(video)
	})

	http.HandleFunc("/api/search-photo", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

		result, err := c.SearchPhotos(query, perPage, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/api/search-video", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

		result, err := c.SearchVideo(query, perPage, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/api/remaining-requests", func(w http.ResponseWriter, r *http.Request) {
		remainingRequests := c.GetRemainingRequestInThisMonth()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int32{"remaining_requests": remainingRequests})
	})

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
