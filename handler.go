package whatsapp

import (
	"fmt"
	"os"

	"github.com/gleandroj/go-whatsapp/binary"
	"github.com/gleandroj/go-whatsapp/binary/proto"
)

/*
The Handler interface is the minimal interface that needs to be implemented
to be accepted as a valid handler for our dispatching system.
The minimal handler is used to dispatch error messages. These errors occur on unexpected behavior by the websocket
connection or if we are unable to handle or interpret an incoming message. Error produced by user actions are not
dispatched through this handler. They are returned as an error on the specific function call.
*/
type Handler interface {
	HandleError(err error)
}

/*
The TextMessageHandler interface needs to be implemented to receive text messages dispatched by the dispatcher.
*/
type TextMessageHandler interface {
	Handler
	HandleTextMessage(message TextMessage)
}

/*
The ImageMessageHandler interface needs to be implemented to receive image messages dispatched by the dispatcher.
*/
type ImageMessageHandler interface {
	Handler
	HandleImageMessage(message ImageMessage)
}

/*
The VideoMessageHandler interface needs to be implemented to receive video messages dispatched by the dispatcher.
*/
type VideoMessageHandler interface {
	Handler
	HandleVideoMessage(message VideoMessage)
}

/*
The AudioMessageHandler interface needs to be implemented to receive audio messages dispatched by the dispatcher.
*/
type AudioMessageHandler interface {
	Handler
	HandleAudioMessage(message AudioMessage)
}

/*
The DocumentMessageHandler interface needs to be implemented to receive document messages dispatched by the dispatcher.
*/
type DocumentMessageHandler interface {
	Handler
	HandleDocumentMessage(message DocumentMessage)
}

/*
The LocationMessageHandler interface needs to be implemented to receive location messages dispatched by the dispatcher.
*/
type LocationMessageHandler interface {
	Handler
	HandleLocationMessage(message LocationMessage)
}

/*
The LiveLocationMessageHandler interface needs to be implemented to receive live location messages dispatched by the dispatcher.
*/
type LiveLocationMessageHandler interface {
	Handler
	HandleLiveLocationMessage(message LiveLocationMessage)
}

/*
The JsonMessageHandler interface needs to be implemented to receive json messages dispatched by the dispatcher.
These json messages contain status updates of every kind sent by WhatsAppWeb servers. WhatsAppWeb uses these messages
to built a Store, which is used to save these "secondary" information. These messages may contain
presence (available, last see) information, or just the battery status of your phone.
*/
type JsonMessageHandler interface {
	Handler
	HandleJsonMessage(message string)
}

/**
The RawMessageHandler interface needs to be implemented to receive raw messages dispatched by the dispatcher.
Raw messages are the raw protobuf structs instead of the easy-to-use structs in TextMessageHandler, ImageMessageHandler, etc..
*/
type RawMessageHandler interface {
	Handler
	HandleRawMessage(message *proto.WebMessageInfo)
}

/*
AddHandler adds an handler to the list of handler that receive dispatched messages.
The provided handler must at least implement the Handler interface. Additionally implemented
handlers(TextMessageHandler, ImageMessageHandler) are optional. At runtime it is checked if they are implemented
and they are called if so and needed.
*/
func (wac *Conn) AddHandler(handler Handler) {
	wac.handler = append(wac.handler, handler)
}

func handleMessage(message interface{}, handlers []Handler) {
	switch m := message.(type) {
	case error:
		for _, h := range handlers {
			go h.HandleError(m)
		}
	case string:
		for _, h := range handlers {
			if x, ok := h.(JsonMessageHandler); ok {
				go x.HandleJsonMessage(m)
			}
		}
	case TextMessage:
		for _, h := range handlers {
			if x, ok := h.(TextMessageHandler); ok {
				go x.HandleTextMessage(m)
			}
		}
	case ImageMessage:
		for _, h := range handlers {
			if x, ok := h.(ImageMessageHandler); ok {
				go x.HandleImageMessage(m)
			}
		}
	case VideoMessage:
		for _, h := range handlers {
			if x, ok := h.(VideoMessageHandler); ok {
				go x.HandleVideoMessage(m)
			}
		}
	case AudioMessage:
		for _, h := range handlers {
			if x, ok := h.(AudioMessageHandler); ok {
				go x.HandleAudioMessage(m)
			}
		}
	case DocumentMessage:
		for _, h := range handlers {
			if x, ok := h.(DocumentMessageHandler); ok {
				go x.HandleDocumentMessage(m)
			}
		}
	case LocationMessage:
		for _, h := range handlers {
			if x, ok := h.(LocationMessageHandler); ok {
				go x.HandleLocationMessage(m)
			}
		}
	case LiveLocationMessage:
		for _, h := range handlers {
			if x, ok := h.(LiveLocationMessageHandler); ok {
				go x.HandleLiveLocationMessage(m)
			}
		}
	case *proto.WebMessageInfo:
		for _, h := range handlers {
			if x, ok := h.(RawMessageHandler); ok {
				go x.HandleRawMessage(m)
			}
		}
	}

}

func (wac *Conn) handle(message interface{}) {
	handleMessage(message, wac.handler)
}

func (wac *Conn) dispatch(msg interface{}) {
	if msg == nil {
		return
	}

	switch message := msg.(type) {
	case *binary.Node:
		if message.Description == "action" {
			if con, ok := message.Content.([]interface{}); ok {
				for a := range con {
					if v, ok := con[a].(*proto.WebMessageInfo); ok {
						wac.handle(v)
						wac.handle(parseProtoMessage(v))
					}
				}
			}
		} else if message.Description == "response" && message.Attributes["type"] == "contacts" {
			wac.updateContacts(message.Content)
		}
	case error:
		wac.handle(message)
	case string:
		wac.handle(message)
	default:
		fmt.Fprintf(os.Stderr, "unknown type in dipatcher chan: %T", msg)
	}
}
