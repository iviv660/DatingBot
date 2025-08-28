package tg

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"time"

	"app/notifier/internal"

	tb "gopkg.in/telebot.v4"
)

type Handler struct {
	bot  *tb.Bot
	core *internal.Core
}

func NewHandler(bot *tb.Bot, core *internal.Core) *Handler {
	return &Handler{bot: bot, core: core}
}

func (h *Handler) Register() {
	h.bot.Handle("/start", h.onStart)
	h.bot.Handle(tb.OnText, h.onText)
	h.bot.Handle(tb.OnPhoto, h.onPhoto) // ← убрали опечатку
}

const (
	tmoShort     = 10 * time.Second
	tmoText      = 15 * time.Second
	tmoPhoto     = 20 * time.Second
	maxPhotoSize = 8 << 20 // 8MB
)

func (h *Handler) onStart(c tb.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), tmoShort)
	defer cancel()

	out, err := h.core.OnStart(ctx, c.Sender().ID)
	if err != nil {
		log.Printf("core.OnStart: %v", err)
		return c.Send("Что-то пошло не так. Попробуй ещё раз.")
	}
	return h.render(c, out)
}

func (h *Handler) onText(c tb.Context) error {
	txt := c.Text()

	// Лайк/дизлайк/пауза — теперь приходят как текстовые кнопки
	if txt == "❤️" || txt == "👎" || txt == "💤" {
		var action string
		switch txt {
		case "❤️":
			action = "like"
		case "👎":
			action = "dislike"
		case "💤":
			action = "sleep"
		}
		ctx, cancel := context.WithTimeout(context.Background(), tmoShort)
		defer cancel()
		out, err := h.core.OnCallback(ctx, c.Sender().ID, action)
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			log.Printf("core.OnCallback(%s): %v", action, err)
			return c.Send("Действие не удалось. Попробуй ещё раз.")
		}
		return h.render(c, out)
	}

	// Всё остальное — обычный текстовый поток (меню 1/2/3, «Парень/Девушка», и т.д.)
	ctx, cancel := context.WithTimeout(context.Background(), tmoText)
	defer cancel()
	out, err := h.core.OnText(ctx, c.Sender().ID, txt)
	if err != nil {
		log.Printf("core.OnText: %v", err)
		return c.Send("Не понял сообщение. Попробуй ещё раз.")
	}
	return h.render(c, out)
}

func (h *Handler) onPhoto(c tb.Context) error {
	p := c.Message().Photo
	if p == nil {
		return c.Send("Пришли, пожалуйста, фото изображением, не файлом.")
	}

	file := p.MediaFile()
	rc, err := h.bot.File(file) // ВАЖНО: передаём адрес
	if err != nil {
		log.Printf("tg.getFile: %v", err)
		return c.Send("Не удалось получить фото, попробуй ещё раз.")
	}
	defer func() {
		if closer, ok := rc.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	lr := &io.LimitedReader{R: rc, N: maxPhotoSize}
	data, err := io.ReadAll(lr)
	if err != nil {
		log.Printf("tg.readPhoto: %v", err)
		return c.Send("Не удалось прочитать фото, попробуй ещё раз.")
	}
	if lr.N <= 0 {
		return c.Send("Фото слишком большое. Отправь файл поменьше (до 8MB).")
	}

	ctx, cancel := context.WithTimeout(context.Background(), tmoPhoto)
	defer cancel()

	out, err := h.core.OnPhoto(ctx, c.Sender().ID, data)
	if err != nil {
		log.Printf("core.OnPhoto: %v", err)
		return c.Send("Не удалось обработать фото, попробуй ещё раз.")
	}
	return h.render(c, out)
}

func (h *Handler) render(c tb.Context, out internal.Output) error {
	// Если есть картинка — отправляем как фото с подписью
	if out.PhotoString != "" {
		var photo *tb.Photo
		if strings.HasPrefix(out.PhotoString, "file_id:") {
			id := strings.TrimPrefix(out.PhotoString, "file_id:")
			photo = &tb.Photo{File: tb.File{FileID: id}, Caption: out.Text}
		} else if strings.HasPrefix(out.PhotoString, "http://") || strings.HasPrefix(out.PhotoString, "https://") {
			photo = &tb.Photo{File: tb.FromURL(out.PhotoString), Caption: out.Text}
		}
		if photo != nil {
			return c.Send(photo, keyboardByKind(out.Kind))
		}
		// неизвестный формат — упадём в текст
	}

	// Иначе — обычный текст
	return c.Send(out.Text, keyboardByKind(out.Kind))
}

func keyboardByKind(k internal.ReplyKind) *tb.ReplyMarkup {
	switch k {
	case internal.ReplyMenu:
		return MenuKeyboard()
	case internal.ReplyGender:
		return GenderKeyboard()
	case internal.ReplyBrowse:
		return BrowseKeyboard()
	default:
		return nil
	}
}
