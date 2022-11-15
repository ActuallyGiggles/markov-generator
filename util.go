package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	Config = Configuration{}
)

func addUser(user string) {
	u := User{
		Name: user,
	}
	Users = append(Users, u)
}

func RandomNumber(min int, max int) (num int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	num = r.Intn(max-min) + min
	return num
}

func readConfig() {
	f, err := os.Open("./config.json")
	defer f.Close()
	if err != nil {
		configSetup()
		return
	}

	err = json.NewDecoder(f).Decode(&Config)
	if err != nil {
		panic(err)
	}

	for _, channel := range Config.Channels {
		addUser(channel)
	}
}

func writeConfig() {
	f, err := os.OpenFile("./config.json", os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c, err := json.MarshalIndent(Config, "", " ")
	if err != nil {
		panic(err)
	}

	_, err = f.Write(c)
	if err == nil {
		fmt.Println("Config successfully created.")
	}

	readConfig()
}

func configSetup() {
	// Intro
	clearTerminal()
	fmt.Println("First time? Let's setup your bot.")
	fmt.Println()
	fmt.Println("Press Enter...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	clearTerminal()

	// Name
	fmt.Println("First, what is the login name of the account you are going to use?")
	fmt.Println()
	fmt.Print("Name: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := strings.ToLower(scanner.Text())
	Config.Name = name
	clearTerminal()

	// Client ID
	fmt.Println("Now let's get your Client ID, it can be found here after registering your application: (https://dev.twitch.tv/console). \n\nSteps:\n1. Give your application a name.\n2. Set the redirect URL to (https://twitchapps.com/tokengen/).\n3. Choose the chatbot category.\n4. Copy and paste the Client ID here.")
	fmt.Println()
	fmt.Print("Client ID: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	clientID := scanner.Text()
	Config.ClientID = clientID
	clearTerminal()

	// OAuth
	fmt.Println("Let's generate an OAuth. To do that, go to this website (https://twitchapps.com/tokengen/). \n\nSteps:\n1. Paste in the Client ID\n2. For scopes, type in: 'chat:read chat:edit'.\n3. Click connect and copy and paste the OAuth Token here.")
	fmt.Println()
	fmt.Print("OAuth: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	oauth := scanner.Text()
	Config.OAuth = oauth
	clearTerminal()

	// Channels
	fmt.Println(`Now list the channels that you want your bot to be active in. (Separate them with spaces.) (Example: "39daph nmplol sodapoppin veibae")`)
	fmt.Println()
	fmt.Print("Channels: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	channels := scanner.Text()
	for _, channel := range strings.Split(channels, " ") {
		Config.Channels = append(Config.Channels, channel)
	}
	clearTerminal()

	// Blacklisted Emotes
	fmt.Println(`Enter the emotes that you want to be blacklisted. (Separate them with spaces.) (Example: "TriHard KEKW ResidentSleeper")`)
	fmt.Println()
	fmt.Print("Blacklist emotes: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	blacklist := scanner.Text()
	for _, blacked := range strings.Split(blacklist, " ") {
		Config.BlacklistEmotes = append(Config.BlacklistEmotes, blacked)
	}
	clearTerminal()

	// Message Sample
	fmt.Println("How many messages do you want the bot the sample at a time? (Recommended: 10)")
	fmt.Println()
	fmt.Print("Sample: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	sample := scanner.Text()
	s, err := strconv.Atoi(sample)
	if err != nil {
		fmt.Println(sample, "is not a number!")
		os.Exit(3)
	}
	Config.MessageSample = s
	clearTerminal()

	// Message Threshold
	fmt.Println("Out of that sample size, how many times does an emote have to repeat itself to force your account to send it? (Recommended: 3)")
	fmt.Println()
	fmt.Print("Threshold: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	threshold := scanner.Text()
	t, err := strconv.Atoi(threshold)
	if err != nil {
		fmt.Println(threshold, "is not a number!")
		os.Exit(3)
	}
	Config.MessageThreshold = t
	clearTerminal()

	// Messaging Interval
	fmt.Println(`Finally, please specify the range of minutes for the bot to wait in between message sends. (Example: "5 10")`)
	fmt.Println()
	fmt.Print("Range: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	interval := scanner.Text()
	tS := strings.Split(interval, " ")
	min, err := strconv.Atoi(tS[0])
	if err != nil {
		fmt.Println(tS[0], "is not a number!")
		os.Exit(3)
	}
	max, err := strconv.Atoi(tS[1])
	if err != nil {
		fmt.Println(tS[1], "is not a number!")
		os.Exit(3)
	}
	Config.IntervalMin = min
	Config.IntervalMax = max
	fmt.Println()

	writeConfig()
}

func clearTerminal() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
