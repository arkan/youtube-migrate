package yt_migrate

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// getHTTPClient ...
func getHTTPClient() (*http.Client, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/youtube")
	if err != nil {
		return nil, err
	}

	config.RedirectURL = "http://localhost:8080"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	tok, err := getTokenFromWeb(config, authURL)
	if err != nil {
		return nil, err
	}

	return config.Client(ctx, tok), nil
}

func startWebServer() (chan string, error) {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return nil, err
	}
	codeCh := make(chan string)

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("In the handler\n")
		code := r.FormValue("code")
		codeCh <- code // send code to OAuth flow
		close(codeCh)
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	}))

	return codeCh, nil
}

func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Cannot open URL %s on this platform", url)
	}
	return err
}

func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token %v", err)
	}
	return tok, nil
}

func getTokenFromWeb(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	codeCh, err := startWebServer()
	if err != nil {
		fmt.Printf("Unable to start a web server.")
		return nil, err
	}

	if err := openURL(authURL); err != nil {
		log.Fatalf("Unable to open authorization URL in web server: %v", err)
	}

	fmt.Println("Your browser has been opened to an authorization URL.", " This program will resume once authorization has been provided.")
	fmt.Println(authURL)

	// Wait for the web server to get the code.
	log.Printf("Waiting code\n")
	code := <-codeCh
	log.Printf("Code received: %s\n", code)
	return exchangeToken(config, code)
}
