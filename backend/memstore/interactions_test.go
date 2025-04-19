package memstore

import (
	"context"
	"testing"

	"github.com/Karitham/corde"
	"github.com/go-redis/redis/v8"
)

func TestClient_IncrementInteractionCount(t *testing.T) {
	u, _ := redis.ParseURL(getEnv(t, "REDIS_URL"))
	c := New(u)

	tests := []struct {
		ctx       context.Context
		name      string
		channelID corde.Snowflake
		wantErr   bool
	}{
		{
			ctx:       context.Background(),
			name:      "increment",
			channelID: corde.Snowflake(287739410286379019),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.IncrementInteractionCount(tt.ctx, tt.channelID); (err != nil) != tt.wantErr {
				t.Errorf("Client.IncrementInteractionCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if c, err := c.GetInteractionCount(tt.ctx, tt.channelID); (err != nil) != tt.wantErr && c != 1 {
				t.Errorf("Client.GetInteractionCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := c.ResetInteractionCount(tt.ctx, tt.channelID); (err != nil) != tt.wantErr {
				t.Errorf("Client.ResetInteractionCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if c, err := c.GetInteractionCount(tt.ctx, tt.channelID); (err != nil) != tt.wantErr && c != 0 {
				t.Errorf("Client.GetInteractionCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
