package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token       string
	lastEvent   time.Time
	recordTime  time.Duration
	leaderboard map[string]int
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	leaderboard = make(map[string]int)

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

	leaderboard[m.Author.Username]++

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

	message := fmt.Sprintf("Days without piss: %d\n\nTime since last piss: %s\nLongest time between pisses: %s\nMost frequent pissers:\n%s", days, since.String(), recordTime.String(), getLeaderboard())

	s.ChannelMessageSend(m.ChannelID, message)
}

type entry struct {
	name  string
	count int
}

type entrySlice []entry

func (e entrySlice) Len() int {
	return len(e)
}

func (e entrySlice) Less(i, j int) bool {
	return e[i].count < e[j].count
}

func (e entrySlice) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func getLeaderboard() string {
	var entries entrySlice

	for n, c := range leaderboard {
		if c > 1 {
			entries = append(entries, entry{name: n, count: c})
		}
	}

	sort.Sort(sort.Reverse(entries))

	var s strings.Builder
	for i := 0; i < 3 && i < len(entries); i++ {
		s.WriteString(fmt.Sprintf("%s: %d\n", entries[i].name, entries[i].count))
	}

	return s.String()
}
