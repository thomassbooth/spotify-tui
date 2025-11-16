package view

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Subscriber interface {
	OnMessage(t MsgType, msg tea.Msg) tea.Cmd
}

type MessageBus struct {
	subscribers map[MsgType][]Subscriber
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscribers: make(map[MsgType][]Subscriber),
	}
}

func (m *MessageBus) Subscribe(msgType MsgType, sub Subscriber) {
	m.subscribers[msgType] = append(m.subscribers[msgType], sub)
}

func (m *MessageBus) Publish(t MsgType, msg tea.Msg) []tea.Cmd {

	var cmds []tea.Cmd

	for _, sub := range m.subscribers[t] {
		if cmd := sub.OnMessage(t, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}
