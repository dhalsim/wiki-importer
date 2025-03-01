package names

import (
	"context"
	"log"

	"fiatjaf/wiki-importer/common"

	"github.com/nbd-wtf/go-nostr"
	"github.com/urfave/cli/v3"
)

var (
	pool     *nostr.SimplePool
	nostrKey string
	relay    string
)

func HandleNames(ctx context.Context, l *log.Logger, c *cli.Command) error {
	var err error

	continueFrom := int(c.Uint("continue"))

	nostrKey, err = common.GetRequiredEnv("BEHINDTHENAME_NOSTR_KEY")
	if err != nil {
		return err
	}

	relay, err = common.GetRequiredEnv("BEHINDTHENAME_RELAY")
	if err != nil {
		return err
	}

	pool = nostr.NewSimplePool(ctx)

	if err := HandleBehindthename(ctx, NewBehindTheNameParams(
		nostrKey,
		pool,
		relay,
		continueFrom,
		l,
	)); err != nil {
		return err
	}

	return nil
}
