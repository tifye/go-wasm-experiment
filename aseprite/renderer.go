package aseprite

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"tifye/go-wasm-test/render"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	idCounter atomic.Uint32
)

type asepriteEvent struct {
	Id    uint32 `json:"id"`
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
}

type asepriteMessage struct {
	Method string         `json:"method"`
	Id     uint32         `json:"id"`
	Type   string         `json:"type"`
	Data   map[string]any `json:"data"`
}

type asepriteActionMessage struct {
	asepriteMessage
	Action string `json:"action"`
}

type AsepriteComponent struct {
	name     string
	id       uint32
	comps    []render.Component
	attrs    map[string]any
	renderer *AsepriteRenderer
}

func (c *AsepriteComponent) SetAttribute(key string, val any) {
	if c.attrs == nil {
		c.attrs = make(map[string]any)
	}
	c.attrs[key] = val

	if c.renderer != nil {
		c.renderer.setAttribute(c, key, val)
	}
}

func (c *AsepriteComponent) Name() string {
	return c.name
}

func (c *AsepriteComponent) Children() []render.Component {
	return c.comps
}

func (c *AsepriteComponent) Attributes() map[string]any {
	return c.attrs
}

type AsepriteRenderer struct {
	writeChan  chan []byte
	notifyConn chan struct{}
	events     map[string]func()
}

func NewAsepriteRenderer() *AsepriteRenderer {
	renderer := &AsepriteRenderer{
		writeChan:  make(chan []byte),
		notifyConn: make(chan struct{}),
		events:     make(map[string]func()),
	}

	http.HandleFunc("/", renderer.handleConnection)

	go func() {
		log.Println("running")
		if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
			panic(err)
		}
	}()

	<-renderer.notifyConn
	return renderer
}

func (ar *AsepriteRenderer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	log.Println("connected")
	ar.notifyConn <- struct{}{}

	go func() {
		for msg := range ar.writeChan {
			log.Println("writing", string(msg))
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println(err)
				close(ar.writeChan)
				break
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		fmt.Println(string(msg))

		var ev asepriteEvent
		if err := json.Unmarshal(msg, &ev); err != nil {
			log.Println(err)
			continue
		}

		evFunc, ok := ar.events[fmt.Sprintf("%d:%s", ev.Id, ev.Event)]
		if ok {
			evFunc()
		}
	}
}

func (r *AsepriteRenderer) NewComponent(name string) *AsepriteComponent {
	return &AsepriteComponent{
		id:   idCounter.Add(1),
		name: name,
	}
}

func (r *AsepriteRenderer) Render(comps ...render.Component) {
	for _, c := range comps {
		aspC, ok := c.(*AsepriteComponent)
		if !ok {
			panic("invalid comp type")
		}

		r.createElement(aspC)
	}
}

func (r *AsepriteRenderer) createElement(c *AsepriteComponent) {
	switch c.name {
	case "dialog":
		msg := asepriteMessage{
			Method: "create",
			Id:     c.id,
			Type:   c.name,
			Data: map[string]any{
				"title":      c.Attributes()["title"],
				"notitlebar": false,
			},
		}
		msgBytes, _ := json.Marshal(msg)
		r.writeChan <- msgBytes

		showMsg := asepriteActionMessage{
			Action: "show",
			asepriteMessage: asepriteMessage{
				Method: "action",
				Id:     c.id,
				Type:   c.name,
			},
		}
		msgBytes, _ = json.Marshal(showMsg)
		r.writeChan <- msgBytes
	case "button":
		if val, ok := c.Attributes()["on:click"]; ok {
			valFunc, ok := val.(func())
			if !ok {
				log.Fatal("only func allowed for eventlisteners")
			}

			r.events[fmt.Sprintf("%d:%s", c.id, "click")] = valFunc
		}

		msg := asepriteMessage{
			Method: "create",
			Id:     c.id,
			Type:   c.name,
			Data: map[string]any{
				"dialogId": 1,
				"text":     c.Attributes()["text"],
				"label":    c.Attributes()["label"],
			},
		}
		msgBytes, _ := json.Marshal(msg)
		r.writeChan <- msgBytes

	default:
		panic("not implemented")
	}

	for _, cc := range c.comps {
		aspCC, ok := cc.(*AsepriteComponent)
		if !ok {
			panic("invalid comp type")
		}

		r.createElement(aspCC)
	}

	c.renderer = r
}

func (r *AsepriteRenderer) setAttribute(c *AsepriteComponent, key string, val any) {
	switch c.name {
	case "button":
		msg := asepriteMessage{
			Method: "update",
			Id:     c.id,
			Type:   c.name,
			Data: map[string]any{
				key: val,
			},
		}
		msgBytes, _ := json.Marshal(msg)
		r.writeChan <- msgBytes
	default:
		// panic("not implemented")
	}
}

func (r *AsepriteRenderer) Append(parent render.Component, child render.Component) {
	p, ok := parent.(*AsepriteComponent)
	if !ok {
		panic("invalid parent comp type")
	}

	c, ok := child.(*AsepriteComponent)
	if !ok {
		panic("invalid parent comp type")
	}

	p.comps = append(p.comps, c)
}
