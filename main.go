package main

import (
	"github.com/tilschuenemann/golang-spotifycode/spotifyCodeClient"
)

func main() {

	clientId := ""
	clientSecret := ""

	uri := "spotify:album:6BzxX6zkDsYKFJ04ziU5xQ"

	sc := spotifyCodeClient.New(clientId, clientSecret)
	img := sc.GetCode(uri)
	sc.SaveImage(img, "beyonce")
}
