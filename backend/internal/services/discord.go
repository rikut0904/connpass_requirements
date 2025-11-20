package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

var ErrMissingAccess = errors.New("discord: missing access")

// DiscordService はdiscordgoラッパー。
type DiscordService struct {
	session *discordgo.Session
}

func NewDiscordService(botToken string) (*DiscordService, error) {
	sess, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	return &DiscordService{session: sess}, nil
}

func (s *DiscordService) Open() error {
	return s.session.Open()
}

func (s *DiscordService) Close() {
	s.session.Close()
}

// Session は内部のdiscordgo.Sessionを返す。
func (s *DiscordService) Session() *discordgo.Session {
	return s.session
}

func (s *DiscordService) SendMessage(ctx context.Context, channelID, message string) error {
	_, err := s.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{Content: message})
	if err != nil {
		return fmt.Errorf("send discord message: %w", err)
	}
	return nil
}

func (s *DiscordService) CreateTextChannel(ctx context.Context, guildID, name, parentID string) (*discordgo.Channel, error) {
	data := discordgo.GuildChannelCreateData{
		Name: name,
		Type: discordgo.ChannelTypeGuildText,
	}
	if parentID != "" {
		data.ParentID = parentID
	}
	channel, err := s.session.GuildChannelCreateComplex(guildID, data)
	if err != nil {
		if isMissingAccessErr(err) {
			return nil, ErrMissingAccess
		}
		return nil, fmt.Errorf("create discord channel: %w", err)
	}
	return channel, nil
}

func (s *DiscordService) CreateCategory(ctx context.Context, guildID, name string) (*discordgo.Channel, error) {
	category, err := s.session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name: name,
		Type: discordgo.ChannelTypeGuildCategory,
	})
	if err != nil {
		if isMissingAccessErr(err) {
			return nil, ErrMissingAccess
		}
		return nil, fmt.Errorf("create discord category: %w", err)
	}
	return category, nil
}

func (s *DiscordService) ListTextChannels(ctx context.Context, guildID string) ([]*discordgo.Channel, error) {
	text, _, err := s.ListTextChannelsWithCategories(ctx, guildID)
	return text, err
}

func (s *DiscordService) ListTextChannelsWithCategories(ctx context.Context, guildID string) ([]*discordgo.Channel, []*discordgo.Channel, error) {
	channels, err := s.session.GuildChannels(guildID)
	if err != nil {
		if isMissingAccessErr(err) {
			return nil, nil, ErrMissingAccess
		}
		return nil, nil, fmt.Errorf("list guild channels: %w", err)
	}

	var textChannels []*discordgo.Channel
	var categories []*discordgo.Channel
	for _, ch := range channels {
		switch ch.Type {
		case discordgo.ChannelTypeGuildText:
			textChannels = append(textChannels, ch)
		case discordgo.ChannelTypeGuildCategory:
			categories = append(categories, ch)
		}
	}
	return textChannels, categories, nil
}

func (s *DiscordService) IsBotInGuild(ctx context.Context, guildID string) (bool, error) {
	_, err := s.session.Guild(guildID)
	if err == nil {
		return true, nil
	}
	if isMissingAccessErr(err) {
		return false, nil
	}
	return false, fmt.Errorf("get guild %s: %w", guildID, err)
}

func isMissingAccessErr(err error) bool {
	var restErr *discordgo.RESTError
	if !errors.As(err, &restErr) {
		return false
	}
	if restErr.Response != nil && restErr.Response.StatusCode == http.StatusForbidden {
		return true
	}
	if restErr.Message != nil {
		switch restErr.Message.Code {
		case discordgo.ErrCodeMissingAccess,
			discordgo.ErrCodeUnknownGuild,
			discordgo.ErrCodeUnknownChannel:
			return true
		}
	}
	return false
}
