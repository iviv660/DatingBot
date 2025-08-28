package tg

import tb "gopkg.in/telebot.v4"

const (
	ActLike    = "like"
	ActDislike = "dislike"
	ActSleep   = "sleep"
)

func MenuKeyboard() *tb.ReplyMarkup {
	m := &tb.ReplyMarkup{ResizeKeyboard: true}
	btn1 := m.Text("1")
	btn2 := m.Text("2")
	btn3 := m.Text("3")
	m.Reply(m.Row(btn1, btn2, btn3))
	return m
}

func GenderKeyboard() *tb.ReplyMarkup {
	m := &tb.ReplyMarkup{ResizeKeyboard: true}
	male := m.Text("Парень")
	female := m.Text("Девушка")
	m.Reply(m.Row(male, female))
	return m
}

func BrowseKeyboard() *tb.ReplyMarkup {
	m := &tb.ReplyMarkup{ResizeKeyboard: true}
	like := m.Text("❤️")
	dislike := m.Text("👎")
	sleep := m.Text("💤")
	m.Reply(m.Row(like, dislike, sleep))
	return m
}
