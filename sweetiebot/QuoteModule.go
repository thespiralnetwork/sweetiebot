package sweetiebot

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/blackhole12/discordgo"
)

// QuoteModule manages the quoting system
type QuoteModule struct {
}

// Name of the module
func (w *QuoteModule) Name() string {
	return "Quotes"
}

// Commands in the module
func (w *QuoteModule) Commands() []Command {
	return []Command{
		&quoteCommand{},
		&addquoteCommand{},
		&removequoteCommand{},
		&searchQuoteCommand{},
	}
}

// Description of the module
func (w *QuoteModule) Description() string { return "Manages the quoting system." }

type quoteCommand struct {
}

func (c *quoteCommand) Name() string {
	return "Quote"
}
func (c *quoteCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) < 1 {
		l := 0
		for _, v := range info.config.Quote.Quotes {
			l += len(v)
		}
		if l <= 0 {
			return "```There are no quotes.```", false, nil
		}
		i := rand.Intn(l)

		for k, v := range info.config.Quote.Quotes {
			if i < len(v) {
				return "**" + getUserName(k, info) + "**: " + v[i], false, nil
			}
			i -= len(v)
		}
		return "```Error: invalid random quote chosen???```", false, nil
	}

	arg := strings.ToLower(args[0])
	IDs := FindUsername(arg, info)
	if len(IDs) == 0 { // no matches!
		return "```Error: Could not find any usernames or aliases matching " + arg + "!```", false, nil
	}
	if len(IDs) > 1 {
		return "```Could be any of the following users or their aliases:\n" + strings.Join(IDsToUsernames(IDs, info, true), "\n") + "```", len(IDs) > 5, nil
	}

	q, ok := info.config.Quote.Quotes[IDs[0]]
	l := len(q)
	if !ok || l <= 0 {
		return "```That user has no quotes.```", false, nil
	}
	i := rand.Intn(l)
	if len(args) >= 2 {
		var err error
		i, err = strconv.Atoi(args[1])
		if err != nil {
			return "```Could not parse quote index. Make sure your username is in quotes. Use !searchquote [user] to list a user's quotes and their indexes.```", false, nil
		}
		i--
		if i >= l || i < 0 {
			return "```Invalid quote index. Use !searchquote [user] to list a user's quotes and their indexes.```", false, nil
		}
	}
	return "**" + IDsToUsernames(IDs, info, false)[0] + "**: " + q[i], false, nil
}
func (c *quoteCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "If no arguments are specified, returns a random quote. If a user is specified, returns a random quote from that user. If a quote index is specified, returns that specific quote.",
		Params: []CommandUsageParam{
			{Name: "user", Desc: "A @user ping or simply the name of the user to quote.", Optional: true},
			{Name: "quote", Desc: "A specific quote index. Use `" + info.config.Basic.CommandPrefix + "searchquote` to find a quote index.", Optional: true},
		},
	}
}
func (c *quoteCommand) UsageShort() string { return "Quotes a user." }

type addquoteCommand struct {
}

func (c *addquoteCommand) Name() string {
	return "AddQuote"
}
func (c *addquoteCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) < 1 {
		return "```Must specify username.```", false, nil
	}
	if len(args) < 2 {
		return "```Can't add a blank quote!```", false, nil
	}

	arg := strings.ToLower(args[0])
	IDs := FindUsername(arg, info)
	if len(IDs) == 0 { // no matches!
		return "```Error: Could not find any usernames or aliases matching " + arg + "!```", false, nil
	}
	if len(IDs) > 1 {
		return "```Could be any of the following users or their aliases:\n" + strings.Join(IDsToUsernames(IDs, info, true), "\n") + "```", len(IDs) > 5, nil
	}

	if len(info.config.Quote.Quotes) == 0 {
		info.config.Quote.Quotes = make(map[uint64][]string)
	}
	info.config.Quote.Quotes[IDs[0]] = append(info.config.Quote.Quotes[IDs[0]], msg.Content[indices[1]:])
	info.SaveConfig()
	return "```Quote added to " + IDsToUsernames(IDs, info, false)[0] + ".```", false, nil
}
func (c *addquoteCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Adds a quote to the quote database for the given user. If the user is ambiguous, sweetiebot will return all possible matches.",
		Params: []CommandUsageParam{
			{Name: "user", Desc: "A @user ping or simply the name of the user to quote. If the username has spaces, it must be in quotes.", Optional: false},
			{Name: "quote", Desc: "A specific quote index. Use `" + info.config.Basic.CommandPrefix + "searchquote` to find a quote index.", Optional: false},
		},
	}
}
func (c *addquoteCommand) UsageShort() string { return "Adds a quote." }

type removequoteCommand struct {
}

func (c *removequoteCommand) Name() string {
	return "RemoveQuote"
}
func (c *removequoteCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) < 1 {
		return "```Must specify username.```", false, nil
	}
	if len(args) < 2 {
		return "```Must specify quote index. Use !searchquote to list them.```", false, nil
	}

	arg := strings.ToLower(args[0])
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return "```Error: could not parse quote index. Did you surround your username with quotes? Use !searchquote to find a quote index.```", false, nil
	}

	IDs := FindUsername(arg, info)
	if len(IDs) == 0 { // no matches!
		return "```Error: Could not find any usernames or aliases matching " + arg + "!```", false, nil
	}
	if len(IDs) > 1 {
		return "```Could be any of the following users or their aliases:\n" + strings.Join(IDsToUsernames(IDs, info, true), "\n") + "```", len(IDs) > 5, nil
	}

	index--
	if index >= len(info.config.Quote.Quotes[IDs[0]]) || index < 0 {
		return "```Invalid quote index. Use !searchquote [user] to list a user's quotes and their indexes.```", false, nil
	}
	info.config.Quote.Quotes[IDs[0]] = append(info.config.Quote.Quotes[IDs[0]][:index], info.config.Quote.Quotes[IDs[0]][index+1:]...)
	info.SaveConfig()
	return "```Deleted quote #" + strconv.Itoa(index+1) + " from " + IDsToUsernames(IDs, info, false)[0] + ".```", false, nil
}
func (c *removequoteCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Removes the quote with the given quote index from the user's set of quotes. If the user is ambiguous, sweetiebot will return all possible matches.",
		Params: []CommandUsageParam{
			{Name: "user", Desc: "A @user ping or simply the name of the user to quote. If the username has spaces, it must be in quotes.", Optional: false},
			{Name: "quote", Desc: "A specific quote index. Use `" + info.config.Basic.CommandPrefix + "searchquote` to find a quote index.", Optional: false},
		},
	}
}
func (c *removequoteCommand) UsageShort() string { return "Removes a quote." }

type searchQuoteCommand struct {
}

func (c *searchQuoteCommand) Name() string {
	return "SearchQuote"
}
func (c *searchQuoteCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) < 1 {
		s := make([]uint64, 0, len(info.config.Quote.Quotes))
		for k, v := range info.config.Quote.Quotes {
			if len(v) > 0 { // Map entries can have 0 quotes associated with them
				s = append(s, k)
			}
		}
		return "```The following users have at least one quote:\n" + strings.Join(IDsToUsernames(s, info, true), "\n") + "```", len(s) > 6, nil
	}

	arg := strings.ToLower(args[0])
	IDs := FindUsername(arg, info)
	if len(IDs) == 0 { // no matches!
		return "```Error: Could not find any usernames or aliases matching " + arg + "!```", false, nil
	}
	if len(IDs) > 1 {
		return "```Could be any of the following users or their aliases:\n" + strings.Join(IDsToUsernames(IDs, info, true), "\n") + "```", len(IDs) > 5, nil
	}
	l := len(info.config.Quote.Quotes[IDs[0]])
	if l == 0 {
		return "```That user has no quotes.```", false, nil
	}
	quotes := make([]string, l, l)
	for i := 0; i < l; i++ {
		quotes[i] = strconv.Itoa(i+1) + ". " + info.config.Quote.Quotes[IDs[0]][i]
	}
	return "All quotes for " + IDsToUsernames(IDs, info, false)[0] + ":\n" + strings.Join(quotes, "\n"), l > 6, nil
}
func (c *searchQuoteCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Lists all quotes for the given user.",
		Params: []CommandUsageParam{
			{Name: "user", Desc: "A @user ping or simply the name of the user to quote. If the username has spaces, it must be in quotes.", Optional: false},
		},
	}
}
func (c *searchQuoteCommand) UsageShort() string { return "Finds a quote." }
