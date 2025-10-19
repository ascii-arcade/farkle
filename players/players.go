package players

import (
	"context"
	"maps"
	"time"

	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/google/uuid"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, pkn, pk, langPref string) (*Player, error) {
	player := &Player{
		Id:                 uuid.New().String(),
		Username:           utils.GenerateName(language.Languages[langPref]),
		Discriminator:      utils.GenerateDescriminator(),
		SshPubKeys:         map[string]string{pkn: pk},
		LanguagePreference: langPref,

		onDisconnect: []func(){},
		ctx:          ctx,
	}
	return player, player.Save()
}

func (p *Player) Connect() {
	p.updateChan = make(chan struct{})
	p.connected = true
	players[p.Id] = p

	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return
			default:
			}

			p.LastConnectedAt = utils.ToPointer(time.Now())
			p.connected = true
			_ = p.Save()
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		<-p.ctx.Done()
		p.connected = false
		for _, fn := range p.onDisconnect {
			fn()
		}
	}()
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
		close(player.updateChan)
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

func DeletePlayer(player *Player) error {
	RemovePlayer(player)
	_, err := database.GetDB().Collection(database.CollectionPlayers).DeleteOne(context.Background(), map[string]any{
		"id": player.Id,
	})
	return err
}

func Merge(target, source *Player) error {
	maps.Copy(target.SshPubKeys, source.SshPubKeys)
	if err := target.Save(); err != nil {
		return err
	}

	return DeletePlayer(source)
}
