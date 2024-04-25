package spotifyCodeClient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cenkalti/dominantcolor"
)

type SpotifyCode struct {
	spClientId     string
	spClientSecret string
	spAccessToken  string
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifyAlbumResponse struct {
	Images []Image `json:"images"`
}

type Image struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func New(clientId, clientSecret string) *SpotifyCode {
	sp := &SpotifyCode{spClientId: clientId, spClientSecret: clientSecret}
	sp.getAccessToken()
	return sp
}

// getAccessToken retrieves the Spotify Access Token that is used for making requests.
func (sc *SpotifyCode) getAccessToken() {

	authHeader := base64.StdEncoding.EncodeToString([]byte(sc.spClientId + ":" + sc.spClientSecret))

	requestBody := []byte("grant_type=client_credentials")

	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBuffer(requestBody))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", authHeader))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var tokenResponse AccessTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		log.Fatal("Error decoding response:", err)

	}

	sc.spAccessToken = tokenResponse.AccessToken
}

// getSpotifyImage returns an image an album/artist/song.
func (sc *SpotifyCode) getSpotifyImage(spotifyUri string) image.Image {

	parts := strings.SplitN(spotifyUri, ":", 3)
	if len(parts) != 3 {
		log.Fatal("SpotifyURI is malformed!")
	}
	_, content_type, uri := parts[0], parts[1], parts[2]

	var url string
	switch content_type {
	case "album":
		url = fmt.Sprintf("https://api.spotify.com/v1/albums/%v", uri)
	case "artist":
	default:
		log.Fatal("No implemented behavior for content_type!")
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", sc.spAccessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var spotifyAlbumResponse SpotifyAlbumResponse
	if err := json.NewDecoder(res.Body).Decode(&spotifyAlbumResponse); err != nil {
		log.Fatal("Error decoding response:", err)

	}

	req, _ = http.NewRequest("GET", spotifyAlbumResponse.Images[0].Url, nil)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	if err != nil {
		log.Fatal("Error decoding image:", err)
	}
	return img

}

// getCode returns the Spotify Code generated on www.spotifycodes.com
func (sc *SpotifyCode) getCode(spotifyUri string, color string, barColor string, size int, format string) image.Image {

	color = strings.Replace(color, "#", "", 1)
	raw_url := fmt.Sprintf("https://www.spotifycodes.com/downloadCode.php?uri=%v", format)
	raw_url = raw_url + url.QueryEscape(fmt.Sprintf("/%v/%v/%d/%v", color, barColor, size, spotifyUri))
	req, _ := http.NewRequest("GET", raw_url, nil)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	if err != nil {
		log.Fatal("Error decoding image:", err)
	}
	return img

}

func (sc *SpotifyCode) SaveImage(img image.Image, fname string) {
	fname = fmt.Sprintf("%v.png", fname)
	file, err := os.Create(fname)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("Error encoding image:", err)
		return
	}
}

func (sc *SpotifyCode) GetCode(spotifyUri string) image.Image {

	spotifyImage := sc.getSpotifyImage(spotifyUri)
	dominantColor := dominantcolor.Hex(dominantcolor.Find(spotifyImage))
	size := spotifyImage.Bounds().Dx()
	spotifyCode := sc.getCode(spotifyUri, dominantColor, "white", size, "png")

	// stitch image and code together
	result := image.NewRGBA(image.Rect(0, 0, spotifyImage.Bounds().Dx(), spotifyImage.Bounds().Dy()+spotifyCode.Bounds().Dy()))
	draw.Draw(result, spotifyImage.Bounds(), spotifyImage, image.Point{0, 0}, draw.Src)
	draw.Draw(result, spotifyCode.Bounds().Add(image.Point{0, spotifyImage.Bounds().Dy()}), spotifyCode, image.Point{0, 0}, draw.Src)

	return result
}
