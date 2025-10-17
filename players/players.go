package players

import (
	"context"

	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/google/uuid"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, pkn, pk, langPref string) (*Player, error) {
	player, exists := Get(pk)
	if !exists {
		player = &Player{
			Id:                 uuid.New().String(),
			Username:           utils.GenerateName(language.Languages[langPref]),
			Discriminator:      utils.GenerateDescriminator(),
			SshPubKeys:         map[string]string{pkn: pk},
			UpdateChan:         make(chan struct{}),
			LanguagePreference: langPref,
			connected:          true,
			onDisconnect:       []func(){},
			ctx:                ctx,
		}
	}

	player.UpdateChan = make(chan struct{})
	player.connected = true
	player.ctx = ctx
	players[player.Id] = player

	go func() {
		<-player.ctx.Done()
		player.connected = false
		for _, fn := range player.onDisconnect {
			fn()
		}
	}()

	return player, player.Save()
}

func Get(sshPubKey string) (*Player, bool) {
	pipeline := []map[string]any{
		{
			"$match": map[string]any{
				"$expr": map[string]any{
					"$gt": []any{
						map[string]any{
							"$size": map[string]any{
								"$filter": map[string]any{
									"input": map[string]any{"$objectToArray": "$ssh_pub_keys"},
									"cond":  map[string]any{"$eq": []any{"$$this.v", sshPubKey}},
								},
							},
						},
						0,
					},
				},
			},
		},
	}

	cursor, err := database.GetDB().Collection(database.CollectionPlayers).Aggregate(context.Background(), pipeline)
	if err == nil {
		defer cursor.Close(context.Background())
		if cursor.Next(context.Background()) {
			var player Player
			if err := cursor.Decode(&player); err == nil {
				return &player, true
			}
		}
	}

	return nil, false
}

func GetByName(username, discriminator string) (*Player, bool) {
	for _, player := range players {
		if player.Username == username && player.Discriminator == discriminator {
			return player, true
		}
	}

	pipeline := []map[string]any{
		{
			"$match": map[string]any{
				"username":      username,
				"discriminator": discriminator,
			},
		},
	}

	cursor, err := database.GetDB().Collection(database.CollectionPlayers).Aggregate(context.Background(), pipeline)
	if err == nil {
		defer cursor.Close(context.Background())
		if cursor.Next(context.Background()) {
			var player Player
			if err := cursor.Decode(&player); err == nil {
				return &player, true
			}
		}
	}

	return nil, false
}

func RemovePlayer(player *Player) {
	if _, exists := players[player.Id]; exists {
		close(player.UpdateChan)
		delete(players, player.Id)
	}
}

func GetPlayerCount() int {
	return len(players)
}

func GetConnectedPlayerCount() int {
	count := 0
	for _, player := range players {
		if player.connected {
			count++
		}
	}
	return count
}
