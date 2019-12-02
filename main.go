package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token      string
	lastEvent  time.Time
	recordTime time.Duration
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating Discord session:", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening connection:", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	for _, u := range m.Mentions {
		if u.ID == s.State.User.ID {
			printDays(s, m)
			return
		}
	}

	if strings.Contains(strings.ToLower(m.Content), "piss") {
		handleWatchWord(s, m)
	}
}

func handleWatchWord(s *discordgo.Session, m *discordgo.MessageCreate) {

	messageTime, err := m.Timestamp.Parse()
	if err != nil {
		fmt.Println("error parsing discord timestamp,", err)
		return
	}

	if messageTime.Sub(lastEvent) > time.Hour {
		printDays(s, m)
	}

	lastEvent = messageTime
}

func printDays(s *discordgo.Session, m *discordgo.MessageCreate) {

	if lastEvent.IsZero() {
		s.ChannelMessageSend(m.ChannelID, "No piss since I started counting")
		return
	}

	since := time.Since(lastEvent)
	days := int(math.Floor(since.Hours() / 24))

	if since > recordTime {
		recordTime = since
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Days without piss: %d\n\nTime since last piss: %s\nLongest time between pisses: %s", days, since.String(), recordTime.String()))
}
