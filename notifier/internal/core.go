package internal

import (
	userpb "app/user/proto"
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type ReplyKind int

const (
	ReplyNone ReplyKind = iota
	ReplyMenu
	ReplyGender
	ReplyBrowse
)

type Output struct {
	Text        string
	PhotoString string
	Kind        ReplyKind
}

type state int

const (
	stIdle state = iota
	stAskName
	stAskAge
	stAskCity
	stAskGender
	stAskDesc
	stAskPhoto
	stMenu
	stBrowsing
)

type candidate struct {
	UserID     int64
	TelegramID int64
}

type session struct {
	State         state
	Draft         draftProfile
	Candidates    []candidate
	CurrentTarget *candidate
	UpdatedAt     time.Time
}

type draftProfile struct {
	Name        string
	Age         int32
	City        string
	Gender      string
	Description string
	PhotoString string
}

type Core struct {
	users UserClient
	match MatchClient

	mu       sync.RWMutex
	sessions map[int64]*session
}

func NewCore(users UserClient, match MatchClient) *Core {
	return &Core{
		users:    users,
		match:    match,
		sessions: make(map[int64]*session),
	}
}

func (c *Core) get(chatID int64) *session {
	c.mu.RLock()
	s := c.sessions[chatID]
	c.mu.RUnlock()
	if s == nil {
		s = &session{State: stIdle, UpdatedAt: time.Now()}
		c.mu.Lock()
		c.sessions[chatID] = s
		c.mu.Unlock()
	}
	return s
}

func (c *Core) reset(chatID int64) {
	c.mu.Lock()
	delete(c.sessions, chatID)
	c.mu.Unlock()
}

func (c *Core) OnStart(ctx context.Context, chatID int64) (Output, error) {
	u, err := c.users.GetByTelegramID(ctx, chatID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "user not found") {
			u = nil
		} else {
			log.Printf("core: GetByTelegramID: %v", err)
			return Output{Text: "Сервис недоступен. Попробуй позже."}, nil
		}
	}

	s := c.get(chatID)
	if u == nil {
		s.State = stAskName
		s.Draft = draftProfile{}
		s.UpdatedAt = time.Now()
		return Output{Text: "Привет! Давай создадим анкету.\nКак тебя зовут?"}, nil
	}

	s.State = stMenu
	s.UpdatedAt = time.Now()
	return Output{
		Text: "\nВыбери действие:\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏️",
		Kind: ReplyMenu,
	}, nil
}

func (c *Core) OnText(ctx context.Context, chatID int64, text string) (Output, error) {
	s := c.get(chatID)
	switch s.State {
	case stAskName:
		s.Draft.Name = text
		s.State = stAskAge
		s.UpdatedAt = time.Now()
		return Output{Text: "Сколько тебе лет?"}, nil

	case stAskAge:
		var age int32
		_, err := fmt.Sscanf(text, "%d", &age)
		if err != nil || age <= 0 {
			return Output{Text: "Возраст должен быть числом. Введи корректный возраст."}, nil
		}
		s.Draft.Age = age
		s.State = stAskCity
		s.UpdatedAt = time.Now()
		return Output{Text: "Где ты живёшь? Укажи город."}, nil

	case stAskCity:
		s.Draft.City = text
		s.State = stAskGender
		s.UpdatedAt = time.Now()
		return Output{Text: "Выбери пол:", Kind: ReplyGender}, nil

	case stAskGender:
		if text != "Парень" && text != "Девушка" {
			return Output{Text: "Пожалуйста, выбери кнопкой: Парень или Девушка.", Kind: ReplyGender}, nil
		}
		s.Draft.Gender = text
		s.State = stAskDesc
		s.UpdatedAt = time.Now()
		return Output{Text: "Кратко опиши себя (интересы, что ищешь)."}, nil

	case stAskDesc:
		s.Draft.Description = text
		s.State = stAskPhoto
		s.UpdatedAt = time.Now()
		return Output{Text: "Пришли фото для анкеты (одно изображение)."}, nil

	case stMenu:
		switch text {
		case "1":
			return c.startBrowsing(ctx, chatID)
		case "2":
			return c.showProfile(ctx, chatID)
		case "3":
			s.State = stAskName
			s.Draft = draftProfile{}
			s.UpdatedAt = time.Now()
			return Output{Text: "Ок, обновим анкету. Как тебя зовут?"}, nil
		default:
			return Output{Text: "Выбери пункт меню: 1 (смотреть), 2 (моя анкета), 3 (изменить).", Kind: ReplyMenu}, nil
		}

	case stBrowsing:
		return Output{Text: "Используй кнопки: ❤️ / 👎 / 💤", Kind: ReplyBrowse}, nil

	default:
		s.State = stAskName
		return Output{Text: "Давай начнём с начала. Как тебя зовут?"}, nil
	}
}

func (c *Core) OnPhoto(ctx context.Context, chatID int64, photo []byte) (Output, error) {
	s := c.get(chatID)
	if s.State != stAskPhoto {
		return Output{Text: "Фото сейчас не требуется. Используй меню."}, nil
	}

	existing, _ := c.users.GetByTelegramID(ctx, chatID)

	u := &userpb.User{
		TelegramId:  chatID,
		Username:    s.Draft.Name,
		Age:         s.Draft.Age,
		Gender:      s.Draft.Gender,
		Location:    s.Draft.City,
		Description: s.Draft.Description,
		IsVisible:   true,
	}

	var saved *userpb.User
	var err error

	if existing == nil {
		saved, err = c.users.Create(ctx, u)
		if err != nil {
			log.Printf("core: Create user: %v", err)
			return Output{Text: "Не удалось сохранить анкету. Попробуй ещё раз."}, nil
		}
	} else {
		u.Id = existing.GetId()
		saved, err = c.users.Update(ctx, u)
		if err != nil {
			log.Printf("core: Update user: %v", err)
			return Output{Text: "Не удалось обновить анкету. Попробуй ещё раз."}, nil
		}
	}

	if len(photo) > 0 {
		if u2, err := c.users.UpdatePhoto(ctx, saved.GetId(), bytes.NewReader(photo)); err == nil && u2 != nil {
			saved = u2
		}
	}

	s.State = stMenu
	s.Draft = draftProfile{}
	s.UpdatedAt = time.Now()

	return Output{
		Text: "Анкета сохранена! Что дальше?\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏️",
		Kind: ReplyMenu,
	}, nil
}

func (c *Core) OnCallback(ctx context.Context, chatID int64, action string) (Output, error) {
	s := c.get(chatID)

	if s.State == stAskGender && (action == "gender_male" || action == "gender_female") {
		if action == "gender_male" {
			s.Draft.Gender = "Парень"
		} else {
			s.Draft.Gender = "Девушка"
		}
		s.State = stAskDesc
		s.UpdatedAt = time.Now()
		return Output{Text: "Кратко опиши себя."}, nil
	}

	if s.State != stBrowsing {
		return Output{Text: "Действие сейчас недоступно. Используй меню."}, nil
	}

	switch action {
	case "like", "dislike":
		if s.CurrentTarget == nil {
			s.State = stMenu
			s.UpdatedAt = time.Now()
			return Output{Text: "Кандидатов больше нет.\nЧто дальше?\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏", Kind: ReplyMenu}, nil
		}

		me, err := c.users.GetByTelegramID(ctx, chatID)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "user not found") {
				return Output{Text: "Сначала зарегистрируй анкету: /start"}, nil
			}
			log.Printf("core: GetByTelegramID: %v", err)
			return Output{Text: "Сервис недоступен, попробуй позже."}, nil
		}

		isLike := action == "like"
		if err := c.match.Like(ctx, me.GetId(), s.CurrentTarget.UserID, isLike); err != nil {
			log.Printf("core: Like(%v): %v", isLike, err)
		}

		if isLike {
			if ok, err := c.match.Match(ctx, me.GetId(), s.CurrentTarget.UserID); err == nil && ok {
				out, err := c.nextCandidate(ctx, chatID)
				if err == nil {
					if out.Text != "" {
						out.Text = "🎉 У тебя совпадение!\n\n" + out.Text
					} else {
						out.Text = "🎉 У тебя совпадение!"
					}
					return out, nil
				}
			}
		}
		return c.nextCandidate(ctx, chatID)

	case "sleep":
		s.State = stMenu
		s.UpdatedAt = time.Now()
		return Output{
			Text: "Ок, вернулись в меню.\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏️",
			Kind: ReplyMenu,
		}, nil
	}

	return Output{Text: "Неизвестное действие."}, nil
}

func (c *Core) startBrowsing(ctx context.Context, chatID int64) (Output, error) {
	u, err := c.users.GetByTelegramID(ctx, chatID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "user not found") {
			s := c.get(chatID)
			s.State = stAskName
			return Output{Text: "Похоже, анкеты нет. Давай создадим! Как тебя зовут?"}, nil
		}
		return Output{Text: "Сервис недоступен, попробуй позже."}, nil
	}

	cands, err := c.match.GetCandidates(ctx, u.GetId())
	if err != nil {
		return Output{Text: "Не удалось получить кандидатов. Попробуй позже."}, nil
	}
	if len(cands) == 0 {
		return Output{Text: "Пока нет подходящих анкет.\nЧто дальше?\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏", Kind: ReplyMenu}, nil
	}

	s := c.get(chatID)
	s.Candidates = s.Candidates[:0]
	for _, cand := range cands {
		s.Candidates = append(s.Candidates, candidate{
			UserID:     cand.GetId(),
			TelegramID: cand.GetTelegramId(),
		})
	}
	s.State = stBrowsing
	s.UpdatedAt = time.Now()

	return c.nextCandidate(ctx, chatID)
}

func (c *Core) nextCandidate(ctx context.Context, chatID int64) (Output, error) {
	s := c.get(chatID)
	if len(s.Candidates) == 0 {
		s.State = stMenu
		return Output{Text: "Анкеты закончились. Возвращаемся в меню.\nЧто дальше?\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏", Kind: ReplyMenu}, nil
	}

	last := s.Candidates[len(s.Candidates)-1]
	s.Candidates = s.Candidates[:len(s.Candidates)-1]
	s.CurrentTarget = &last
	s.UpdatedAt = time.Now()

	target, err := c.users.GetByID(ctx, last.UserID)
	if err != nil || target == nil {
		target, _ = c.users.GetByTelegramID(ctx, last.TelegramID)
	}
	if target == nil {
		return Output{Text: "Не удалось получить профиль кандидата. Пробуем следующего…"}, nil
	}

	caption := fmt.Sprintf("%s, %d, %s\n%s",
		target.GetUsername(), target.GetAge(), target.GetLocation(), target.GetDescription())

	return Output{
		Text:        caption,
		Kind:        ReplyBrowse,
		PhotoString: target.GetPhotoUrl(),
	}, nil
}

func (c *Core) showProfile(ctx context.Context, chatID int64) (Output, error) {
	u, err := c.users.GetByTelegramID(ctx, chatID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "user not found") {
			return Output{Text: "Анкета не найдена. Давай создадим! Как тебя зовут?"}, nil
		}
		return Output{Text: "Сервис недоступен. Попробуй позже."}, nil
	}
	caption := fmt.Sprintf("Твоя анкета:\n%s, %d, %s\n%s",
		u.GetUsername(), u.GetAge(), u.GetLocation(), u.GetDescription())

	return Output{
		Text:        caption,
		Kind:        ReplyMenu,
		PhotoString: u.GetPhotoUrl(),
	}, nil
}
