package hub

import (
	"errors"
	"fmt"
)

type PluginRegistryConfig struct{}

type PluginRegistry[T any] struct {
	broadcastMessagePlugins map[WsMessageType][]BroadcastMessagePlugin[T]
	clientJoinPlugins       []ClientJoinPlugin[T]
	clientLeavePlugins      []ClientLeavePlugin[T]
}

func NewPluginRegistry[T any](cfg *PluginRegistryConfig) *PluginRegistry[T] {
	return &PluginRegistry[T]{
		broadcastMessagePlugins: make(map[WsMessageType][]BroadcastMessagePlugin[T]),
		clientJoinPlugins:       make([]ClientJoinPlugin[T], 0),
		clientLeavePlugins:      make([]ClientLeavePlugin[T], 0),
	}
}

func (pr *PluginRegistry[T]) RegisterClientJoinPlugin(plugin ClientJoinPlugin[T]) {
	pr.clientJoinPlugins = append(pr.clientJoinPlugins, plugin)
}

func (pr *PluginRegistry[T]) RegisterBroadcastMessagePlugin(plugin BroadcastMessagePlugin[T]) {
	messageType := plugin.MessageType()
	pr.broadcastMessagePlugins[messageType] = append(pr.broadcastMessagePlugins[messageType], plugin)
}

func (pr *PluginRegistry[T]) RegisterClientLeavePlugin(plugin ClientLeavePlugin[T]) {
	pr.clientLeavePlugins = append(pr.clientLeavePlugins, plugin)
}

func (pr *PluginRegistry[T]) handleClientJoin(room *ActiveRoom[T], client *Client[T]) error {
	var joinedErr error
	for _, plugin := range pr.clientJoinPlugins {
		if err := plugin.HandleClientJoin(room, client); err != nil {
			joinedErr = errors.Join(joinedErr, err)
		}
	}

	return joinedErr
}

func (pr *PluginRegistry[T]) handleBroadcastMessage(room *ActiveRoom[T], msg BroadcastMessage[T]) error {
	plugins, ok := pr.broadcastMessagePlugins[msg.wsMessage.Type]
	if !ok {
		return fmt.Errorf("no plugins registered for message type: %s", string(msg.wsMessage.Type))
	}

	var joinedErr error
	for _, plugin := range plugins {
		if err := plugin.HandleBroadcastMessage(room, msg); err != nil {
			joinedErr = errors.Join(joinedErr, err)
		}
	}

	return joinedErr
}

func (pr *PluginRegistry[T]) handleClientLeave(room *ActiveRoom[T], client *Client[T]) error {
	var joinedErr error
	for _, plugin := range pr.clientLeavePlugins {
		if err := plugin.HandleClientLeave(room, client); err != nil {
			joinedErr = errors.Join(joinedErr, err)
		}
	}

	return joinedErr
}
