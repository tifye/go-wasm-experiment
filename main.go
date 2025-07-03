//go:build js

package main

import (
	"syscall/js"
)

func main() {
	counter := &Signal[int64]{
		value: 0,
		deps:  make([]Dependency[int64], 0),
	}
	items := &Signal[[]string]{
		value: make([]string, 0),
		deps:  make([]Dependency[[]string], 0),
	}

	renderer := NewDOMRenderer()

	Incrementer(renderer, IncrementerProps{
		counter: counter,
	})

	MyAmazingList(renderer, MyAmazingListProps{
		counter: counter,
		items:   items,
	})

	// Next steps:
	// 1. Look at how other frameworks update the DOM
	// 		- Look at the ones using Div()-like functions
	// 2. Imagine what the renderer would look like
	// 		- How do I create elements and create/destory children? I imagine its similar to the above
	// 3. Can I create the above program with Renderer calls and Signals?

	select {}
}

type jsFunc func(this js.Value, args []js.Value) any

type IncrementerProps struct {
	counter *Signal[int64]
}

func Incrementer(renderer *DOMRenderer, props IncrementerProps) {
	counter := props.counter

	incBtn := NewComponent("button")
	incBtn.text = "increment"
	incBtn.SetAttribute("on:click", func() {
		counter.Set(counter.Value() + 1)
	})
	renderer.Render(incBtn)

	decBtn := NewComponent("button")
	decBtn.text = "decrement"
	decBtn.SetAttribute("on:click", func() {
		counter.Set(counter.Value() - 1)
	})
	renderer.Render(decBtn)

	lbl := NewComponent("span")
	lbl.SetAttribute("innerText", 0)
	counter.Effect(func() {
		renderer.SetAttribute(lbl, "innerText", counter.Value())
	})
	renderer.Render(lbl)

	br := NewComponent("br")
	renderer.Render(br)
}

type MyAmazingListProps struct {
	counter *Signal[int64]
	items   *Signal[[]string]
}

func MyAmazingList(renderer *DOMRenderer, props MyAmazingListProps) {
	counter := props.counter
	items := props.items

	input := NewComponent("input")
	input.SetAttribute("type", "text")
	counter.Effect(func() {
		renderer.SetAttribute(input, "value", counter.Value())
	})
	renderer.Render(input)

	addItemBtn := NewComponent("button")
	addItemBtn.SetAttribute("innerText", "add item")
	// todo: handle defer jsHandleAddItem.Release()
	addItemBtn.SetAttribute("on:click", func() {
		items.Set(append(items.Value(), input.el.Get("value").String()))
	})
	renderer.Render(addItemBtn)

	list := NewComponent("ul")
	items.Effect(func() {
		item := NewComponent("li")
		item.SetAttribute("innerText", items.Value()[len(items.Value())-1])
		renderer.append(list, item)
	})
	renderer.Render(list)
}
