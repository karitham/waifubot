package memstore

import (
	"context"
	"os"
	"testing"

	"github.com/karitham/waifubot/discord"
	"github.com/Karitham/corde"
	"github.com/go-redis/redis/v8"
)

func TestClient_Char(t *testing.T) {
	rc, _ := redis.ParseURL(getEnv(t, "REDIS_URL"))
	c := New(rc)

	type args struct {
		ctx       context.Context
		char      discord.MediaCharacter
		channelID corde.Snowflake
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "set & get",
			args: args{
				ctx: context.Background(),
				char: discord.MediaCharacter{
					ID:          169493,
					Name:        "Korone Inugami",
					ImageURL:    "https://s4.anilist.co/file/anilistcdn/character/large/b169493-rdLkoVzej1vk.png",
					URL:         "https://anilist.co/character/169493",
					Description: "DOOG",
					MediaTitle:  "Holo no Graffiti",
				},
				channelID: corde.Snowflake(287739410286379019),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.SetChannelChar(tt.args.ctx, tt.args.channelID, tt.args.char); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetChannelChar() error = %v, wantErr %v", err, tt.wantErr)
			}

			char, err := c.GetChannelChar(tt.args.ctx, tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetChannelChar() error = %v, wantErr %v", err, tt.wantErr)
			}

			if char != tt.args.char {
				t.Errorf("Client.GetChannelChar() = %+v, want %+v", char, tt.args.char)
			}

			if err := c.RemoveChannelChar(tt.args.ctx, tt.args.channelID); (err != nil) != tt.wantErr {
				t.Errorf("Client.RemoveChannelChar() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, err := c.GetChannelChar(tt.args.ctx, tt.args.channelID); (err == nil) != tt.wantErr {
				t.Errorf("Client.GetChannelChar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getEnv(t *testing.T, name string) string {
	t.Helper()
	e := os.Getenv(name)
	if e == "" {
		t.Skipf("%s not defined", name)
		return ""
	}

	return e
}
