package discord

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Karitham/corde"
)

func (b *Bot) info(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	infob := strings.Builder{}
	infob.WriteString("Version: ")
	infob.WriteString(runtime.Version())
	infob.WriteRune('\n')

	infob.WriteString("Runtime: ")
	infob.WriteString(runtime.GOOS + "/" + runtime.GOARCH)
	infob.WriteRune('\n')

	infob.WriteString("Goroutines: ")
	infob.WriteString(strconv.Itoa(runtime.NumGoroutine()))
	infob.WriteRune('\n')

	s := &runtime.MemStats{}
	runtime.ReadMemStats(s)
	infob.WriteString("Memory: ")
	infob.WriteString(strconv.Itoa(int(s.Alloc / 1024)))
	infob.WriteString(" KB")
	infob.WriteRune('\n')

	infob.WriteString("GC Pauses: ")
	infob.WriteString(time.Duration(int64(s.PauseTotalNs)).String())
	infob.WriteRune('\n')

	w.Respond(corde.NewResp().Ephemeral().Embeds(
		corde.NewEmbed().
			Title("Info").
			Descriptionf(
				"A gacha game bot to collect and trade characters, and discover anything manga related.\n```info\n%s```",
				infob.String(),
			).
			URL("https://github.com/Karitham/WaifuBot/tree/corde"),
	))
}
