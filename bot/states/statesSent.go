package states

import (
	tb "gopkg.in/telebot.v3"
	"sync"
)

// TODO: Do this shit on redis if you have time
//  probably didn't have time for this

type sentStruct struct {
	SentMessagesId []int
}

type sentMap struct {
	Mx  sync.RWMutex
	Map map[int64]sentStruct
}

var Sent *sentMap

func MakeSentMap() {
	Sent = &sentMap{Map: make(map[int64]sentStruct), Mx: sync.RWMutex{}}
}

func (sent *sentStruct) Delete(c tb.Context) {
	if len(sent.SentMessagesId) > 0 {
		for _, msgId := range sent.SentMessagesId {
			_ = c.Bot().Delete(&tb.Message{
				ID:   msgId,
				Chat: c.Chat(),
			})
		}

		Sent.Mx.RLock()
		delete(Sent.Map, c.Sender().ID)
		Sent.Mx.RUnlock()
	}
}
