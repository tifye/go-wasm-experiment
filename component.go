//go:build js

package main

import (
	"log"
	"strings"
	"syscall/js"
)

type Component struct {
	name  string
	text  string
	comps []*Component
	attrs map[string]any
	el    *js.Value
}

func NewComponent(name string) *Component {
	return &Component{
		name: name,
	}
}

func (c *Component) SetAttribute(key string, val any) {
	if c.attrs == nil {
		c.attrs = make(map[string]any)
	}
	c.attrs[key] = val
}

type DOMRenderer struct {
	doc   js.Value
	body  js.Value
	comps []*Component
}

func NewDOMRenderer() *DOMRenderer {
	doc := js.Global().Get("document")
	body := doc.Get("body")

	return &DOMRenderer{
		doc:  doc,
		body: body,
	}
}

func (r *DOMRenderer) Render(comps ...*Component) {
	for _, c := range comps {
		frag, el := r.createElement(c)
		c.el = &el
		r.body.Call("appendChild", frag)
	}
}

func (r *DOMRenderer) createElement(c *Component) (frag, el js.Value) {
	frag = r.doc.Call("createDocumentFragment")
	el = r.doc.Call("createElement", c.name)

	if c.text != "" {
		el.Set("innerText", c.text)
	}

	for key, val := range c.attrs {
		if strings.HasPrefix(key, "on:") {
			valFunc, ok := val.(func())
			if !ok {
				log.Fatal("only func allowed for eventlisteners")
			}

			jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
				valFunc()
				return nil
			})
			// todo: handle release
			el.Call("addEventListener", strings.TrimPrefix(key, "on:"), jsFunc)

			continue
		}

		r.setAttribute(&el, key, val)
	}

	for _, cc := range c.comps {
		cfrag, cel := r.createElement(cc)
		cc.el = &cel
		el.Call("appendChild", cfrag)
	}

	frag.Call("appendChild", el)
	return frag, el
}

func (r *DOMRenderer) SetAttribute(c *Component, key string, val any) {
	c.SetAttribute(key, val)
	if c.el == nil {
		panic("nil el")
	}

	r.setAttribute(c.el, key, val)
}

func (r *DOMRenderer) setAttribute(el *js.Value, key string, val any) {
	switch key {
	case "innerText", "value":
		el.Set(key, val)
	case "type":
		el.Call("setAttribute", key, val)
	default:
		log.Fatalf("not implemented: %s", key)
	}
}

func (r *DOMRenderer) append(parent *Component, child *Component) {
	if parent.el == nil {
		panic("nil parent el")
	}

	if child.el != nil {
		parent.el.Call("appendChild", *child.el)
		return
	}

	frag, el := r.createElement(child)
	child.el = &el
	parent.el.Call("appendChild", frag)
}
