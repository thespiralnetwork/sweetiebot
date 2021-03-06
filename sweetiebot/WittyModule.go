package sweetiebot

import (
	"math/rand"
	"regexp"
	"strings"

	"github.com/blackhole12/discordgo"
)

// WittyModule is intended for any witty comments sweetie bot makes in response to what users say or do.
type WittyModule struct {
	lastdelete   int64
	lastcomment  int64
	wittyregex   *regexp.Regexp
	triggerregex []*regexp.Regexp
	remarks      [][]string
}

// Name of the module
func (w *WittyModule) Name() string {
	return "Witty"
}

// Commands in the module
func (w *WittyModule) Commands() []Command {
	return []Command{
		&addWitCommand{w},
		&removeWitCommand{w},
	}
}

// Description of the module
func (w *WittyModule) Description() string {
	return "In response to certain patterns (determined by a regex) will post a response picked randomly from a list of them associated with that trigger. Rate limits itself to make sure it isn't too annoying."
}

// UpdateRegex updates the witty module regex
func (w *WittyModule) UpdateRegex(info *GuildInfo) bool {
	l := len(info.config.Witty.Responses)
	w.triggerregex = make([]*regexp.Regexp, 0, l)
	w.remarks = make([][]string, 0, l)
	if l < 1 {
		w.wittyregex = nil
		return true
	}

	var err error
	w.wittyregex, err = regexp.Compile("(" + strings.Join(MapStringToSlice(info.config.Witty.Responses), "|") + ")")

	if err == nil {
		var r *regexp.Regexp
		for k, v := range info.config.Witty.Responses {
			r, err = regexp.Compile(k)
			if err != nil {
				break
			}
			w.triggerregex = append(w.triggerregex, r)
			w.remarks = append(w.remarks, strings.Split(v, "|"))
		}
	}

	if len(w.triggerregex) != len(w.remarks) { // This should never happen but we check just in case
		info.Log("ERROR! triggers do not equal remarks!!")
		return false
	}
	return err == nil
}

func (w *WittyModule) sendWittyComment(channel string, comment string, info *GuildInfo) {
	if RateLimit(&w.lastcomment, info.config.Witty.Cooldown) {
		info.SendMessage(channel, comment)
	}
}

// OnMessageCreate discord hook
func (w *WittyModule) OnMessageCreate(info *GuildInfo, m *discordgo.Message) {
	str := strings.ToLower(m.Content)
	if CheckRateLimit(&w.lastcomment, info.config.Witty.Cooldown) {
		if w.wittyregex != nil && w.wittyregex.MatchString(str) {
			for i := 0; i < len(w.triggerregex); i++ {
				if w.triggerregex[i].MatchString(str) {
					w.sendWittyComment(m.ChannelID, w.remarks[i][rand.Intn(len(w.remarks[i]))], info)
					break
				}
			}
		}
	}
}

type addWitCommand struct {
	wit *WittyModule
}

func (c *addWitCommand) Name() string {
	return "AddWit"
}
func WitRemove(wit string, info *GuildInfo) bool {
	wit = strings.ToLower(wit)
	_, ok := info.config.Witty.Responses[wit]
	if ok {
		delete(info.config.Witty.Responses, wit)
	}
	return ok
}

func (c *addWitCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) < 2 {
		return "```You must provide both a trigger and a remark (both must be in quotes if they have spaces).```", false, nil
	}

	trigger := strings.ToLower(args[0])
	remark := args[1]

	CheckMapNilString(&info.config.Witty.Responses)
	info.config.Witty.Responses[trigger] = remark
	info.SaveConfig()
	r := c.wit.UpdateRegex(info)
	if !r {
		WitRemove(trigger, info)
		c.wit.UpdateRegex(info)
		return "```Failed to add " + trigger + " because regex compilation failed.```", false, nil
	}
	return "```Adding " + trigger + " and recompiled the wittyremarks regex.```", false, nil
}
func (c *addWitCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Adds a `response` that is triggered by `trigger`.",
		Params: []CommandUsageParam{
			{Name: "trigger", Desc: "Any valid regex string, but it must be in quotes if it has spaces.", Optional: false},
			{Name: "response", Desc: "All possible responses, split up by `|`. Also requires quotes if it has spaces.", Optional: false},
		},
	}
}
func (c *addWitCommand) UsageShort() string { return "Adds a line to wittyremarks." }

type removeWitCommand struct {
	wit *WittyModule
}

func (c *removeWitCommand) Name() string {
	return "RemoveWit"
}

func (c *removeWitCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) < 1 {
		return "```You must provide both a trigger to remove!```", false, nil
	}

	arg := strings.Join(args, " ")
	if !WitRemove(arg, info) {
		return "```Could not find " + arg + "!```", false, nil
	}
	info.SaveConfig()
	c.wit.UpdateRegex(info)
	return "```Removed " + arg + " and recompiled the wittyremarks regex.```", false, nil
}
func (c *removeWitCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Removes `trigger` from wittyremarks, provided it exists.",
		Params: []CommandUsageParam{
			{Name: "trigger", Desc: "Any valid regex string.", Optional: false},
		},
	}
}
func (c *removeWitCommand) UsageShort() string { return "Removes a remark from wittyremarks." }
