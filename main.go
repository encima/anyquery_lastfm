package main

import (
	"github.com/julien040/anyquery/rpc"
)

func main() {
	plugin := rpc.NewPlugin(
		recentTracksCreator,
		topArtistsCreator,
		topTracksCreator,
		topAlbumsCreator,
	)
	plugin.Serve()
}
