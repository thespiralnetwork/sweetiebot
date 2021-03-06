package sweetiebot

import (
	"fmt"
	"time"

	"github.com/blackhole12/discordgo"
)

// BoredModule picks a random action to do whenever #manechat has been idle for several minutes (configurable)
type BoredModule struct {
	lastmessage int64 // Ensures discord screwing up doesn't make us spam the chatroom.
}

// Name of the module
func (w *BoredModule) Name() string {
	return "Bored"
}

// Commands in the module
func (w *BoredModule) Commands() []Command { return []Command{} }

// Description of the module
func (w *BoredModule) Description() string {
	return "After the chat is inactive for a given amount of time, chooses a random action from the `boredcommands` configuration option to run, such posting a link from the bored collection or throwing an item from her bucket."
}

// OnIdle discord hook
func (w *BoredModule) OnIdle(info *GuildInfo, c *discordgo.Channel) {
	id := c.ID

	if RateLimit(&w.lastmessage, w.IdlePeriod(info)) && len(info.config.Bored.Commands) > 0 {
		m := &discordgo.Message{ChannelID: id, Content: MapGetRandomItem(info.config.Bored.Commands),
			Author: &discordgo.User{
				ID:       sb.SelfID,
				Username: "Sweetie",
				Verified: true,
				Bot:      true,
			},
			Timestamp: discordgo.Timestamp(time.Now().UTC().Format(time.RFC3339Nano)),
		}
		fmt.Println("Sending bored command ", m.Content, " on ", id)

		SBProcessCommand(sb.dg, m, info, time.Now().UTC().Unix(), sb.IsDBGuild(info), info.IsDebug(m.ChannelID))
	}
}

// IdlePeriod discord hook
func (w *BoredModule) IdlePeriod(info *GuildInfo) int64 {
	return info.config.Bored.Cooldown
}
