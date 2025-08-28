package internal

import (
	userpb "app/user/proto"
	"context"
	"fmt"
	"io"
	"log"
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

type session struct {
	State         state
	Draft         draftProfile
	Candidates    []int64
	CurrentTarget int64
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
		log.Printf("core: GetByTelegramID: %v", err)
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
		Text: fmt.Sprintf(
			"%s, %d, %s — %s\n\nВыбери действие:\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏️",
			u.GetUsername(), u.GetAge(), u.GetLocation(), u.GetDescription(),
		),
		Kind:        ReplyMenu,
		PhotoString: u.GetPhotoUrl(),
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
			return Output{Text: "Укажи Свой пол", Kind: ReplyGender}, nil
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
		return Output{Text: "Используй кнопки: ❤ / 👎 / 💤", Kind: ReplyBrowse}, nil
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
	// s.Draft.PhotoBytes = photo // ← УДАЛИТЬ, такого поля больше нет

	u := &userpb.User{
		TelegramId:  chatID,
		Username:    s.Draft.Name,
		Age:         s.Draft.Age,
		Gender:      s.Draft.Gender,
		Location:    s.Draft.City,
		Description: s.Draft.Description,
		IsVisible:   true,
	}
	created, err := c.users.Create(ctx, u)
	if err != nil {
		log.Printf("core: Create user: %v", err)
		return Output{Text: "Не удалось сохранить анкету. Попробуй ещё раз."}, nil
	}

	updated := created
	if len(photo) > 0 {
		if u2, err := c.users.UpdatePhoto(ctx, created.GetId(), bytesReader(photo)); err != nil {
			log.Printf("core: UpdatePhoto: %v", err)
		} else if u2 != nil {
			updated = u2
		}
	}

	s.State = stMenu
	s.Draft = draftProfile{}
	s.UpdatedAt = time.Now()

	return Output{
		Text: fmt.Sprintf(
			"%s, %d, %s — %s\n\nАнкета сохранена! Что дальше?\n1. Смотреть анкеты 🚀\n2. Моя анкета 📱\n3. Изменить анкету ✏️",
			updated.GetUsername(), updated.GetAge(), updated.GetLocation(), updated.GetDescription(),
		),
		Kind:        ReplyMenu,
		PhotoString: updated.GetPhotoUrl(),
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
	if s.State == stBrowsing {
		switch action {
		case "like":
			if s.CurrentTarget == 0 {
				return Output{Text: "Кандидатов больше нет.", Kind: ReplyMenu}, nil
			}
			if err := c.match.Like(ctx, chatID, s.CurrentTarget, true); err != nil {
				log.Printf("core: Like: %v", err)
			}
			return c.nextCandidate(ctx, chatID)
		case "dislike":
			if err := c.match.Like(ctx, chatID, s.CurrentTarget, false); err != nil {
				log.Printf("core: Dislike: %v", err)
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
	}
	return Output{Text: "Неизвестное действие."}, nil
}

func (c *Core) startBrowsing(ctx context.Context, chatID int64) (Output, error) {
	u, err := c.users.GetByTelegramID(ctx, chatID)
	if err != nil {
		return Output{}, err
	}
	if u == nil {
		s := c.get(chatID)
		s.State = stAskName
		return Output{Text: "Похоже, анкеты нет. Как тебя зовут?"}, nil
	}
	cands, err := c.match.GetCandidates(ctx, u.GetId())
	if err != nil {
		return Output{}, err
	}
	if len(cands) == 0 {
		return Output{Text: "Пока нет подходящих анкет. Попробуй позже.", Kind: ReplyMenu}, nil
	}
	s := c.get(chatID)
	s.Candidates = s.Candidates[:0]
	for _, cand := range cands {
		s.Candidates = append(s.Candidates, cand.GetTelegramId())
	}
	s.State = stBrowsing
	s.UpdatedAt = time.Now()
	return c.nextCandidate(ctx, chatID)
}

func (c *Core) nextCandidate(ctx context.Context, chatID int64) (Output, error) {
	s := c.get(chatID)
	if len(s.Candidates) == 0 {
		s.State = stMenu
		return Output{Text: "Анкеты закончились. Возвращаемся в меню.", Kind: ReplyMenu}, nil
	}
	last := s.Candidates[len(s.Candidates)-1]
	s.Candidates = s.Candidates[:len(s.Candidates)-1]
	s.CurrentTarget = last
	s.UpdatedAt = time.Now()

	target, err := c.users.GetByTelegramID(ctx, last)
	if err != nil || target == nil {
		return Output{Text: "Не удалось получить профиль кандидата. Пробуем следующего…"}, nil
	}

	caption := fmt.Sprintf("%s, %d, %s — %s",
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
		return Output{}, err
	}
	if u == nil {
		return Output{Text: "Анкета не найдена. Давай создадим! Как тебя зовут?"}, nil
	}
	return Output{
		Text: fmt.Sprintf("Твоя анкета: %s, %d, %s — %s",
			u.GetUsername(), u.GetAge(), u.GetLocation(), u.GetDescription()),
		Kind:        ReplyMenu,
		PhotoString: u.GetPhotoUrl(),
	}, nil
}

type byteReader struct{ b []byte }

func bytesReader(b []byte) *byteReader { return &byteReader{b: b} }

func (r *byteReader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}
